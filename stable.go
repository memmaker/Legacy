package main

import (
    "Legacy/ega"
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/gridmap"
    "Legacy/renderer"
    "Legacy/ui"
    "Legacy/util"
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    "image/color"
    "strings"
)

func (g *GridEngine) PlayerMovement(direction geometry.Point) {
    oldPos := g.GetAvatar().Pos()

    g.playerParty.Move(direction)

    newPos := g.GetAvatar().Pos()

    if oldPos == newPos {
        return
    }

    g.flags.IncrementFlag("steps_taken")
    if g.flags.GetFlag("steps_taken") == 1 {
        g.onVeryFirstStep()
    }
    g.advanceTimeFromMovement()
    if g.playerParty.NeedsRestAfterMovement() {
        for _, actor := range g.playerParty.GetMembers() {
            g.AddStatusEffect(actor, game.StatusWeak(), 1)
        }
    }

    g.onActorMovedOrTeleported(g.currentMap, g.GetAvatar(), newPos)
}

func (g *GridEngine) onActorMovedOrTeleported(loadedMap *gridmap.GridMap[*game.Actor, game.Item, game.Object], actor *game.Actor, newPos geometry.Point) {
    if g.IsPlayerControlled(actor) {
        g.onPartyMemberMovedOrTeleported(loadedMap, actor, newPos)
    } else {
        g.onNPCMovedOrTeleported(loadedMap, actor, newPos)
    }
    if loadedMap.IsObjectAt(newPos) {
        object := g.currentMap.ObjectAt(newPos)
        object.OnActorWalkedOn(actor)
    }
    g.checkMoveHooks(actor, newPos)
}

func (g *GridEngine) onNPCMovedOrTeleported(loadedMap *gridmap.GridMap[*game.Actor, game.Item, game.Object], actor *game.Actor, pos geometry.Point) {
    // calculate the zone of the actor
    engagementZone := loadedMap.GetDijkstraMapWithActorsNotBlocking(pos, actor.GetNPCEngagementRange())
    actor.SetNPCEngagementZone(engagementZone)
}

func (g *GridEngine) checkMoveHooks(actor *game.Actor, newPos geometry.Point) {
    moveHooks := g.levelHooks.MovementHooks
    for i := len(moveHooks) - 1; i >= 0; i-- {
        hook := moveHooks[i]
        if hook.Applies(actor, newPos) {
            hook.Action(actor, newPos)
            if hook.Consume {
                g.levelHooks.MovementHooks = append(g.levelHooks.MovementHooks[:i], g.levelHooks.MovementHooks[i+1:]...)
            }
        }
    }
}

func (g *GridEngine) onPartyMemberMovedOrTeleported(loadedMap *gridmap.GridMap[*game.Actor, game.Item, game.Object], partyMember *game.Actor, newLocation geometry.Point) {

    if partyMember == g.GetAvatar() {
        g.onViewedActorMoved(newLocation)
    }

    if trigger, isAtTrigger := loadedMap.GetNamedTriggerAt(newLocation); isAtTrigger {
        g.TriggerEvent(trigger.Name)
        if trigger.OneShot {
            loadedMap.RemoveNamedTrigger(trigger.Name)
        }
    }
    if g.IsInCombat() {
        return
    }
    if g.isSneaking {
        g.updateSneakOverlays()
    }
    // check if we are near any aggressive actors, that would want to start combat
    for _, actor := range loadedMap.GetFilteredActorsInRadius(newLocation, 11, g.aggressiveActorsFilter(newLocation)) {
        if actor.IsInEngagementZone(newLocation) {
            if !g.isSneaking || !g.SkillCheckVs(partyMember, game.ThievingSkillSneak, actor, game.Perception) {
                g.EnemyStartsCombat(actor)
                return
            }
        }
    }
}
func (g *GridEngine) aggressiveActorsFilter(loc geometry.Point) func(actor *game.Actor) bool {
    return func(actor *game.Actor) bool {
        return actor.IsAggressive() &&
            !g.IsPlayerControlled(actor) &&
            actor.IsAlive() &&
            g.playerParty.CanSee(actor.Pos()) &&
            geometry.DistanceManhattan(actor.Pos(), loc) <= actor.GetNPCEngagementRange()
    }
}
func (g *GridEngine) moveActorInCombat(actor *game.Actor, dest geometry.Point) {
    g.currentMap.MoveActor(actor, dest)
    g.onActorMovedOrTeleported(g.currentMap, actor, dest)
}

