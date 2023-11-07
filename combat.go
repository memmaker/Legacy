package main

import (
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/gocoro"
    "Legacy/renderer"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "image/color"
    "time"
)

// CombatState should handle
// Participants
// Switching to turn-based
// Turn order
// Player input & Actions
// AI
type CombatState struct {
    opponents            map[*game.Actor]bool
    engine               *GridEngine
    movesTakenThisTurn   map[*game.Actor]int
    hasUsedPrimaryAction map[*game.Actor]bool
    animationRoutine     gocoro.Coroutine
    hitAnimations        []HitAnimation
    isPlayerTurn         bool
    waitForTarget        func(targetPos geometry.Point)
    isInCombat           bool
    didAlertNearbyActors bool
    iconGenericMissile   int32
    partyAutoAttacks     bool
}
type HitAnimation struct {
    Path          []geometry.Point
    CurrentIndex  int
    TicksLeft     int
    TicksForReset int
    Icon          int32
    UseTiles      renderer.AtlasName
    WhenDone      func()
    TintColor     color.Color
}

func (h HitAnimation) Position() geometry.Point {
    return h.Path[h.CurrentIndex]
}

func (h HitAnimation) IsFinished() bool {
    return h.TicksLeft <= 0 && h.CurrentIndex == len(h.Path)-1
}

func NewCombatState(gridEngine *GridEngine) *CombatState {
    return &CombatState{
        opponents:            make(map[*game.Actor]bool),
        movesTakenThisTurn:   make(map[*game.Actor]int),
        hasUsedPrimaryAction: make(map[*game.Actor]bool),
        engine:               gridEngine,
        animationRoutine:     gocoro.NewCoroutine(),
        hitAnimations:        []HitAnimation{},
        iconGenericMissile:   26,
    }
}

type AttackActionType int

const (
    AttackActionTypeMelee AttackActionType = iota
    AttackActionTypeRanged
    AttackActionTypeMagic
)

type AttackAction struct {
    attacker *game.Actor
    target   *game.Actor
    action   AttackActionType
}

func (c *CombatState) Update() {
    animationsRunning := false
    // handle player input
    if c.animationRoutine.Running() {
        c.animationRoutine.Update()
        animationsRunning = true
    }

    for i := len(c.hitAnimations) - 1; i >= 0; i-- {
        hitAnim := &c.hitAnimations[i]
        hitAnim.TicksLeft--
        if hitAnim.TicksLeft <= 0 {
            if hitAnim.IsFinished() {
                c.hitAnimations = append(c.hitAnimations[:i], c.hitAnimations[i+1:]...)
                if hitAnim.WhenDone != nil {
                    hitAnim.WhenDone()
                }
            } else {
                hitAnim.CurrentIndex++
                hitAnim.TicksLeft = hitAnim.TicksForReset
            }
        }
        animationsRunning = true
    }

    if animationsRunning {
        return
    }

    if c.isPlayerTurn {
        if c.canPartyAct() {
            if !c.canAct(c.engine.GetAvatar()) {
                c.switchToNextPartyMember()
            }
            if c.partyAutoAttacks {
                c.automaticBattleActionFor(c.engine.GetAvatar(), c.listOfEnemies())
            } else {
                // wait for input
                c.handleInput()
            }
        } else {
            c.endPlayerTurn()
        }
    } else {
        // do ai
        if c.canEnemyAct() {
            // wait for input
            c.takeEnemyTurn()
        } else {
            c.endEnemyTurn()
        }
    }
}

func (c *CombatState) Draw(screen *ebiten.Image) {
    // draw the combat UI
    offset := c.engine.gridRenderer.GetScaledSmallGridSize()
    offsetPoint := geometry.Point{X: offset, Y: offset}
    for _, hitAnim := range c.hitAnimations {
        screenPos := c.engine.MapToScreenCoordinates(hitAnim.Position())
        c.engine.gridRenderer.DrawOnBigGridWithColor(screen, screenPos, offsetPoint, hitAnim.UseTiles, hitAnim.Icon, hitAnim.TintColor)
    }

    if c.isPlayerTurn && !c.engine.IsWindowOpen() {
        avatarPos := c.engine.GetAvatar().Pos()
        screenPos := c.engine.MapToScreenCoordinates(avatarPos)
        selectionIndicator := int32(195)
        c.engine.gridRenderer.DrawOnBigGrid(screen, screenPos, offsetPoint, renderer.AtlasEntities, selectionIndicator)
    }
}

