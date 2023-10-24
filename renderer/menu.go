package renderer

import (
    "Legacy/ega"
    "Legacy/geometry"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
)

type MenuItem struct {
    Text   string
    Action func()
}

type GridMenu struct {
    topLeft          geometry.Point
    menuItems        []MenuItem
    gridRenderer     *DualGridRenderer
    bottomRight      geometry.Point
    currentSelection int
}

func (g *GridMenu) OnMouseMoved(x int, y int) {
    relativeLine := y - g.topLeft.Y - 1
    if relativeLine < 0 || relativeLine >= len(g.menuItems) {
        return
    }
    if x < g.topLeft.X+1 || x >= g.bottomRight.X-1 {
        return
    }
    g.currentSelection = relativeLine
}

func (g *GridMenu) OnMouseClicked(x int, y int) bool {
    relativeLine := y - g.topLeft.Y - 1
    if relativeLine < 0 || relativeLine >= len(g.menuItems) {
        return false
    }
    if x < g.topLeft.X+1 || x >= g.bottomRight.X-1 {
        return false
    }
    g.currentSelection = relativeLine
    g.ActionConfirm()
    return true
}

func (g *GridMenu) ActionConfirm() bool {
    g.menuItems[g.currentSelection].Action()
    return true
}

func (g *GridMenu) ActionUp() {
    g.currentSelection--
    if g.currentSelection < 0 {
        g.currentSelection = len(g.menuItems) - 1
    }
}

func (g *GridMenu) ActionDown() {
    g.currentSelection++
    if g.currentSelection >= len(g.menuItems) {
        g.currentSelection = 0
    }
}

func NewGridMenu(gridRenderer *DualGridRenderer, topLeft geometry.Point, menuItems []MenuItem) *GridMenu {
    height := min(len(menuItems)+2, 15)
    width := min(maxLenOfItems(menuItems)+2, 36)
    return &GridMenu{
        gridRenderer: gridRenderer,
        topLeft:      topLeft,
        bottomRight:  geometry.Point{X: topLeft.X + width, Y: topLeft.Y + height},
        menuItems:    menuItems,
    }
}

func (g *GridMenu) Draw(screen *ebiten.Image) {
    var textColor color.Color
    g.gridRenderer.DrawFilledBorder(screen, g.topLeft, g.bottomRight)
    for i, item := range g.menuItems {
        textColor = color.White
        if i == g.currentSelection {
            textColor = ega.BrightGreen
        }
        g.gridRenderer.DrawColoredString(screen, g.topLeft.X+1, g.topLeft.Y+1+i, item.Text, textColor)
    }
}
func maxLenOfItems(items []MenuItem) int {
    maxLength := 0
    for _, item := range items {
        if len(item.Text) > maxLength {
            maxLength = len(item.Text)
        }
    }
    return maxLength
}