func (g *GridEngine) Update() error {
    if g.wantsToQuit {
        return ebiten.Termination
    }

    if g.isGameOver {
        return nil
    } else if g.gameOverConditionReached() {
        g.setGameOver()
    }
    if g.ticksUntilTooltipAppears > 0 {
        g.ticksUntilTooltipAppears--
    }
    if g.ticksForPrint > 0 {
        g.ticksForPrint--
    }

    g.WorldTicks++
    g.advanceTimeFromTicks(g.WorldTicks)

    if g.IsModalOpen() { // prio #0 Text Input fields, never pass through
        textInput, isModalTextInput := g.topModal().(ui.TextInputReceiver)
        if isModalTextInput {
            textInput.SetTick(g.CurrentTick())
            var keys []ebiten.Key
            keys = inpututil.AppendJustPressedKeys(keys)
            for _, key := range keys {
                textInput.OnKeyPressed(key)
            }
            return nil
        }
    }

    var receivers []ui.InputReceiver
    // prio #1 conversation
    if g.IsInConversation() {
        receivers = append(receivers, g.conversationModal)
    }
    // prio #2 modal
    if g.IsModalOpen() {
        receivers = append(receivers, g.topModal())
    }

    if g.IsInCombat() {
        receivers = append(receivers, g.combatManager)
    }
    receivers = append(receivers, g)
    for _, receiver := range receivers {
        if g.handleKeyboardInput(receiver) {
            break
        }
    }
    screenX, screenY := ebiten.CursorPosition()
    mouseHandled := false
    for _, receiver := range receivers {
        if g.handleMouseInput(receiver, screenX, screenY) {
            mouseHandled = true
            break
        }
    }

    if !mouseHandled && !g.IsInCombat() && !g.IsInConversation() && !g.IsModalOpen() {
        g.handleMapMouseInput(screenX, screenY) // make this an option? it can be a bit fiddly..
    }

    g.handleShortcuts()
    g.handleDebugKeys()

    for i := len(g.activeEvents) - 1; i >= 0; i-- {
        event := g.activeEvents[i]
        if event.IsOver() {
            g.activeEvents = append(g.activeEvents[:i], g.activeEvents[i+1:]...)
        } else {
            event.Update()
        }
    }

    if g.animationRoutine.Running() {
        g.animationRoutine.Update()
    }

    g.animator.Update()

    if g.movementRoutine.Running() {
        g.movementRoutine.Update()
    }

    if g.combatManager.IsInCombat() {
        g.combatManager.Update()
    }
    return nil
}

func (g *GridEngine) gameOverConditionReached() bool {
    return g.playerParty.IsDefeated() || g.isGameOver
}
func (g *GridEngine) Draw(screen *ebiten.Image) {
    g.drawUIOverlay(screen)

    g.mapRenderer.Draw(g.playerParty.GetFoV(), screen, g.CurrentTick())
    g.drawMapOverlays(screen)

    if g.combatManager.IsInCombat() {
        g.drawWarTimeStatusBar(screen)
        g.combatManager.Draw(screen)
    } else {
        g.drawPeaceTimeStatusBar(screen)
        g.animator.Draw(g.gridRenderer, screen)
    }

    if g.conversationModal != nil {
        if g.conversationModal.ShouldClose() {
            g.conversationModal = nil
        } else {
            g.conversationModal.Draw(screen)
        }
    }

    if g.IsModalOpen() {
        if g.topModal().ShouldClose() {
            g.PopModal()
        }
        for _, modal := range g.modalStack {
            modal.Draw(screen)
        }
    }

    if g.currentTooltip != nil && g.ticksUntilTooltipAppears == 0 {
        g.currentTooltip.Draw(screen)
    }
}

func (g *GridEngine) drawMapOverlays(screen *ebiten.Image) {
    if len(g.overlayPositions) == 0 {
        return
    }
    offset := g.gridRenderer.GetScaledSmallGridSize()
    offsetPoint := geometry.Point{X: offset, Y: offset}

    icon := int32(229)
    for p, drawColor := range g.overlayPositions {
        screenPos := g.MapToScreenCoordinates(p)
        g.gridRenderer.DrawOnBigGridWithColor(screen, screenPos, offsetPoint, renderer.AtlasEntities, icon, drawColor)
    }
}
func (g *GridEngine) SetOverlay(overlayPositions map[geometry.Point]color.Color) {
    g.overlayPositions = overlayPositions
}
func (g *GridEngine) ClearOverlay() {
    clear(g.overlayPositions)
}
func (g *GridEngine) updateSneakOverlays() {
    g.ClearOverlay()
    for _, actor := range g.currentMap.Actors() {
        if !actor.IsAggressive() {
            continue
        }
        dist := geometry.DistanceManhattan(actor.Pos(), g.GetAvatar().Pos())
        if dist > 10+actor.GetNPCEngagementRange() {
            continue
        }
        zone := actor.GetEngagementZone()
        for pos, _ := range zone {
            if g.IsMapPosOnScreen(pos) && g.playerParty.CanSee(pos) && g.currentMap.IsCurrentlyPassable(pos) {
                g.overlayPositions[pos] = ega.BrightRed
            }
        }
    }
}

