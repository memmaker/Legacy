package main

import (
    "Legacy/ega"
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/gocoro"
    "Legacy/renderer"
    "Legacy/ui"
    "Legacy/util"
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
    "math"
    "slices"
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
    isPlayerTurn         bool
    waitForTarget        func(targetPos geometry.Point)
    validTargets         map[geometry.Point]bool
    isInCombat           bool
    didAlertNearbyActors bool
    iconGenericMissile   int32
    partyAutoAttacks     bool
    animator             *renderer.Animator
    stuckCounter         map[*game.Actor]int
}

func (c *CombatState) OnMouseWheel(x int, y int, dy float64) bool {
    return false
}

func (c *CombatState) OnCommand(command ui.CommandType) bool {
    if !c.isPlayerTurn {
        return false
    }
    if !c.canAct(c.engine.GetAvatar()) {
        return false
    }
    switch command {
    case ui.PlayerCommandConfirm:
        c.openCombatMenu(c.engine.GetAvatar())
    case ui.PlayerCommandUp:
        c.onDirectionalInput(geometry.Point{X: 0, Y: -1})
    case ui.PlayerCommandDown:
        c.onDirectionalInput(geometry.Point{X: 0, Y: 1})
    case ui.PlayerCommandLeft:
        c.onDirectionalInput(geometry.Point{X: -1, Y: 0})
    case ui.PlayerCommandRight:
        c.onDirectionalInput(geometry.Point{X: 1, Y: 0})
    case ui.PlayerCommandCancel:
        if c.waitForTarget != nil {
            c.waitForTarget = nil
            clear(c.validTargets)
            c.engine.ClearOverlay()
            c.checkForEndOfCombat()
            return true
        }
    }
    return true
}

func NewCombatState(gridEngine *GridEngine) *CombatState {
    return &CombatState{
        opponents:            make(map[*game.Actor]bool),
        movesTakenThisTurn:   make(map[*game.Actor]int),
        hasUsedPrimaryAction: make(map[*game.Actor]bool),
        stuckCounter:         make(map[*game.Actor]int),
        engine:               gridEngine,
        iconGenericMissile:   26,
        animator: renderer.NewAnimator(
            func(mapPos geometry.Point) (bool, geometry.Point) {
                isOnScreen := gridEngine.IsMapPosOnScreen(mapPos)
                if isOnScreen {
                    return true, gridEngine.MapToScreenCoordinates(mapPos)
                } else {
                    return false, geometry.Point{}
                }
            }),
    }
}

type AttackActionType int

const (
    AttackActionTypeMelee AttackActionType = iota
    AttackActionTypeRanged
    AttackActionTypeSpell
    AttackActionTypeActiveSkill
    AttackActionTypeMove
)

type AttackAction struct {
    attacker *game.Actor
    target   *game.Actor
    action   AttackActionType
}