func (c *CombatState) IsInCombat() bool {
    return c.isInCombat || len(c.hitAnimations) > 0 || c.animationRoutine.Running()
}

func (c *CombatState) animateMeleeHit(attacker *game.Actor, npc *game.Actor) {
    doesDamage := c.engine.rules.CalculateHit(attacker, npc)
    hitPos := npc.Pos()
    icon := int32(194)
    useAtlas := renderer.AtlasEntities
    if doesDamage {
        icon = int32(104)
        useAtlas = renderer.AtlasWorld
    }
    c.hitAnimations = append(c.hitAnimations, HitAnimation{
        UseTiles:  useAtlas,
        Path:      []geometry.Point{hitPos},
        TicksLeft: 35,
        Icon:      icon,
        TintColor: color.White,
        WhenDone: func() {
            c.hasUsedPrimaryAction[attacker] = true
            if doesDamage {
                c.deliverDamage(attacker, npc)
            }
        },
    })
}
func (c *CombatState) animateProjectile(attacker *game.Actor, target geometry.Point, icon int32, tint color.Color, onImpact func(dest geometry.Point, actor *game.Actor) func()) {
    currentMap := c.engine.currentMap
    source := attacker.Pos()
    path := c.projectilePath(source, target)
    finalDestination := path[len(path)-1]

    var actorHit *game.Actor
    if currentMap.IsActorAt(finalDestination) {
        actorHit = currentMap.ActorAt(finalDestination)
        c.addOpponent(actorHit)
    }
    c.hitAnimations = append(c.hitAnimations, HitAnimation{
        Path:          path,
        TicksLeft:     12,
        TicksForReset: 12,
        Icon:          icon,
        TintColor:     tint,
        UseTiles:      renderer.AtlasEntitiesGrayscale,
        WhenDone:      onImpact(finalDestination, actorHit),
    })
}

func (c *CombatState) projectilePath(source geometry.Point, destination geometry.Point) []geometry.Point {
    currentMap := c.engine.currentMap
    los := geometry.BresenhamLoS(source, destination, func(x, y int) bool {
        p := geometry.Point{X: x, Y: y}
        return (currentMap.Contains(p) && currentMap.IsPassableForProjectile(p)) || p == source
    })
    losWithoutSource := los[1:]
    return losWithoutSource
}

func (c *CombatState) getMovesLeft(actor *game.Actor) int {
    return actor.GetMovementAllowance() - c.movesTakenThisTurn[actor]
}

func (c *CombatState) canAct(actor *game.Actor) bool {
    if !actor.IsAlive() {
        return false
    }
    actorMovementAllowance := actor.GetMovementAllowance()
    if c.hasUsedPrimaryAction[actor] {
        return false
    }
    return c.movesTakenThisTurn[actor] < actorMovementAllowance
}
func (c *CombatState) canPartyAct() bool {
    for _, partyMember := range c.engine.GetPartyMembers() {
        if c.canAct(partyMember) {
            return true
        }
    }
    return false
}

func (c *CombatState) handleInput() {
    if c.engine.inputElement != nil {
        if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
            c.engine.inputElement.ActionDown()
        } else if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
            c.engine.inputElement.ActionUp()
        } else if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
            c.engine.inputElement.ActionConfirm()
        } else if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
            c.engine.inputElement = nil
        }
        return
    }
    if c.engine.modalElement != nil {
        if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
            c.engine.modalElement.ActionConfirm()
        } else if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
            c.engine.modalElement = nil
        }
        return
    }
    avatar := c.engine.GetAvatar()
    curPos := avatar.Pos()
    var direction, dest geometry.Point
    if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
        direction = geometry.Point{X: 1, Y: 0}
    }

    if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
        direction = geometry.Point{X: -1, Y: 0}
    }

    if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
        direction = geometry.Point{X: 0, Y: -1}
    }

    if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
        direction = geometry.Point{X: 0, Y: 1}
    }

    if direction != (geometry.Point{}) {
        if c.waitForTarget != nil {
            c.waitForTarget(curPos.Add(direction.Mul(10)))
            return
        }

        dest = curPos.Add(direction)
        if isThere, enemy := c.isEnemyAt(dest); isThere {
            c.animateMeleeHit(avatar, enemy)
        } else {
            c.engine.MoveAvatarInDirection(direction)
            if curPos != avatar.Pos() {
                c.movesTakenThisTurn[avatar]++
            }
        }
    }

    if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
        c.openCombatMenu(avatar)
    }

    if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
        x, y := ebiten.CursorPosition()
        c.OnMouseClicked(x, y)
    }
}