func (g *GridEngine) CloseAllModals() {
    g.modalStack = make([]Modal, 0)
    g.currentTooltip = nil
}

func (g *GridEngine) PushModal(nextWidget Modal) {
    if g.IsInConversation() {
        return
    }
    g.modalStack = append(g.modalStack, nextWidget)
    g.currentTooltip = nil
}

func (g *GridEngine) PopModal() {
    g.modalStack = g.modalStack[:len(g.modalStack)-1]
    g.currentTooltip = nil
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
        if g.combatManager.IsAITurn() {
            g.drawTextOnUpperStatusbar(screen, "Wait for your turn...")
        } else {
            g.drawCombatActionBar(screen)
        }
    }
    /*
       else {
           g.drawUpperStatusBar(screen)
       }
    */
}
func (g *GridEngine) TransitionToNamedLocation(targetMap, targetLocation string) {
    g.ensureMapInMemory(targetMap)
    nextMap := g.mapsInMemory[targetMap]
    location := nextMap.GetNamedLocation(targetLocation)
    g.transitionToLocation(targetMap, location)
}
func (g *GridEngine) isMapInMemory(targetMap string) bool {
    _, isInMemory := g.mapsInMemory[targetMap]
    return isInMemory
}
func (g *GridEngine) ensureMapInMemory(targetMap string) {
    // check if the next map is already loaded
    if nextMap, isInMemory := g.mapsInMemory[targetMap]; !isInMemory {
        // if not, load it from ldtk
        if strings.HasPrefix(targetMap, "!gen_") {
            nextMap = g.generateMap(targetMap)
        } else {
            nextMap = g.loadMap(targetMap)
        }
        g.mapsInMemory[targetMap] = nextMap
    }
}
func (g *GridEngine) transitionToLocation(targetMap string, destPos geometry.Point) {
    if g.playerParty.IsInVehicle() {
        g.Print("Must exit vehicle first.")
        return
    }
    currentMapName := g.currentMap.GetName()

    nextMapName := targetMap

    // remove the party from the current map
    g.RemovePartyFromMap(g.currentMap)

    // save it
    g.mapsInMemory[currentMapName] = g.currentMap

    g.ensureMapInMemory(nextMapName)

    nextMap := g.mapsInMemory[nextMapName]

    g.initMapWindow(nextMap.MapWidth, nextMap.MapHeight)

    // set the new map
    g.setMap(nextMap)

    // add the party to the new map
    g.PlaceParty(destPos)
}

func (g *GridEngine) setMap(nextMap *gridmap.GridMap[*game.Actor, game.Item, game.Object]) {
    g.currentMap = nextMap
    g.levelHooks = game.GetHooksForLevel(g, g.currentMap.GetName())
}

type Interactables struct {
    Actors       []*game.Actor
    AllItems     []game.Item
    UniqueItems  []game.Item
    Objects      []game.Object
    SpecialTiles map[geometry.Point]gridmap.MapCell[*game.Actor, game.Item, game.Object]
    Transitions  []gridmap.Transition
}

func (i Interactables) IsEmpty() bool {
    return len(i.Actors) == 0 && len(i.AllItems) == 0 && len(i.Objects) == 0 && len(i.SpecialTiles) == 0
}

