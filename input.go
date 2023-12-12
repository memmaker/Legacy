package main

import (
    "Legacy/geometry"
    "Legacy/recfile"
    "Legacy/ui"
    "Legacy/util"
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *GridEngine) ActionUp() {
    g.movementRoutine.Stop()
    g.PlayerMovement(geometry.Point{X: 0, Y: -1})
}

func (g *GridEngine) ActionDown() {
    g.movementRoutine.Stop()
    g.PlayerMovement(geometry.Point{X: 0, Y: 1})
}

func (g *GridEngine) ActionRight() {
    g.movementRoutine.Stop()
    g.PlayerMovement(geometry.Point{X: 1, Y: 0})
}

func (g *GridEngine) ActionLeft() {
    g.movementRoutine.Stop()
    g.PlayerMovement(geometry.Point{X: -1, Y: 0})
}

func (g *GridEngine) ActionConfirm() {
    g.openContextMenu(g.getContextActions())
}

func (g *GridEngine) ActionCancel() {

}
func (g *GridEngine) isBeingPressedRepeatedly(key ebiten.Key) bool {
    pressedForTicks := inpututil.KeyPressDuration(key)
    ticksBetweenMovement := 20
    if pressedForTicks > 30 {
        ticksBetweenMovement = 8
    }
    isRepetitionTick := pressedForTicks > 0 && pressedForTicks%ticksBetweenMovement == 0
    return inpututil.IsKeyJustPressed(key) || (ebiten.IsKeyPressed(key) && isRepetitionTick)
}
func (g *GridEngine) handleKeyboardInput(actionReceiver ui.InputReceiver) bool {

    if g.isBeingPressedRepeatedly(ebiten.KeyArrowRight) {
        return actionReceiver.OnCommand(ui.PlayerCommandRight)
    }
    if g.isBeingPressedRepeatedly(ebiten.KeyArrowLeft) {
        return actionReceiver.OnCommand(ui.PlayerCommandLeft)
    }
    if g.isBeingPressedRepeatedly(ebiten.KeyArrowUp) {
        return actionReceiver.OnCommand(ui.PlayerCommandUp)
    }
    if g.isBeingPressedRepeatedly(ebiten.KeyArrowDown) {
        return actionReceiver.OnCommand(ui.PlayerCommandDown)
    }

    if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
        return actionReceiver.OnCommand(ui.PlayerCommandConfirm)
    }

    if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
        return actionReceiver.OnCommand(ui.PlayerCommandCancel)
    }

    return false
}

func (g *GridEngine) handleMouseInput(actionReceiver ui.InputReceiver, screenX int, screenY int) bool {
    // mouse
    cellX, cellY := g.gridRenderer.ScreenToSmallCell(screenX, screenY)

    if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
        g.lastInteractionWasMouse = true
        mouseHandled := actionReceiver.OnMouseClicked(cellX, cellY)
        return mouseHandled
    } else {
        g.lastInteractionWasMouse = false
    }
    _, dy := ebiten.Wheel()
    if dy != 0 {
        g.currentTooltip = nil
        mouseHandled := actionReceiver.OnMouseWheel(cellX, cellY, dy)
        if mouseHandled {
            return true
        }
    }

    if cellX != g.lastMousePosX || cellY != g.lastMousePosY {
        mouseHandled, tooltip := actionReceiver.OnMouseMoved(cellX, cellY)
        g.handleTooltip(tooltip)
        g.lastMousePosX = cellX
        g.lastMousePosY = cellY
        return mouseHandled
    }
    return false
}

func (g *GridEngine) handleMapMouseInput(screenX int, screenY int) {
    if !g.IsScreenPosInsideMap(screenX, screenY) {
        return
    }
    if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
        mapPos := g.ScreenToMap(screenX, screenY)
        g.onMapClicked(mapPos)
        return
    }
    if g.altIsPressed {
        mapPos := g.ScreenToMap(screenX, screenY)
        if g.currentMap.IsActorAt(mapPos) {
            actor := g.currentMap.ActorAt(mapPos)
            cellX, cellY := g.gridRenderer.ScreenToSmallCell(screenX, screenY)
            g.handleTooltipWithDelay(ui.NewTextTooltip(g.gridRenderer, actor.LookDescription(), geometry.Point{X: cellX, Y: cellY}), 0)
        }
    }
    if g.debugInfoMode {
        mapPos := g.ScreenToMap(screenX, screenY)
        cellX, cellY := g.gridRenderer.ScreenToSmallCell(screenX, screenY)
        toolTipText := g.getDebugInfo(mapPos)
        g.handleTooltipWithDelay(ui.NewTextTooltip(g.gridRenderer, toolTipText, geometry.Point{X: cellX, Y: cellY}), 0)
    }
}