func (c *CombatState) openCombatMenu(partyMember *game.Actor) {
    combatOptions := []renderer.MenuItem{
        {
            Text: "Ranged",
            Action: func() {
                c.selectRangedTarget(partyMember)
                c.engine.inputElement = nil
            },
        },
        {
            Text: "Magic",
            Action: func() {
                c.engine.inputElement = nil
                c.engine.openCombatSpellMenu(partyMember)
            },
        },
        {
            Text: "Auto-Attack",
            Action: func() {
                c.partyAutoAttacks = true
                c.engine.inputElement = nil
            },
        },
        {
            Text: "End turn",
            Action: func() {
                c.hasUsedPrimaryAction[partyMember] = true
                c.engine.inputElement = nil
            },
        },
    }
    c.engine.openMenu(combatOptions)
}
func (c *CombatState) endPlayerTurn() {
    c.isPlayerTurn = false
    clear(c.movesTakenThisTurn)
    clear(c.hasUsedPrimaryAction)
    c.checkForEndOfCombat()
}

func (c *CombatState) endEnemyTurn() {
    c.isPlayerTurn = true
    clear(c.movesTakenThisTurn)
    clear(c.hasUsedPrimaryAction)
    c.checkForEndOfCombat()
    if c.isInCombat {
        c.switchToNextPartyMember()
    }
}
func (c *CombatState) combatInitByPlayer() {
    clear(c.movesTakenThisTurn)
    clear(c.hasUsedPrimaryAction)
    clear(c.opponents)
    c.partyAutoAttacks = false
    c.isInCombat = true
    c.isPlayerTurn = true
    c.didAlertNearbyActors = false
}

func (c *CombatState) canEnemyAct() bool {
    for opponent, _ := range c.opponents {
        if c.canAct(opponent) {
            return true
        }
    }
    return false
}

func (c *CombatState) actorDied(actor *game.Actor) {
    if !c.engine.IsPlayerControlled(actor) {
        c.removeOpponent(actor)
    }
    c.engine.actorDied(actor)
}

func (c *CombatState) deliverDamage(attacker, victim *game.Actor) {
    c.engine.DeliverCombatDamage(attacker, victim)
    if !victim.IsAlive() {
        c.actorDied(victim)
    }
}

func (c *CombatState) takeEnemyTurn() {
    // does nothing for now
    for opponent, _ := range c.opponents {
        if c.canAct(opponent) {
            //c.engine.Print(fmt.Sprintf("'%s' turn", opponent.Name()))
            c.automaticBattleActionFor(opponent, c.engine.GetPartyMembers())
            return
        }
    }
}

func (c *CombatState) switchToNextPartyMember() {
    for _, partyMember := range c.engine.GetPartyMembers() {
        if c.canAct(partyMember) {
            c.engine.SwitchAvatarTo(partyMember)
            //c.engine.Print(fmt.Sprintf("It's '%s' turn", partyMember.Name()))
            return
        }
    }
}

func (c *CombatState) isEnemyAt(dest geometry.Point) (bool, *game.Actor) {
    for opponent, _ := range c.opponents {
        if opponent.Pos() == dest {
            return true, opponent
        }
    }
    return false, nil
}

type BattleAction struct {
    Movement   []geometry.Point
    ActionType AttackActionType
    Target     *game.Actor
}

func (a BattleAction) IsEndTurn() bool {
    return len(a.Movement) == 0 && a.Target == nil
}

func (c *CombatState) automaticBattleActionFor(ourActor *game.Actor, listOfEnemies []*game.Actor) {
    autoAction := c.calculateBestAction(ourActor, listOfEnemies)
    if autoAction.IsEndTurn() {
        c.removeOpponent(ourActor)
        c.hasUsedPrimaryAction[ourActor] = true
    } else {
        c.animateBattleAction(ourActor, autoAction)
    }
}
func (c *CombatState) calculateBestAction(ourActor *game.Actor, listOfEnemies []*game.Actor) BattleAction {
    // let's start simple..
    // find the nearest enemy, find a free cell position next to him
    // move towards that position
    // if we can reach it this turn: attack
    attackType := AttackActionTypeMelee
    movement, attackTarget := c.closeIntoMeleeRange(ourActor, listOfEnemies)
    return BattleAction{
        Movement:   movement,
        ActionType: attackType,
        Target:     attackTarget,
    }
}