func (g *GridEngine) getInteractablesAt(locations []geometry.Point) Interactables {
    var actorsNearby []*game.Actor
    var uniqueItemsNearby []game.Item
    var allItemsNearby []game.Item
    var objectsNearby []game.Object
    var transitions []gridmap.Transition
    nearbySpecialTiles := make(map[geometry.Point]gridmap.MapCell[*game.Actor, game.Item, game.Object])

    for _, n := range locations {
        neighbor := n
        if g.currentMap.IsActorAt(neighbor) {
            actor := g.currentMap.GetActor(neighbor)
            if !actor.IsHidden() && !g.IsPlayerControlled(actor) {
                actorsNearby = append(actorsNearby, actor)
            }
        }
        if g.currentMap.IsDownedActorAt(neighbor) {
            actor := g.currentMap.DownedActorAt(neighbor)
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
        cellAt := g.currentMap.GetCell(neighbor)
        if cellAt.TileType.IsBreakable() || cellAt.TileType.IsBed() {
            nearbySpecialTiles[neighbor] = cellAt
        } else if g.currentMap.IsSpecialAt(neighbor, gridmap.SpecialTileForest) {
            nearbySpecialTiles[neighbor] = cellAt
        } else if transition, ok := g.currentMap.GetTransitionAt(neighbor); ok {
            transitions = append(transitions, transition)
        }
    }
    return Interactables{
        Actors:       actorsNearby,
        AllItems:     allItemsNearby,
        UniqueItems:  uniqueItemsNearby,
        Objects:      objectsNearby,
        SpecialTiles: nearbySpecialTiles,
        Transitions:  transitions,
    }
}
func (g *GridEngine) getContextActions() []util.MenuItem {
    loc := g.GetAvatar().Pos()

    existingNeighbors := g.currentMap.NeighborsCardinal(loc, func(p geometry.Point) bool {
        return g.currentMap.Contains(p)
    })

    existingNeighbors = append(existingNeighbors, loc)

    interactables := g.getInteractablesAt(existingNeighbors)

    // extended range for talking to NPCs
    twoRangeCardinalRelative := []geometry.Point{
        {X: 0, Y: -2},
        {X: 0, Y: 2},
        {X: -2, Y: 0},
        {X: 2, Y: 0},
    }

    for _, relative := range twoRangeCardinalRelative {
        neighbor := loc.Add(relative)
        if g.currentMap.Contains(neighbor) && g.currentMap.IsActorAt(neighbor) && g.playerParty.GetFoV().Visible(neighbor) {
            actor := g.currentMap.GetActor(neighbor)
            if !actor.IsHidden() && !g.IsPlayerControlled(actor) {
                interactables.Actors = append(interactables.Actors, actor)
            }
        }
    }

    return g.contextActionsFromInteractables(interactables)
}

func (g *GridEngine) contextActionsFromInteractables(interactables Interactables) []util.MenuItem {

    contextActions := make([]util.MenuItem, 0)

    for _, t := range interactables.Transitions {
        transition := t
        contextActions = append(contextActions, util.MenuItem{
            Text: "Go to",
            Action: func() {
                g.TransitionToNamedLocation(transition.TargetMap, transition.TargetLocation)
            },
        })
    }

    if len(interactables.AllItems) > 1 {
        contextActions = append(contextActions, util.MenuItem{
            Text: "Pick up all",
            Action: func() {
                for _, item := range interactables.AllItems {
                    g.PickUpItem(item)
                }
            },
        })
    }

    for _, a := range interactables.Actors {
        actor := a
        var additionalActions []util.MenuItem
        actionHooks := g.levelHooks.ActorActionHooks
        for _, hook := range actionHooks {
            if hook.Applies(actor) {
                additionalActions = append(additionalActions, hook.Action(actor)...)
            }
        }
        contextActions = append(contextActions, util.MenuItem{
            Text: actor.Name(),
            Action: func() {
                g.openMenuWithTitle(actor.Name(), append(additionalActions, actor.GetContextActions(g)...))
            },
        })
    }

    for _, i := range interactables.UniqueItems {
        item := i
        contextActions = append(contextActions, util.MenuItem{
            Text: item.Name(),
            Action: func() {
                g.openMenuWithTitle(item.Name(), item.GetContextActions(g))
            },
        })
    }

    for _, o := range interactables.Objects {
        object := o
        contextActions = append(contextActions, util.MenuItem{
            Text: object.Name(),
            Action: func() {
                g.openMenuWithTitle(object.Name(), object.GetContextActions(g))
            },
        })
    }

    breakingTool := g.playerParty.GetNameOfBreakingTool()

    for m, cell := range interactables.SpecialTiles {
        mapPos := m
        if cell.TileType.IsBreakable() && breakingTool != "" {
            contextActions = append(contextActions, util.MenuItem{
                Text: fmt.Sprintf("Break (%s)", breakingTool),
                Action: func() {
                    g.breakTileAt(mapPos)
                },
            })
        } else if cell.TileType.IsBed() {
            if g.GetMapName() == "Bed_Room" {
                contextActions = append(contextActions, util.MenuItem{
                    Text:   "Your bed",
                    Action: g.goBackToBed,
                })
            } else {
                contextActions = append(contextActions, util.MenuItem{
                    Text:   "Rest in bed",
                    Action: g.TryRestParty,
                })
            }
        } else if cell.TileType.IsForest() {
            /* TODO: add these back in
               // removed because of "no clear concept"
               // also: appears multiple times in the context menu
               contextActions = append(contextActions, util.MenuItem{
                   Text: "Hunt game",
                   BaseAction: func() {
                       g.AddFood(1)
                   },
               })
               contextActions = append(contextActions, util.MenuItem{
                   Text: "Gather herbs",
                   BaseAction: func() {
                       g.AddFood(2)
                   },
               })

            */
        }
    }

    return contextActions
}