func (g *GridEngine) getDebugInfo(mapPos geometry.Point) []string {
    if g.currentMap.IsActorAt(mapPos) {
        actor := g.currentMap.ActorAt(mapPos)
        meleeHitChance := fmt.Sprintf("Melee CtH: %d%%", int(g.rules.GetMeleeHitChance(g.GetAvatar(), actor)*100))
        rangedHitChance := fmt.Sprintf("Ranged CtH: %d%%", int(g.rules.GetRangedHitChance(g.GetAvatar(), actor)*100))
        debugInfos := append([]string{mapPos.Encode()}, actor.GetDebugInfos()...)
        debugInfos = append(debugInfos, meleeHitChance, rangedHitChance)
        return debugInfos
    }
    if g.currentMap.IsDownedActorAt(mapPos) {
        actor := g.currentMap.DownedActorAt(mapPos)
        return append([]string{mapPos.Encode()}, actor.GetDebugInfos()...)
    }
    if g.currentMap.IsItemAt(mapPos) {
        item := g.currentMap.ItemAt(mapPos)
        return append([]string{mapPos.Encode()}, item.GetDebugInfos()...)
    }
    if g.currentMap.IsObjectAt(mapPos) {
        object := g.currentMap.ObjectAt(mapPos)
        return append([]string{mapPos.Encode()}, object.GetDebugInfos()...)
    }
    return []string{mapPos.Encode()}
}

func (g *GridEngine) IsScreenPosInsideMap(x int, y int) bool {
    smallX, smallY := g.gridRenderer.ScreenToSmallCell(x, y)
    screenSize := g.gridRenderer.GetSmallGridScreenSize()
    // 1 cell border at top, left, right and 2 cells at the bottom
    return smallX >= 1 && smallX < screenSize.X-1 && smallY >= 1 && smallY < screenSize.Y-2
}

func (g *GridEngine) onMapClicked(pos geometry.Point) {
    if g.GetAvatar().Pos() == pos {
        g.openContextMenu(g.getContextActions())
        return
    }
    dist := geometry.DistanceManhattan(g.GetAvatar().Pos(), pos)
    if dist == 1 {
        interactables := g.getInteractablesAt([]geometry.Point{pos})
        if !interactables.IsEmpty() {
            g.openContextMenu(g.contextActionsFromInteractables(interactables))
            return
        }
    }

    if g.currentMap.IsCurrentlyPassable(pos) {
        g.TryMoveAvatarWithPathfinding(pos)
    }
}

func (g *GridEngine) handleDebugKeys() {
    mapPosSlot := -1

    if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
        mapPosSlot = 0
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
        mapPosSlot = 1
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
        mapPosSlot = 2
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF4) {
        mapPosSlot = 3
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF5) {
        mapPosSlot = 4
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF6) {
        mapPosSlot = 5
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF7) {
        mapPosSlot = 6
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF8) {
        mapPosSlot = 7
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF9) {
        mapPosSlot = 8
    }
    if mapPosSlot >= 0 {
        posKey := fmt.Sprintf("debug_map_pos_%d", mapPosSlot)
        if g.altIsPressed { // save
            util.Persist(posKey, g.GetMapName()+g.avatar.Pos().Encode())
            g.Print(fmt.Sprintf("Saved map position: %s", g.avatar.Pos().Encode()))
        } else { // recall
            mapPosPred := recfile.StrPredicate(util.Get(posKey))
            mapName := mapPosPred.Name()
            xPos := mapPosPred.GetInt(0)
            yPos := mapPosPred.GetInt(1)
            g.transitionToLocation(mapName, geometry.Point{X: xPos, Y: yPos})
        }
    }

    if inpututil.IsKeyJustPressed(ebiten.KeyF12) {
        ebiten.SetFullscreen(!ebiten.IsFullscreen())
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF10) {
        g.openDebugMenu()
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
        g.debugInfoMode = !g.debugInfoMode
        g.Print(fmt.Sprintf("Debug info mode: %t", g.debugInfoMode))
    }
}