func (c *CombatState) closeIntoMeleeRange(ourActor *game.Actor, listOfEnemies []*game.Actor) ([]geometry.Point, *game.Actor) {
    // find the nearest enemy
    var nearestEnemy *game.Actor
    var nearestPath []geometry.Point
    currentMap := c.engine.currentMap
    for _, enemy := range listOfEnemies {
        if enemy.IsAlive() {
            if enemy.IsRightNextTo(ourActor) {
                return []geometry.Point{}, enemy
            }
            freeNeighbors := currentMap.NeighborsCardinal(enemy.Pos(), func(p geometry.Point) bool {
                return currentMap.Contains(p) && currentMap.IsWalkableFor(p, ourActor)
            })
            if len(freeNeighbors) == 0 {
                continue
            }
            for _, freeNeighbor := range freeNeighbors {
                currentPath := currentMap.GetJPSPath(ourActor.Pos(), freeNeighbor, func(point geometry.Point) bool {
                    return currentMap.Contains(point) && currentMap.IsWalkableFor(point, ourActor)
                })
                if nearestEnemy == nil || (len(currentPath) > 0 && len(currentPath) < len(nearestPath)) {
                    nearestEnemy = enemy
                    nearestPath = currentPath
                }
            }
        }
    }

    if nearestEnemy != nil {
        distance := geometry.DistanceManhattan(ourActor.Pos(), nearestEnemy.Pos())
        if distance > 1 && len(nearestPath) == 0 {
            return []geometry.Point{}, nil // we're stuck
        }
    }

    movesLeft := c.getMovesLeft(ourActor)
    if len(nearestPath) > movesLeft {
        return nearestPath[:movesLeft], nil
    }
    // cannot reach any enemies
    return nearestPath, nearestEnemy
}

func (c *CombatState) animateBattleAction(actor *game.Actor, move BattleAction) {
    _ = c.animationRoutine.Run(func(exe *gocoro.Execution) {
        for _, dest := range move.Movement {
            c.engine.moveActor(actor, dest)
            if c.engine.IsPlayerControlled(actor) {
                c.engine.onViewedActorMoved(actor.Pos()) // UNCOMMENT FOR CAM FOLLOW BEHAVIOR
            }
            c.movesTakenThisTurn[actor]++
            _ = exe.YieldTime(200 * time.Millisecond)
        }

        if move.ActionType == AttackActionTypeMelee && move.Target != nil {
            c.animateMeleeHit(actor, move.Target)
        }
    })
}

func (c *CombatState) findAlliesOfOpponent(attackedNPC *game.Actor) {
    if c.didAlertNearbyActors {
        return
    }
    if c.canAct(attackedNPC) {
        c.opponents[attackedNPC] = true
    }
    radius := 25
    // TODO: better filter for guards, allies, etc.
    nearbyNPCs := c.engine.currentMap.FindAllNearbyActors(attackedNPC.Pos(), radius, func(actor *game.Actor) bool {
        return !c.engine.IsPlayerControlled(actor) && actor.IsAlive() && !actor.IsHidden() && actor != attackedNPC
    })

    for _, nearbyNPC := range nearbyNPCs {
        c.opponents[nearbyNPC] = true
    }
    c.didAlertNearbyActors = true
}

func (c *CombatState) removeOpponent(actor *game.Actor) {
    if c.engine.IsPlayerControlled(actor) {
        return
    }
    delete(c.opponents, actor)
    c.checkForEndOfCombat()
}

func (c *CombatState) checkForEndOfCombat() {
    c.removeDeadOpponents()
    if len(c.opponents) == 0 && len(c.hitAnimations) == 0 && !c.animationRoutine.Running() {
        c.isInCombat = false
        c.engine.ForceJoinParty()
    }
}

func (c *CombatState) OnMouseClicked(screenX, screenY int) {
    if c.waitForTarget != nil {
        mapPos := c.engine.ScreenToMap(screenX, screenY)
        c.waitForTarget(mapPos)
    }
    //c.engine.mapWindow.GetScreenGridPositionFromMapGridPosition()
}

