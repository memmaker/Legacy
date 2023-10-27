package main

import (
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/renderer"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *GridEngine) handleInput() {
    if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
        g.onMove(geometry.Point{X: 1, Y: 0})
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
        g.onMove(geometry.Point{X: -1, Y: 0})
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
        if g.inputElement != nil {
            g.inputElement.ActionUp()
        } else if g.modalElement != nil {
            g.modalElement.ActionUp()
        } else {
            g.onMove(geometry.Point{X: 0, Y: -1})
        }
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
        if g.inputElement != nil {
            g.inputElement.ActionDown()
        } else if g.modalElement != nil {
            g.modalElement.ActionDown()
        } else {
            g.onMove(geometry.Point{X: 0, Y: 1})
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

    if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
        if g.currentMap.IsItemAt(g.avatar.Pos()) {
            item := g.currentMap.ItemAt(g.avatar.Pos())
            g.PickUpItem(item)
        }
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyI) {
        g.openPartyInventory()
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
        if g.inputElement != nil {
            g.inputElement = nil
        }
        if g.modalElement != nil {
            g.modalElement = nil
        }
    }

    if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
        g.openCharDetails(0)
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
        //g.openCharDetails(1)
        g.StartConversation(game.NewActor("Tim", 22))

    } else if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
        //g.openCharDetails(2)
        g.openMenu([]renderer.MenuItem{
            {Text: "Item 1", Action: func() {
                println("Item 1")
            }},
            {Text: "Item 2", Action: func() {
                println("Item 2")
            }},
        })
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF4) {
        g.openCharDetails(3)
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF5) {
        g.openPartyMenu()
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF10) {
        g.openDebugMenu()
    }

    if inpututil.IsKeyJustPressed(ebiten.KeyR) {
        g.lastSelectedAction()
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
