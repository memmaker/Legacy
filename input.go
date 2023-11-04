package main

import (
    "Legacy/geometry"
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *GridEngine) handleInput() {
    if g.textInput != nil {
        var keys []ebiten.Key
        keys = inpututil.AppendJustPressedKeys(keys)
        for _, key := range keys {
            g.textInput.OnKeyPressed(key)
        }
        return
    }

    windowsOpen := g.inputElement != nil || g.modalElement != nil
    if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
        if g.inputElement != nil {
            g.inputElement.ActionRight()
        } else if !windowsOpen {
            g.playerMovement(geometry.Point{X: 1, Y: 0})
        }
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
        if g.inputElement != nil {
            g.inputElement.ActionLeft()
        } else if !windowsOpen {
            g.playerMovement(geometry.Point{X: -1, Y: 0})
        }
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
        if g.inputElement != nil {
            g.inputElement.ActionUp()
        } else if g.modalElement != nil {
            g.modalElement.ActionUp()
        } else {
            g.playerMovement(geometry.Point{X: 0, Y: -1})
        }
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
        if g.inputElement != nil {
            g.inputElement.ActionDown()
        } else if g.modalElement != nil {
            g.modalElement.ActionDown()
        } else {
            g.playerMovement(geometry.Point{X: 0, Y: 1})
        }
    }

    if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
        if g.inputElement != nil {
            g.inputElement.ActionConfirm()
        } else if g.modalElement != nil {
            g.modalElement.ActionConfirm()
        } else if transition, ok := g.currentMap.GetTransitionAt(g.avatar.Pos()); ok {
            g.transition(transition)
        } else {
            g.openContextMenu()
        }
    }
    // SHORTCUTS
    // (P)ickup
    // (I)nventory
    // (J)ournal
    // (M)agic
    // (S)earch
    // (B)ow
    // (D)ivide party
    // (T)ry to join party
    // (L)og
    // Space - Party menu
    // Enter - Context menu

    // A - repeat last selected (a)ction
    // Q - repeat last (q)uote

    // 1-4         - Character details (1-4)
    // Shift + 1-4 - Switch control to character (1-4)
    // E + 1-4     - Auto equip character (1-4)
    // U + 1-4     - Strip gear from character (1-4)
    // F9 - Toggle fullscreen

    if inpututil.IsKeyJustPressed(ebiten.KeyP) && !windowsOpen {
        if g.currentMap.IsItemAt(g.avatar.Pos()) {
            item := g.currentMap.ItemAt(g.avatar.Pos())
            g.PickUpItem(item)
        }
    } else if inpututil.IsKeyJustPressed(ebiten.KeyI) {
        g.openPartyInventory()
    } else if inpututil.IsKeyJustPressed(ebiten.KeyJ) {
        g.openJournal()
    } else if inpututil.IsKeyJustPressed(ebiten.KeyL) {
        g.openPrintLog()
    } else if inpututil.IsKeyJustPressed(ebiten.KeyM) {
        g.openSpellMenu()
    } else if inpututil.IsKeyJustPressed(ebiten.KeyS) {
        g.searchForHiddenObjects()
    } else if inpututil.IsKeyJustPressed(ebiten.KeyR) {
        g.TryRestParty()
    } else if inpututil.IsKeyJustPressed(ebiten.KeyB) {
        g.combatManager.PlayerStartsRangedAttack()
    } else if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
        g.openPartyMenu()
    } else if inpututil.IsKeyJustPressed(ebiten.KeyD) {
        if g.playerParty.HasFollowers() {
            g.openMenu(g.playerParty.GetSplitActions(g))
        } else {
            g.ShowText([]string{"You don't have any followers."})
        }
    } else if inpututil.IsKeyJustPressed(ebiten.KeyT) {
        if g.playerParty.HasFollowers() {
            g.TryJoinParty()
        } else {
            g.ShowText([]string{"You don't have any followers."})
        }
    }

    if inpututil.IsKeyJustPressed(ebiten.KeyA) {
        g.lastSelectedAction()
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
        g.ShowText(g.lastShownText)
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
        if g.inputElement != nil {
            g.inputElement = nil
        }
        if g.modalElement != nil {
            g.modalElement = nil
        }
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

    if charIndex >= 0 && charIndex < len(g.playerParty.GetMembers()) {
        if ebiten.IsKeyPressed(ebiten.KeyE) {
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
        } else if ebiten.IsKeyPressed(ebiten.KeyAlt) {
            g.openCharSkills(charIndex)
        } else if ebiten.IsKeyPressed(ebiten.KeyF) {
            g.openCharBuffs(charIndex)
        } else {
            g.openCharDetails(charIndex)
        }
    }

    if inpututil.IsKeyJustPressed(ebiten.KeyF9) {
        ebiten.SetFullscreen(!ebiten.IsFullscreen())
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF10) {
        g.openDebugMenu()
    }

    cellX, cellY := g.gridRenderer.ScreenToSmallCell(ebiten.CursorPosition())

    if cellX != g.lastMousePosX || cellY != g.lastMousePosY {
        g.onMouseMoved(cellX, cellY)
        g.lastMousePosX = cellX
        g.lastMousePosY = cellY
    }
    if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
        g.onMouseClick(cellX, cellY)
    }
}