func (g *GridEngine) handleShortcuts() {
    // SHORTCUTS
    // (P)ickup
    // (I)nventory
    // (J)ournal
    // (M)agic
    // (S)earch
    // S(N)eak
    // (B)ows
    // (D)ivide party
    // (T)ry to join party
    // (L)og
    // (C)enter camera on avatar
    // (K)eys
    // cl(o)ck
    // Space - Party menu
    // Enter - Context menu

    // A - repeat last selected (a)ction
    // Q - repeat last (q)uote

    // 1-4             - Character equipment (1-4)
    // 5               - Party overview
    // D + 1-4         - Character details (1-4)
    // F + 1-4         - Character status effects (1-4)
    // Shift + 1-4 - Switch control to character (1-4)
    // O + 1-4     - Optimize equip for character (1-4)
    // U + 1-4     - Strip gear from character (1-4)
    // F12 - Toggle fullscreen
    g.altIsPressed = ebiten.IsKeyPressed(ebiten.KeyAlt)

    if inpututil.IsKeyJustPressed(ebiten.KeyP) && !g.IsWindowOpen() {
        if g.currentMap.IsItemAt(g.avatar.Pos()) {
            item := g.currentMap.ItemAt(g.avatar.Pos())
            g.PickUpItem(item)
        }
    } else if inpututil.IsKeyJustPressed(ebiten.KeyI) {
        //g.openKeyInventory()
        g.openExtendedInventory()
    } else if inpututil.IsKeyJustPressed(ebiten.KeyJ) {
        g.openJournal()
    } else if inpututil.IsKeyJustPressed(ebiten.KeyL) {
        g.openPrintLog()
    } else if inpututil.IsKeyJustPressed(ebiten.KeyM) {
        g.openActiveSkillsMenu(g.GetAvatar(), g.GetAvatar().GetEquippedSpells())
    } else if inpututil.IsKeyJustPressed(ebiten.KeyN) {
        g.isSneaking = !g.isSneaking
        if !g.isSneaking {
            g.ClearOverlay()
            g.Print("You stop sneaking.")
        } else {
            g.updateSneakOverlays()
            g.Print("You start sneaking.")
        }
    } else if inpututil.IsKeyJustPressed(ebiten.KeyS) {
        g.searchForHiddenObjects()
    } else if inpututil.IsKeyJustPressed(ebiten.KeyR) {
        g.TryRestParty()
    } else if inpututil.IsKeyJustPressed(ebiten.KeyB) {
        g.combatManager.PlayerControlledRangedAttack()
    } else if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
        g.openPartyMenu()
    } else if inpututil.IsKeyJustPressed(ebiten.KeyD) {
        if g.playerParty.HasFollowers() {
            g.OpenMenu(g.playerParty.GetSplitActions(g))
        } else {
            g.ShowText([]string{"You don't have any followers."})
        }
    } else if inpututil.IsKeyJustPressed(ebiten.KeyT) {
        if g.playerParty.HasFollowers() {
            g.TryJoinParty()
        } else {
            g.ShowText([]string{"You don't have any followers."})
        }
    } else if inpututil.IsKeyJustPressed(ebiten.KeyK) {
        g.openKeyInventory()
    } else if inpututil.IsKeyJustPressed(ebiten.KeyO) {
        g.Print(g.worldTime.GetTimeAndDate())
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyC) {
        g.mapWindow.CenterOn(g.avatar.Pos())
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyA) {
        if g.lastSelectedAction != nil {
            g.lastSelectedAction()
        }
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
        g.ShowText(g.lastShownText)
    }
    charIndex := -1
    if inpututil.IsKeyJustPressed(ebiten.Key1) {
        charIndex = 0
    } else if inpututil.IsKeyJustPressed(ebiten.Key2) {
        charIndex = 1
    } else if inpututil.IsKeyJustPressed(ebiten.Key3) {
        charIndex = 2
    } else if inpututil.IsKeyJustPressed(ebiten.Key4) {
        charIndex = 3
    } else if inpututil.IsKeyJustPressed(ebiten.Key5) {
        g.openPartyOverView()
    }

    if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
        if g.playerParty.HasFollowers() {
            nextMember := g.playerParty.GetNextActiveMember(g.GetAvatar())
            g.SwitchAvatarTo(nextMember)
        }
    }

    if charIndex >= 0 && charIndex < len(g.playerParty.GetMembers()) {
        if ebiten.IsKeyPressed(ebiten.KeyO) {
            g.playerParty.GetMember(charIndex).AutoEquip(g)
            g.Print(fmt.Sprintf("Equipped %s", g.playerParty.GetMember(charIndex).Name()))
        } else if ebiten.IsKeyPressed(ebiten.KeyU) {
            g.playerParty.GetMember(charIndex).StripGear()
            g.Print(fmt.Sprintf("Stripped gear from %s", g.playerParty.GetMember(charIndex).Name()))
        } else if ebiten.IsKeyPressed(ebiten.KeyShift) {
            nextMember := g.playerParty.GetMember(charIndex)
            if nextMember != g.GetAvatar() {
                g.SwitchAvatarTo(nextMember)
                g.Print(fmt.Sprintf("Switched to %s", nextMember.Name()))
            }
        } else if g.altIsPressed {
            g.openCharSkills(charIndex)
        } else if ebiten.IsKeyPressed(ebiten.KeyF) {
            g.openCharStatusEffects(charIndex)
        } else if ebiten.IsKeyPressed(ebiten.KeyE) {
            g.openCharMainStats(charIndex)
        } else {
            g.OpenEquipmentDetails(charIndex)
        }
    }
}
