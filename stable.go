package main

import (
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/gridmap"
    "Legacy/renderer"
    "github.com/hajimehoshi/ebiten/v2"
)

func (g *GridEngine) playerMovement(direction geometry.Point) {
    g.flags.IncrementFlag("steps_taken")

    if g.splitControlled != nil {
        g.currentMap.MoveActor(g.splitControlled, g.splitControlled.Pos().Add(direction))
        g.onViewedActorMoved(g.splitControlled.Pos())
    } else {
        g.playerParty.Move(direction)
        g.onViewedActorMoved(g.avatar.Pos())
    }

    if g.flags.GetFlag("steps_taken") == 1 {
        g.onVeryFirstStep()
    }
}

func (g *GridEngine) moveActor(npc *game.Actor, dest geometry.Point) {
    g.currentMap.MoveActor(npc, dest)
}

func (g *GridEngine) Update() error {
    if g.wantsToQuit {
        return ebiten.Termination
    }

    if g.playerParty.IsDefeated() {
        g.showGameOver()
        return nil
    }

    if g.ticksForPrint > 0 {
        g.ticksForPrint--
    }
    if g.combatManager.IsInCombat() {
        g.combatManager.Update()
        return nil
    }

    g.handleInput()

    if g.currentEncounter != nil && !g.currentEncounter.IsOver() {
        g.currentEncounter.Update()
    }
    if g.animationRoutine.Running() {
        g.animationRoutine.Update()
    }

    g.WorldTicks++

    return nil
}
func (g *GridEngine) Draw(screen *ebiten.Image) {
    g.drawUIOverlay(screen)

    g.mapRenderer.Draw(g.playerParty.GetFoV(), screen, g.CurrentTick())

    if g.inputElement != nil {
        if g.inputElement.ShouldClose() {
            if gridMenu, ok := g.inputElement.(*renderer.GridMenu); ok {
                lastAction := gridMenu.GetLastAction()
                g.lastSelectedAction = lastAction
            }
            g.inputElement = nil
            g.updateContextActions()
        } else {
            g.inputElement.Draw(screen)
        }
    }
    if g.modalElement != nil {
        if g.modalElement.ShouldClose() && g.inputElement == nil {
            g.modalElement = nil
            g.updateContextActions()
        } else {
            g.modalElement.Draw(screen)
        }
    }
    if g.combatManager.IsInCombat() {
        g.drawWarTimeStatusBar(screen)
        g.combatManager.Draw(screen)
    } else {
        g.drawPeaceTimeStatusBar(screen)
    }
}

func (g *GridEngine) drawPeaceTimeStatusBar(screen *ebiten.Image) {
    g.drawUpperStatusBar(screen)

    if g.ticksForPrint > 0 {
        g.drawPrintMessage(screen, false)
    } else {
        g.drawLowerStatusBar(screen)
    }
}
func (g *GridEngine) drawWarTimeStatusBar(screen *ebiten.Image) {
    g.drawLowerStatusBar(screen)
    if g.ticksForPrint > 0 {
        g.drawPrintMessage(screen, true)
    } else {
        g.drawUpperStatusBar(screen)
    }
}

func (g *GridEngine) transition(transition gridmap.Transition) {

    currentMapName := g.currentMap.GetName()
    nextMapName := transition.TargetMap

    // remove the party from the current map
    g.RemovePartyFromMap(g.currentMap)

    // save it
    g.mapsInMemory[currentMapName] = g.currentMap

    var nextMap *gridmap.GridMap[*game.Actor, game.Item, game.Object]
    var isInMemory bool
    // check if the next map is already loaded
    if nextMap, isInMemory = g.mapsInMemory[nextMapName]; !isInMemory {
        // if not, load it from ldtk
        nextMap = g.loadMap(transition.TargetMap)
    } else {
        g.initMapWindow(nextMap.MapWidth, nextMap.MapHeight)
    }

    // set the new map
    g.currentMap = nextMap

    // add the party to the new map
    g.PlaceParty(transition.TargetPos)
}
func (g *GridEngine) updateContextActions() {

    // NOTE: We need to reverse the dependency here
    // The objects in the world, should provide us with context actions.
    // We should not know about them beforehand.

    loc := g.GetAvatar().Pos()

    g.contextActions = make([]renderer.MenuItem, 0)

    neighborsWithStuff := g.currentMap.NeighborsCardinal(loc, func(p geometry.Point) bool {
        return g.currentMap.Contains(p) && (g.currentMap.IsActorAt(p) || g.currentMap.IsItemAt(p) || g.currentMap.IsObjectAt(p))
    })

    neighborsWithStuff = append(neighborsWithStuff, loc)

    var actorsNearby []*game.Actor
    var uniqueItemsNearby []game.Item
    var allItemsNearby []game.Item
    var objectsNearby []game.Object

    for _, neighbor := range neighborsWithStuff {
        if g.currentMap.IsActorAt(neighbor) {
            actor := g.currentMap.GetActor(neighbor)
            if !actor.IsHidden() && !g.IsPlayerControlled(actor) {
                actorsNearby = append(actorsNearby, actor)
            }
        }
        if g.currentMap.IsItemAt(neighbor) {
            item := g.currentMap.ItemAt(neighbor)
            if !item.IsHidden() {
                allItemsNearby = append(allItemsNearby, item)
                if len(uniqueItemsNearby) == 0 {
                    uniqueItemsNearby = append(uniqueItemsNearby, item)
                } else { // can we stack it?
                    for _, otherItem := range uniqueItemsNearby {
                        if !otherItem.CanStackWith(item) {
                            uniqueItemsNearby = append(uniqueItemsNearby, item)
                        }
                    }
                }
            }
        }
        if g.currentMap.IsObjectAt(neighbor) {
            object := g.currentMap.ObjectAt(neighbor)
            if !object.IsHidden() {
                objectsNearby = append(objectsNearby, object)
            }
        }
    }

    // extended range for talking to NPCs
    twoRangeCardinalRelative := []geometry.Point{
        {X: 0, Y: -2},
        {X: 0, Y: 2},
        {X: -2, Y: 0},
        {X: 2, Y: 0},
    }

    for _, relative := range twoRangeCardinalRelative {
        neighbor := loc.Add(relative)
        if g.currentMap.Contains(neighbor) && g.currentMap.IsActorAt(neighbor) {
            actor := g.currentMap.GetActor(neighbor)
            if !actor.IsHidden() && !g.IsPlayerControlled(actor) {
                actorsNearby = append(actorsNearby, actor)
            }
        }
    }

    if len(allItemsNearby) > 1 {
        g.contextActions = append(g.contextActions, renderer.MenuItem{
            Text: "Pick up all",
            Action: func() {
                for _, item := range allItemsNearby {
                    g.PickUpItem(item)
                }
            },
        })
    }

    for _, a := range actorsNearby {
        actor := a
        g.contextActions = append(g.contextActions, renderer.MenuItem{
            Text: actor.Name(),
            Action: func() {
                g.openMenuWithTitle(actor.Name(), actor.GetContextActions(g))
            },
        })
    }

    for _, i := range uniqueItemsNearby {
        item := i
        g.contextActions = append(g.contextActions, renderer.MenuItem{
            Text: item.Name(),
            Action: func() {
                g.openMenuWithTitle(item.Name(), item.GetContextActions(g))
            },
        })
    }

    for _, o := range objectsNearby {
        object := o
        g.contextActions = append(g.contextActions, renderer.MenuItem{
            Text: object.Name(),
            Action: func() {
                g.openMenuWithTitle(object.Name(), object.GetContextActions(g))
            },
        })
    }

}