func (c *CombatState) PlayerStartsMeleeAttack(attacker *game.Actor, attackedNPC *game.Actor) {
    c.combatInitByPlayer()
    c.findAlliesOfOpponent(attackedNPC)
    c.animateMeleeHit(attacker, attackedNPC)
}

func (c *CombatState) PlayerStartsRangedAttack() {
    c.combatInitByPlayer()
    c.selectRangedTarget(c.engine.GetAvatar())
}

func (c *CombatState) OrchestratedRangedAttack() {
    c.combatInitByPlayer()
    c.selectOrchestratedRangedTarget()
}

func (c *CombatState) PlayerStartsOffensiveSpell(spellUsed *game.Spell) {
    c.combatInitByPlayer()
    c.SelectSpellTarget(c.engine.GetAvatar(), spellUsed)
}
func (c *CombatState) SelectSpellTarget(attacker *game.Actor, spell *game.Spell) {
    spellIcon := int32(28)
    spellColor := spell.Color()
    c.waitForTarget = func(targetPos geometry.Point) {
        c.hasUsedPrimaryAction[attacker] = true
        c.animateProjectile(attacker, targetPos, spellIcon, spellColor, func(pos geometry.Point, actorHit *game.Actor) func() {
            return func() {
                c.onSpellImpact(attacker, spell, pos, actorHit)
            }
        })
        c.waitForTarget = nil
    }
}
func (c *CombatState) selectRangedTarget(attacker *game.Actor) {
    c.waitForTarget = func(targetPos geometry.Point) {
        c.animateProjectile(attacker, targetPos, c.iconGenericMissile, color.White, func(pos geometry.Point, actorHit *game.Actor) func() {
            return func() {
                c.onRangedHit(attacker, actorHit)
            }
        })
        c.waitForTarget = nil
    }
}
func (c *CombatState) selectOrchestratedRangedTarget() {
    c.waitForTarget = func(targetPos geometry.Point) {
        for _, partyMember := range c.engine.GetPartyMembers() {
            c.animateProjectile(partyMember, targetPos, c.iconGenericMissile, color.White, func(pos geometry.Point, actorHit *game.Actor) func() {
                return func() {
                    c.onRangedHit(partyMember, actorHit)
                }
            })
        }
        c.waitForTarget = nil
    }
}
func (c *CombatState) onSpellImpact(attacker *game.Actor, spell *game.Spell, finalDestination geometry.Point, actorHit *game.Actor) {
    c.hasUsedPrimaryAction[attacker] = true
    if actorHit != nil {
        if actorHit.IsHidden() {
            actorHit.SetHidden(false)
        }
        if !c.didAlertNearbyActors {
            c.findAlliesOfOpponent(actorHit)
        }
    }

    if spell.IsTargeted() {
        spell.CastOnTarget(c.engine, attacker, finalDestination)
    }
}
func (c *CombatState) onRangedHit(attacker, actorHit *game.Actor) {
    c.hasUsedPrimaryAction[attacker] = true
    if actorHit != nil {
        if actorHit.IsHidden() {
            actorHit.SetHidden(false)
        }
        c.deliverDamage(attacker, actorHit)
        if !c.didAlertNearbyActors {
            c.findAlliesOfOpponent(actorHit)
        }
    }
}

func (c *CombatState) addOpponent(opponent *game.Actor) {
    if c.engine.IsPlayerControlled(opponent) {
        return
    }
    if !c.didAlertNearbyActors {
        c.findAlliesOfOpponent(opponent)
    }
    if opponent.IsAlive() {
        c.opponents[opponent] = true
    }
}

func (c *CombatState) addHitAnimation(pos geometry.Point, atlas renderer.AtlasName, icon int32, tintColor color.Color, done func()) {
    c.hitAnimations = append(c.hitAnimations, HitAnimation{
        Path:          []geometry.Point{pos},
        TicksLeft:     20,
        TicksForReset: 20,
        Icon:          icon,
        WhenDone:      done,
        UseTiles:      atlas,
        TintColor:     tintColor,
    })
}

func (c *CombatState) listOfEnemies() []*game.Actor {
    var enemies []*game.Actor
    for enemy, _ := range c.opponents {
        enemies = append(enemies, enemy)
    }
    return enemies
}

func (c *CombatState) removeDeadOpponents() {
    for opponent, _ := range c.opponents {
        if !opponent.IsAlive() {
            delete(c.opponents, opponent)
        }
    }
}
