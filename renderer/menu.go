package renderer

import (
    "Legacy/ega"
    "Legacy/geometry"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
)

type MenuItem struct {
    Text      string
    Action    func()
    TextColor color.Color
}

type GridMenu struct {
    topLeft          geometry.Point
    menuItems        []MenuItem
    gridRenderer     *DualGridRenderer
    bottomRight      geometry.Point
    currentSelection int

    shouldClose        bool
    autoCloseOnConfirm bool

    lastAction func()
}

func (g *GridMenu) ShouldClose() bool {
    return g.shouldClose
}

func (g *GridMenu) GetLastAction() func() {
    return g.lastAction
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

func (g *GridMenu) OnMouseClicked(x int, y int) {
    relativeLine := y - g.topLeft.Y - 1
    if relativeLine < 0 || relativeLine >= len(g.menuItems) {
        return
    }
    if x < g.topLeft.X+1 || x >= g.bottomRight.X-1 {
        return
    }
    g.currentSelection = relativeLine
    g.ActionConfirm()
}

func (g *GridMenu) ActionConfirm() {
    action := g.menuItems[g.currentSelection].Action
    g.lastAction = action
    action()
    if g.autoCloseOnConfirm {
        g.shouldClose = true
    }
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

func NewGridMenu(gridRenderer *DualGridRenderer, menuItems []MenuItem) *GridMenu {
    height := min(len(menuItems)+2, 15)
    width := min(maxLenOfItems(menuItems)+2, 36)
    screenGridSize := gridRenderer.GetSmallGridScreenSize()
    // center
    topLeft := geometry.Point{
        X: (screenGridSize.X - width) / 2,
        Y: (screenGridSize.Y - height) / 4,
    }
    bottomRight := geometry.Point{X: topLeft.X + width, Y: topLeft.Y + height}
    return &GridMenu{
        gridRenderer: gridRenderer,
        topLeft:      topLeft,
        bottomRight:  bottomRight,
        menuItems:    menuItems,
    }
}
func (g *GridMenu) SetAutoClose() {
    g.autoCloseOnConfirm = true
}
func (g *GridMenu) Draw(screen *ebiten.Image) {
    var textColor color.Color
    g.gridRenderer.DrawFilledBorder(screen, g.topLeft, g.bottomRight)
    for i, item := range g.menuItems {
        textColor = color.White
        if i == g.currentSelection {
            textColor = ega.BrightGreen
        } else if item.TextColor != nil {
            textColor = item.TextColor
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