func (c *CombatState) Update() {

    c.animator.Update()
    if c.animator.IsRunning() {
        return
    }

    if c.isPlayerTurn {
        if c.canPartyAct() {
            if !c.canAct(c.engine.GetAvatar()) {
                c.switchToNextPartyMember()
            }
            if c.partyAutoAttacks {
                c.automaticBattleActionFor(c.engine.GetAvatar())
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

    c.animator.Draw(c.engine.gridRenderer, screen)

    if c.isPlayerTurn && !c.engine.IsWindowOpen() {
        avatarPos := c.engine.GetAvatar().Pos()
        screenPos := c.engine.MapToScreenCoordinates(avatarPos)
        selectionIndicator := int32(195)
        c.engine.gridRenderer.DrawOnBigGrid(screen, screenPos, offsetPoint, renderer.AtlasEntities, selectionIndicator)
    }
}

func (c *CombatState) IsInCombat() bool {
    return c.isInCombat || c.animator.IsRunning()
}
func (c *CombatState) meleeHitOnLocation(attacker *game.Actor, pos geometry.Point) {
    if !c.engine.currentMap.Contains(pos) || !c.engine.currentMap.IsActorAt(pos) {
        return
    }
    victim := c.engine.currentMap.ActorAt(pos)
    c.meleeHitOnActor(attacker, victim)
}
func (c *CombatState) meleeHitOnActor(attacker *game.Actor, npc *game.Actor) {
    doesHit := c.engine.rules.DoesMeleeAttackHit(attacker, npc)
    hitPos := npc.Pos()
    icon := int32(194)
    useAtlas := renderer.AtlasEntities
    if doesHit {
        icon = int32(104)
        useAtlas = renderer.AtlasWorld
    }
    animation := &renderer.TileAnimation{
        UseTiles:  useAtlas,
        Positions: []geometry.Point{hitPos},
        Frames:    []int32{icon},
        TintColor: color.White,
        WhenDone: func() {
            c.hasUsedPrimaryAction[attacker] = true
            if doesHit {
                c.deliverMeleeDamage(attacker, npc)
            }
        },
    }
    animation.EndAfterTime(0.33)
    c.animator.AddHitAnimation(animation)
}
func (c *CombatState) animateProjectile(attacker *game.Actor, target geometry.Point, icon int32, tint color.Color, onImpact func(dest geometry.Point, actor *game.Actor) func()) {

    currentMap := c.engine.currentMap
    source := attacker.Pos()

    if source == target {
        //TODO
        return
    }

    path := c.getLineOfSight(source, target, c.isActorBlockingSightForAttacker(attacker))
    if len(path) == 0 {
        //TODO
    }
    finalDestination := path[len(path)-1]

    var actorHit *game.Actor
    if currentMap.IsActorAt(finalDestination) {
        actorHit = currentMap.ActorAt(finalDestination)
        c.addOpponent(actorHit)
    }
    animation := &renderer.TileAnimation{
        Positions:             path,
        Frames:                []int32{icon},
        MoveIntervalInSeconds: 0.13,
        TintColor:             tint,
        UseTiles:              renderer.AtlasEntitiesGrayscale,
        WhenDone:              onImpact(finalDestination, actorHit),
    }
    animation.EndAfterPathComplete()
    c.animator.AddHitAnimation(animation)
}

func (c *CombatState) getLineOfSight(source geometry.Point, destination geometry.Point, isActorBlocking func(actor *game.Actor) bool) []geometry.Point {
    currentMap := c.engine.currentMap
    los := geometry.BresenhamLoS(source, destination, func(x, y int) bool {
        p := geometry.Point{X: x, Y: y}
        if !currentMap.Contains(p) {
            return false
        }
        if currentMap.IsActorAt(p) {
            actorAt := currentMap.ActorAt(p)
            if isActorBlocking(actorAt) {
                return false
            }
        }
        isWalkable := currentMap.IsWalkable(p)
        return isWalkable || p == source
    })
    losWithoutSource := los[1:]
    return losWithoutSource
}

func (c *CombatState) getMovesLeft(actor *game.Actor) int {
    return actor.GetMovementAllowance() - c.movesTakenThisTurn[actor]
}

func (c *CombatState) canAct(actor *game.Actor) bool {
    if !actor.CanAct() {
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

func (c *CombatState) OnMouseClicked(xGrid int, yGrid int) bool {
    x, y := ebiten.CursorPosition()
    return c.OnScreenMouseClicked(x, y)
}

func (c *CombatState) onDirectionalInput(direction geometry.Point) bool {
    avatar := c.engine.GetAvatar()
    curPos := avatar.Pos()
    if c.waitForTarget != nil {
        counter := 0
        potentialTargetPos := curPos.Add(direction)
        _, isValidPos := c.validTargets[potentialTargetPos]
        for counter < 20 && !isValidPos {
            counter++
            potentialTargetPos = potentialTargetPos.Add(direction)
            _, isValidPos = c.validTargets[potentialTargetPos]
        }
        if isValidPos {
            c.waitForTarget(potentialTargetPos)
            return true
        }
        return true
    }

    dest := curPos.Add(direction)
    if isThere, enemy := c.isEnemyAt(dest); isThere {
        c.meleeHitOnActor(avatar, enemy)
        return true
    } else {
        c.engine.PlayerMovement(direction)
        if curPos != avatar.Pos() {
            c.movesTakenThisTurn[avatar]++
            return true
        }
    }
    return false
}

func (c *CombatState) openCombatMenu(partyMember *game.Actor) {
    combatOptions := []util.MenuItem{
        {
            Text: "Ranged",
            Action: func() {
                c.selectRangedTarget(partyMember)
                c.engine.CloseAllModals()
            },
        },
        {
            Text: "Magic",
            Action: func() {
                c.engine.CloseAllModals()
                c.engine.openActiveSkillsMenu(partyMember, partyMember.GetEquippedSpells())
            },
        },
        {
            Text: "Skills",
            Action: func() {
                c.engine.CloseAllModals()
                c.engine.openActiveSkillsMenu(partyMember, partyMember.GetActiveSkills())
            },
        },
        {
            Text: "Auto-Attack",
            Action: func() {
                c.partyAutoAttacks = true
                c.engine.CloseAllModals()
            },
        },
        {
            Text: "End turn",
            Action: func() {
                c.hasUsedPrimaryAction[partyMember] = true
                c.engine.CloseAllModals()
            },
        },
    }
    c.engine.OpenMenu(combatOptions)
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
func (c *CombatState) ensureCombatInit(isPlayerTurn bool) {
    if c.isInCombat {
        return
    }
    clear(c.movesTakenThisTurn)
    clear(c.hasUsedPrimaryAction)
    clear(c.opponents)
    c.partyAutoAttacks = false
    c.isInCombat = true
    c.didAlertNearbyActors = false
    c.isPlayerTurn = isPlayerTurn
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

func (c *CombatState) deliverMeleeDamage(attacker, victim *game.Actor) {
    c.engine.DeliverMeleeDamage(attacker, victim)
    if !victim.IsAlive() {
        c.actorDied(victim)
    }
}
func (c *CombatState) deliverRangedDamage(attacker, victim *game.Actor) {
    c.engine.DeliverRangedDamage(attacker, victim)
    if !victim.IsAlive() {
        c.actorDied(victim)
    }
}

func (c *CombatState) takeEnemyTurn() {
    // does nothing for now
    for opponent, _ := range c.opponents {
        if c.canAct(opponent) {
            //c.engine.Print(fmt.Sprintf("'%s' turn", opponent.Name()))
            c.automaticBattleActionFor(opponent)
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
    MovementDestination  geometry.Point
    ActionType           AttackActionType
    ActiveAction         game.Action // skill / spell
    ActionTargetLocation *geometry.Point
}

func (a BattleAction) IsValid(actorPos geometry.Point) bool {
    return a.MovementDestination != actorPos || a.ActiveAction != nil || a.ActionTargetLocation != nil
}

func (c *CombatState) automaticBattleActionFor(ourActor *game.Actor) {
    autoAction, reachablePositions := c.calculateBestAction(ourActor)
    c.executeBattleAction(ourActor, autoAction, reachablePositions)
}
func (c *CombatState) calculateBestAction(ourActor *game.Actor) (BattleAction, map[geometry.Point]int) {
    // let's start simple..
    // find the nearest enemy, find a free cell position next to him
    // move towards that position
    // if we can reach it this turn: attack

    gridmap := c.engine.GetGridMap()
    reachablePositions := gridmap.GetDijkstraMap(ourActor.Pos(), c.getMovesLeft(ourActor), func(p geometry.Point) bool {
        return gridmap.IsCurrentlyPassable(p)
    })
    reachablePositions[ourActor.Pos()] = 0

    maxUtility := math.MinInt
    bestAction := BattleAction{
        MovementDestination: ourActor.Pos(),
    }
    for pos, _ := range reachablePositions {
        utility, actionAt := c.getUtilityForPosition(ourActor, pos)
        if utility > maxUtility {
            maxUtility = utility
            bestAction = actionAt
        }
    }

    if maxUtility == 0 {
        attackType := AttackActionTypeMelee
        moveTo, attackTarget := c.closeIntoMeleeRange(ourActor, reachablePositions)
        bestAction = BattleAction{
            MovementDestination:  moveTo,
            ActionType:           attackType,
            ActionTargetLocation: attackTarget,
        }
    }

    return bestAction, reachablePositions
}

func (c *CombatState) closeIntoMeleeRange(ourActor *game.Actor, reachablePositions map[geometry.Point]int) (geometry.Point, *geometry.Point) {
    // find the nearest enemy
    var nearestEnemy *game.Actor
    nearestPos := ourActor.Pos()
    var attackPos *geometry.Point
    currentMap := c.engine.currentMap
    for _, enemy := range c.getActiveOpponents(ourActor) {
        if enemy.IsAlive() {
            if enemy.IsRightNextTo(ourActor) {
                enemyPos := enemy.Pos()
                return ourActor.Pos(), &enemyPos
            }
            freeNeighbors := currentMap.NeighborsCardinal(enemy.Pos(), func(p geometry.Point) bool {
                return currentMap.Contains(p) && currentMap.IsWalkableFor(p, ourActor)
            })
            if len(freeNeighbors) == 0 {
                continue
            }
            for _, freeNeighbor := range freeNeighbors {
                if _, ok := reachablePositions[freeNeighbor]; ok {
                    distance := geometry.DistanceManhattan(ourActor.Pos(), freeNeighbor)
                    if nearestEnemy == nil || distance < geometry.DistanceManhattan(ourActor.Pos(), nearestPos) {
                        nearestEnemy = enemy
                        nearestPos = freeNeighbor
                    }
                }
            }
        }
    }
    if nearestEnemy != nil {
        enemyPos := nearestEnemy.Pos()
        attackPos = &enemyPos
    }
    return nearestPos, attackPos
}

func (c *CombatState) executeBattleAction(actor *game.Actor, move BattleAction, reachablePositions map[geometry.Point]int) {
    if !c.canAct(actor) {
        return
    }
    c.hasUsedPrimaryAction[actor] = true

    if !move.IsValid(actor.Pos()) {
        c.stuckCounter[actor]++
        if c.stuckCounter[actor] > 3 {
            c.engine.Print(fmt.Sprintf("%s cannot reach combat zone", actor.Name()))
            c.removeOpponent(actor)
            return
        }
    } else {
        c.stuckCounter[actor] = 0
    }

    combatPathTo := c.getCombatPathTo(actor, reachablePositions, move.MovementDestination)

    c.animator.RunAnimationScript(func(exe *gocoro.Execution) {
        for _, dest := range combatPathTo {
            c.engine.moveActorInCombat(actor, dest)
            c.movesTakenThisTurn[actor]++
            if !actor.IsAlive() {
                c.actorDied(actor)
                return
            }
            if actor.IsSleeping() {
                if !c.engine.IsPlayerControlled(actor) {
                    c.removeOpponent(actor)
                }
                return
            }
            _ = exe.YieldTime(200 * time.Millisecond)
        }

        hasTargetAndCanAct := move.ActionTargetLocation != nil && (actor.IsAlive() && !actor.IsSleeping())

        if !hasTargetAndCanAct {
            return
        }

        if move.ActionType == AttackActionTypeMelee {
            c.meleeHitOnLocation(actor, *move.ActionTargetLocation)
        } else if move.ActionType == AttackActionTypeActiveSkill || move.ActionType == AttackActionTypeSpell {
            c.UseTargetedAction(actor, move.ActiveAction, *move.ActionTargetLocation)
        } else if move.ActionType == AttackActionTypeRanged {
            c.RangedAttack(actor, *move.ActionTargetLocation)
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
    currentMap := c.engine.currentMap
    nearbyNPCs := currentMap.FindAllNearbyActors(attackedNPC.Pos(), radius, func(actor *game.Actor) bool {
        return !c.engine.IsPlayerControlled(actor) &&
            actor.IsAlive() &&
            !actor.IsHidden() &&
            actor != attackedNPC &&
            c.areAllies(attackedNPC, actor)
    })
    avatar := c.engine.GetAvatar()
    for _, nearbyNPC := range nearbyNPCs {
        randomPosNearby := currentMap.GetRandomFreeNeighbor(avatar.Pos())
        path := currentMap.GetJPSPath(nearbyNPC.Pos(), randomPosNearby, currentMap.IsCurrentlyPassable)
        if len(path) > (nearbyNPC.GetMovementAllowance() * 4) {
            continue
        }
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
    c.removeDeadAndSleepingOpponents()
    if len(c.opponents) == 0 && !c.animator.IsRunning() {
        c.endCombat()
    }
}

func (c *CombatState) endCombat() {
    c.isInCombat = false
    c.engine.ForceJoinParty()
    clear(c.movesTakenThisTurn)
    clear(c.hasUsedPrimaryAction)
}

func (c *CombatState) OnScreenMouseClicked(screenX, screenY int) bool {
    if c.waitForTarget != nil {
        mapPos := c.engine.ScreenToMap(screenX, screenY)
        c.waitForTarget(mapPos)
        return true
    }
    return false
}
func (c *CombatState) OnMouseMoved(screenX, screenY int) (bool, ui.Tooltip) {
    return false, ui.NoTooltip{}
}

func (c *CombatState) OnCombatAction(attacker *game.Actor, target *game.Actor) {
    isAttackerPlayerControlled := c.engine.IsPlayerControlled(attacker)
    c.ensureCombatInit(isAttackerPlayerControlled)

    npc := target
    if !isAttackerPlayerControlled {
        npc = attacker
    }
    c.findAlliesOfOpponent(npc)
}
func (c *CombatState) MeleeAttack(attacker *game.Actor, target *game.Actor) {
    c.meleeHitOnActor(attacker, target)
    c.OnCombatAction(attacker, target)
    c.checkForEndOfCombat()
}

func (c *CombatState) RangedAttack(attacker *game.Actor, targetPos geometry.Point) {
    if !attacker.HasRangedWeaponEquipped() {
        return
    }
    c.hasUsedPrimaryAction[attacker] = true
    c.animateProjectile(attacker, targetPos, c.iconGenericMissile, color.White, func(pos geometry.Point, actorHit *game.Actor) func() {
        return func() {
            if actorHit != nil {
                c.onRangedImpact(attacker, actorHit)
                c.OnCombatAction(attacker, actorHit)
            }
            c.checkForEndOfCombat()
        }
    })
}

func (c *CombatState) UseTargetedAction(attacker *game.Actor, action game.Action, targetPos geometry.Point) {
    c.hasUsedPrimaryAction[attacker] = true
    spellIcon := int32(28)
    spellColor := action.GetColor()
    c.engine.Print(fmt.Sprintf("%s uses %s", attacker.Name(), action.Name()))
    c.animateProjectile(attacker, targetPos, spellIcon, spellColor, func(pos geometry.Point, actorHit *game.Actor) func() {
        return func() {
            c.onActionImpact(attacker, action, pos, actorHit)
            if actorHit != nil {
                c.OnCombatAction(attacker, actorHit)
            }
            c.checkForEndOfCombat()
        }
    })
}
func (c *CombatState) onActionImpact(attacker *game.Actor, spell game.Action, finalDestination geometry.Point, actorHit *game.Actor) {
    if actorHit != nil {
        if actorHit.IsHidden() {
            actorHit.SetHidden(false)
        }
    }

    if spell.IsTargeted() {
        spell.ExecuteOnTarget(c.engine, attacker, finalDestination)
    }
}
func (c *CombatState) EnemyStartsCombat(attacker *game.Actor, attackedPartyMember *game.Actor) {
    c.ensureCombatInit(false)
    c.findAlliesOfOpponent(attacker)
    //c.meleeHitOnActor(attacker, attackedNPC)
}

func (c *CombatState) PlayerControlledRangedAttack() {
    avatar := c.engine.GetAvatar()

    if !c.canAct(avatar) {
        c.engine.Print("You cannot act anymore this turn")
        return
    }

    if !avatar.HasRangedWeaponEquipped() {
        c.engine.Print("No ranged weapon equipped")
        return
    }

    c.ensureCombatInit(true)
    c.selectRangedTarget(avatar)
}

func (c *CombatState) OrchestratedRangedAttack() {
    if c.engine.GetParty().HasAnyRangedWeaponsEquipped() {
        c.ensureCombatInit(true)
        c.selectOrchestratedRangedTarget()
    } else {
        c.engine.Print("No ranged weapons equipped")
    }
}

func (c *CombatState) PlayerUsesActiveSkill(user *game.Actor, skillUsed game.Action) {
    c.ensureCombatInit(true)
    c.SelectActionTarget(user, skillUsed)
}
func (c *CombatState) SelectActionTarget(attacker *game.Actor, action game.Action) {
    c.engine.CloseAllModals()
    c.waitForTarget = func(targetPos geometry.Point) {
        c.UseTargetedAction(attacker, action, targetPos)
        c.waitForTarget = nil
        c.validTargets = nil
        c.engine.ClearOverlay()
    }
    c.setValidTargets(action.GetValidTargets(c.engine, attacker, attacker.Pos()))
}
func (c *CombatState) setValidTargets(validTargets map[geometry.Point]bool) {
    c.validTargets = validTargets
    c.engine.SetOverlay(toColorMap(c.validTargets, ega.BrightGreen))
}
func toColorMap(targets map[geometry.Point]bool, drawColor color.Color) map[geometry.Point]color.Color {
    colorMap := make(map[geometry.Point]color.Color)
    for pos, _ := range targets {
        colorMap[pos] = drawColor
    }
    return colorMap
}

func (c *CombatState) selectRangedTarget(attacker *game.Actor) {
    c.engine.CloseAllModals()
    c.waitForTarget = func(targetPos geometry.Point) {
        c.RangedAttack(attacker, targetPos)
        c.waitForTarget = nil
        c.validTargets = nil
        c.engine.ClearOverlay()
    }
    c.setValidTargets(c.getPositionsOfVisibleOpponentsOfActor(attacker))
}

func (c *CombatState) selectOrchestratedRangedTarget() {
    c.waitForTarget = func(targetPos geometry.Point) {
        for _, partyMember := range c.engine.GetPartyMembers() {
            c.RangedAttack(partyMember, targetPos)
        }
        c.waitForTarget = nil
        c.validTargets = nil
        c.engine.ClearOverlay()
    }
    c.setValidTargets(c.getVisibleOpponentsOfParty())
}

func (c *CombatState) onRangedImpact(attacker, actorHit *game.Actor) {
    if actorHit != nil {
        if actorHit.IsHidden() {
            actorHit.SetHidden(false)
        }
        doesHit := c.engine.rules.DoesRangedAttackHit(attacker, actorHit)
        if doesHit {
            c.deliverRangedDamage(attacker, actorHit)
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

func (c *CombatState) listOfEnemies() []*game.Actor {
    var enemies []*game.Actor
    for enemy, _ := range c.opponents {
        enemies = append(enemies, enemy)
    }
    return enemies
}

func (c *CombatState) removeDeadAndSleepingOpponents() {
    for opponent, _ := range c.opponents {
        if !opponent.IsAlive() || opponent.IsSleeping() || c.engine.IsPlayerControlled(opponent) {
            delete(c.opponents, opponent)
        }
    }
}

func (c *CombatState) IsAITurn() bool {
    return !c.isPlayerTurn
}

func (c *CombatState) isActorBlockingSightForAttacker(attacker *game.Actor) func(actor *game.Actor) bool {
    return func(actor *game.Actor) bool {
        return !c.areAllies(attacker, actor)
    }
}

func (c *CombatState) getUtilityForPosition(ourActor *game.Actor, potentialNewLocation geometry.Point) (int, BattleAction) {
    // what's our disposition? Stay near for melee, or stay far for ranged?
    // if we can do ranged attacks, we will. If we can't, we'll try to close in.
    // for ranged: positions were we can attack but cannot be attacked are favorable
    gridMap := c.engine.GetGridMap()

    utility := 0
    bestAction := BattleAction{
        MovementDestination: potentialNewLocation,
        ActionType:          AttackActionTypeMelee,
    }
    allyPositions := make(map[geometry.Point]*game.Actor)
    enemyPositions := make(map[geometry.Point]*game.Actor)
    enemies := c.getActiveOpponents(ourActor)
    allies := c.getActiveAllies(ourActor)
    for _, enemy := range enemies {
        enemyPositions[enemy.Pos()] = enemy
    }
    for _, ally := range allies {
        allyPositions[ally.Pos()] = ally
    }

    activeSkills := ourActor.GetActiveSkills()
    spells := ourActor.GetEquippedSpells()

    /*
       canMeleeAttack := true
       canRangedAttack := ourActor.HasRangedWeaponEquipped()
       canUseActiveSkill := len(activeSkills) > 0
       canUseSpell := len(spells) > 0

    */

    // assign utility to this position first

    var neighborsWithAllies, neighborsWithEnemies []geometry.Point
    neighbors := gridMap.GetAllCardinalNeighbors(potentialNewLocation)
    for _, neighbor := range neighbors {
        if gridMap.IsActorAt(neighbor) {
            actorAt := gridMap.ActorAt(neighbor)
            if actorAt == ourActor {
                continue
            }
            if actorAt.IsAlive() {
                if c.areAllies(ourActor, actorAt) {
                    neighborsWithAllies = append(neighborsWithAllies, neighbor)
                } else {
                    neighborsWithEnemies = append(neighborsWithEnemies, neighbor)
                }
            }
        }
    }

    maxSkillUtility := 0
    var bestSkill game.Action
    var bestLocation geometry.Point
    for _, skill := range append(activeSkills, spells...) {
        if skill.IsTargeted() {
            validTargetsOfSkill := skill.GetValidTargets(c.engine, ourActor, potentialNewLocation)
            for target, _ := range validTargetsOfSkill {
                combatUtility := skill.GetCombatUtilityForTargetedUseOnLocation(c.engine, ourActor, target, allyPositions, enemyPositions)
                if combatUtility > maxSkillUtility {
                    maxSkillUtility = combatUtility
                    bestSkill = skill
                    bestLocation = target
                }
            }
        } else {
            combatUtility := skill.GetCombatUtilityForUseAtLocation(c.engine, ourActor, potentialNewLocation, allyPositions, enemyPositions)
            if combatUtility > maxSkillUtility {
                maxSkillUtility = combatUtility
                bestSkill = skill
                bestLocation = potentialNewLocation
            }
        }
    }
    positionUtility := 0
    if maxSkillUtility > 0 {
        bestAction = BattleAction{
            MovementDestination:  potentialNewLocation,
            ActionType:           AttackActionTypeActiveSkill,
            ActiveAction:         bestSkill,
            ActionTargetLocation: &bestLocation,
        }
        positionUtility += c.getRangedUtility(neighborsWithAllies, neighborsWithEnemies)
    } else {
        positionUtility += c.getMeleeUtility(neighborsWithAllies, neighborsWithEnemies)
    }

    distance := geometry.DistanceManhattan(ourActor.Pos(), potentialNewLocation)
    positionUtility -= distance

    utility = positionUtility + maxSkillUtility
    return utility, bestAction
}

func (c *CombatState) getRangedUtility(neighborsWithAllies, neighborsWithEnemies []geometry.Point) int {
    rangedUtility := 0
    if len(neighborsWithEnemies) == 0 {
        rangedUtility += 30
    } else if len(neighborsWithEnemies) > 1 {
        rangedUtility -= 30
    }

    if len(neighborsWithAllies) == 0 { // by, default, leave a gap
        rangedUtility += 10
    }
    return rangedUtility
}
func (c *CombatState) getMeleeUtility(neighborsWithAllies, neighborsWithEnemies []geometry.Point) int {
    meleeUtility := 0
    if len(neighborsWithEnemies) == 1 {
        meleeUtility += 30
    } else if len(neighborsWithEnemies) > 1 {
        meleeUtility -= 10
    }

    if len(neighborsWithAllies) == 0 { // by, default, leave a gap
        meleeUtility += 10
    }
    return meleeUtility
}
func (c *CombatState) getCombatPathTo(ourActor *game.Actor, dijkstraMap map[geometry.Point]int, dest geometry.Point) []geometry.Point {
    if ourActor.Pos() == dest { // we're already there
        return []geometry.Point{}
    }
    if _, ok := dijkstraMap[dest]; !ok { // we can't reach this position
        return []geometry.Point{}
    }
    path := []geometry.Point{dest}
    currentPos := dest
    for currentPos != ourActor.Pos() {
        currentDistInMap := dijkstraMap[currentPos] // roll down from here..
        for _, neighbor := range c.engine.currentMap.GetAllCardinalNeighbors(currentPos) {
            if distOfNeighbor, ok := dijkstraMap[neighbor]; ok && distOfNeighbor < currentDistInMap {
                path = append(path, neighbor)
                currentPos = neighbor
                break
            }
        }
    }
    // reverse the path
    slices.Reverse(path)
    return path
}
func (c *CombatState) areAllies(one *game.Actor, two *game.Actor) bool {
    return one.GetCombatFaction() == two.GetCombatFaction()
}

func (c *CombatState) neighbouringEnemies(actor *game.Actor) []*game.Actor {
    var enemies []*game.Actor
    gridMap := c.engine.GetGridMap()
    neighbors := gridMap.GetAllCardinalNeighbors(actor.Pos())
    for _, neighbor := range neighbors {
        if gridMap.IsActorAt(neighbor) {
            actorAt := gridMap.ActorAt(neighbor)
            if !c.areAllies(actor, actorAt) {
                enemies = append(enemies, actorAt)
            }
        }
    }
    return enemies
}

func (c *CombatState) getPositionsOfVisibleOpponentsOfActor(attacker *game.Actor) map[geometry.Point]bool {
    visibleEnemies := c.getOpponents(attacker, func(enemy *game.Actor) bool {
        hasLoS := c.canSee(attacker, enemy)
        return hasLoS && enemy.IsAlive() && !enemy.IsHidden() && !enemy.IsSleeping()
    })
    visibleEnemyPositions := make(map[geometry.Point]bool)
    for _, enemy := range visibleEnemies {
        visibleEnemyPositions[enemy.Pos()] = true
    }
    return visibleEnemyPositions
}
func (c *CombatState) canSee(one, other *game.Actor) bool {
    los := c.getLineOfSight(one.Pos(), other.Pos(), c.isActorBlockingSightForAttacker(one))
    return len(los) > 0
}
func (c *CombatState) getOpponents(actor *game.Actor, keep func(actor *game.Actor) bool) []*game.Actor {
    var listOfOpponents []*game.Actor
    if c.engine.IsPlayerControlled(actor) {
        listOfOpponents = toList(c.opponents)
    } else {
        listOfOpponents = c.engine.GetPartyMembers()
    }

    var opponents []*game.Actor
    for _, opponent := range listOfOpponents {
        if keep(opponent) {
            opponents = append(opponents, opponent)
        }
    }
    return opponents
}

func (c *CombatState) getActiveOpponents(actor *game.Actor) []*game.Actor {
    var listOfOpponents []*game.Actor
    if c.engine.IsPlayerControlled(actor) {
        listOfOpponents = toList(c.opponents)
    } else {
        listOfOpponents = c.engine.GetPartyMembers()
    }

    var opponents []*game.Actor
    for _, opponent := range listOfOpponents {
        if opponent.IsAlive() && !opponent.IsSleeping() && !opponent.IsHidden() && opponent != actor {
            opponents = append(opponents, opponent)
        }
    }
    return opponents
}

func (c *CombatState) getActiveAllies(actor *game.Actor) []*game.Actor {
    var listOfAllies []*game.Actor
    if c.engine.IsPlayerControlled(actor) {
        listOfAllies = c.engine.GetPartyMembers()
    } else {
        listOfAllies = toList(c.opponents)
    }

    var opponents []*game.Actor
    for _, opponent := range listOfAllies {
        if opponent.IsAlive() && !opponent.IsSleeping() && !opponent.IsHidden() && opponent != actor {
            opponents = append(opponents, opponent)
        }
    }
    return opponents
}

func (c *CombatState) getVisibleOpponentsOfParty() map[geometry.Point]bool {
    allVisibleOpponents := make(map[geometry.Point]bool)
    for _, partyMember := range c.engine.GetPartyMembers() {
        visibleOpponents := c.getPositionsOfVisibleOpponentsOfActor(partyMember)
        for pos, _ := range visibleOpponents {
            allVisibleOpponents[pos] = true
        }
    }
    return allVisibleOpponents
}

func toList(opponents map[*game.Actor]bool) []*game.Actor {
    var list []*game.Actor
    for opponent, _ := range opponents {
        list = append(list, opponent)
    }
    return list
}
