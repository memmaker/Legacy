package main

import (
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/gocoro"
    "Legacy/renderer"
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "time"
)

// CombatState should handle
// Participants
// Switching to turn-based
// Turn order
// Player input & Actions
// AI
type CombatState struct {
    opponents            []*game.Actor
    engine               *GridEngine
    movesTakenThisTurn   map[*game.Actor]int
    hasUsedPrimaryAction map[*game.Actor]bool
    animationRoutine     gocoro.Coroutine
    hitAnimations        []HitAnimation
    isPlayerTurn         bool
}
type HitAnimation struct {
    Position  geometry.Point
    TicksLeft int
    Icon      int
    WhenDone  func()
}

func NewCombatManager(gridEngine *GridEngine) *CombatState {
    return &CombatState{
        opponents:            []*game.Actor{},
        movesTakenThisTurn:   make(map[*game.Actor]int),
        hasUsedPrimaryAction: make(map[*game.Actor]bool),
        engine:               gridEngine,
        animationRoutine:     gocoro.NewCoroutine(),
        hitAnimations:        []HitAnimation{},
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

func (c *CombatState) PlayerStartsCombat(attacker *game.Actor, attackedNPCs *game.Actor) {
    // TODO
    // determine who participates in combat
    //  - if part of a group, all group members
    //  - if no one else is around it's just the player and the NPC
    //  - if noise is made, and guards/thugs are nearby, they might join the fight
    c.isPlayerTurn = true
    // let's start with just the player and the NPC
    c.opponents = append(c.opponents, attackedNPCs)

    // when done
    c.attackEnemy(attacker, attackedNPCs)
}

func (c *CombatState) attackEnemy(attacker *game.Actor, victim *game.Actor) {
    c.hasUsedPrimaryAction[attacker] = true
    c.animateHit(attacker, victim)
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
            c.hitAnimations = append(c.hitAnimations[:i], c.hitAnimations[i+1:]...)
            hitAnim.WhenDone()
        } else {
            animationsRunning = true
        }
    }

    if animationsRunning {
        return
    }

    if c.isPlayerTurn {
        if c.canPartyAct() {
            if !c.canAct(c.engine.GetAvatar()) {
                c.switchToNextPartyMember()
            }
            // wait for input
            c.handleInput()
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
        screenPos := c.engine.MapToScreenCoordinates(hitAnim.Position)
        c.engine.gridRenderer.DrawOnBigGrid(screen, screenPos, offsetPoint, hitAnim.Icon)
    }
}

func (c *CombatState) IsInCombat() bool {
    return len(c.opponents) > 0
}

func (c *CombatState) animateHit(attacker *game.Actor, npc *game.Actor) {
    hitPos := npc.Pos()
    c.hitAnimations = append(c.hitAnimations, HitAnimation{
        Position:  hitPos,
        TicksLeft: 35,
        Icon:      194,
        WhenDone: func() {
            c.deliverDamage(attacker, npc)
        },
    })
}
func (c *CombatState) canAct(actor *game.Actor) bool {
    actorMovementAllowance := 5
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
        dest = curPos.Add(direction)
        if isThere, enemy := c.isEnemyAt(dest); isThere {
            c.attackEnemy(avatar, enemy)
        } else {
            c.engine.playerMovement(direction)
            c.movesTakenThisTurn[avatar]++
        }
    }

    if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
        c.openCombatMenu(avatar)
    }
}

func (c *CombatState) openCombatMenu(partyMember *game.Actor) {
    combatOptions := []renderer.MenuItem{
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
}

func (c *CombatState) endEnemyTurn() {
    c.isPlayerTurn = true
    clear(c.movesTakenThisTurn)
    clear(c.hasUsedPrimaryAction)
}

func (c *CombatState) canEnemyAct() bool {
    for _, opponent := range c.opponents {
        if c.canAct(opponent) {
            return true
        }
    }
    return false
}

func (c *CombatState) actorDied(actor *game.Actor) {
    if !c.engine.IsPlayerControlled(actor) {
        for i, opponent := range c.opponents {
            if opponent == actor {
                c.opponents = append(c.opponents[:i], c.opponents[i+1:]...)
                break
            }
        }
    }

    c.engine.actorDied(actor)
}

func (c *CombatState) deliverDamage(attacker, victim *game.Actor) {
    damage := 5
    victim.Damage(damage)
    if !victim.IsAlive() {
        c.actorDied(victim)
    }
    c.engine.Print(fmt.Sprintf("'%s' attacks '%s' for %d damage", attacker.Name(), victim.Name(), damage))
}

func (c *CombatState) takeEnemyTurn() {
    // does nothing for now
    for _, opponent := range c.opponents {
        if c.canAct(opponent) {
            c.engine.Print(fmt.Sprintf("'%s' turn", opponent.Name()))
            enemyMove := c.calculateBestMove(opponent)
            c.animateEnemyMove(opponent, enemyMove)
            c.movesTakenThisTurn[opponent]++
            c.hasUsedPrimaryAction[opponent] = true
            return
        }
    }
}

func (c *CombatState) switchToNextPartyMember() {
    for _, partyMember := range c.engine.GetPartyMembers() {
        if c.canAct(partyMember) {
            c.engine.SwitchAvatarTo(partyMember)
            c.engine.Print(fmt.Sprintf("It's '%s' turn", partyMember.Name()))
            return
        }
    }
}

func (c *CombatState) isEnemyAt(dest geometry.Point) (bool, *game.Actor) {
    for _, opponent := range c.opponents {
        if opponent.Pos() == dest {
            return true, opponent
        }
    }
    return false, nil
}

type EnemyAction struct {
    Movement   []geometry.Point
    ActionType AttackActionType
    Target     *game.Actor
}

func (c *CombatState) calculateBestMove(opponent *game.Actor) EnemyAction {
    // let's start simple..
    // find the nearest enemy, find a free cell position next to him
    // move towards that position
    // if we can reach it this turn: attack
    attackType := AttackActionTypeMelee
    movement, attackTarget := c.closeIntoMeleeRange(opponent)
    return EnemyAction{
        Movement:   movement,
        ActionType: attackType,
        Target:     attackTarget,
    }
}

func (c *CombatState) closeIntoMeleeRange(ourActor *game.Actor) ([]geometry.Point, *game.Actor) {
    // find the nearest enemy
    var nearestEnemy *game.Actor
    var nearestPath []geometry.Point
    currentMap := c.engine.currentMap
    for _, enemy := range c.engine.GetPartyMembers() {
        if enemy.IsAlive() {
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

    // find a free cell next to him
    return nearestPath, nearestEnemy
}

func (c *CombatState) animateEnemyMove(npc *game.Actor, move EnemyAction) {
    _ = c.animationRoutine.Run(func(exe *gocoro.Execution) {
        for _, dest := range move.Movement {
            c.engine.moveActor(npc, dest)
            c.movesTakenThisTurn[npc]++
            _ = exe.YieldTime(200 * time.Millisecond)
        }

        if move.ActionType == AttackActionTypeMelee && move.Target != nil {
            c.animateHit(npc, move.Target)
        }
    })
}
