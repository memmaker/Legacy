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
    CharIcon  int32
}

type GridMenu struct {
    topLeft          geometry.Point
    menuItems        []MenuItem
    gridRenderer     *DualGridRenderer
    bottomRight      geometry.Point
    currentSelection int
    scrollOffset     int

    shouldClose        bool
    autoCloseOnConfirm bool

    lastAction    func()
    title         string
    drawCharIcons bool

    upIndicator   int32
    downIndicator int32
}

func (g *GridMenu) OnAvatarSwitched() {

}

func (g *GridMenu) ActionLeft() {
    g.ActionUp()
}

func (g *GridMenu) ActionRight() {
    g.ActionDown()
}

func (g *GridMenu) ShouldClose() bool {
    return g.shouldClose
}

func (g *GridMenu) GetLastAction() func() {
    return g.lastAction
}

func (g *GridMenu) OnMouseMoved(x int, y int) Tooltip {
    relativeLine := y - g.topLeft.Y - 1
    if relativeLine < 0 || relativeLine >= len(g.menuItems) {
        return NoTooltip{}
    }
    if x < g.topLeft.X+1 || x >= g.bottomRight.X-1 {
        return NoTooltip{}
    }
    g.currentSelection = relativeLine
    return NoTooltip{}
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

func (g *GridMenu) ActionConfirm() {
    if g.menuItems == nil || len(g.menuItems) == 0 {
        return
    }
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
        if g.NeedsScrolling() {
            g.scrollOffset = len(g.menuItems) - 13
        }
        return
    } else if g.currentSelection < g.scrollOffset {
        g.scrollOffset--
    }
}

func (g *GridMenu) ActionDown() {
    g.currentSelection++
    if g.currentSelection >= len(g.menuItems) {
        g.currentSelection = 0
        g.scrollOffset = 0
    } else if g.currentSelection >= g.scrollOffset+13 {
        g.scrollOffset++
    }
}

func (g *GridMenu) NeedsScrolling() bool {
    return len(g.menuItems) > 13
}
func (g *GridMenu) CanScrollUp() bool {
    if !g.NeedsScrolling() {
        return false
    }
    return g.scrollOffset > 0
}
func (g *GridMenu) CanScrollDown() bool {
    if !g.NeedsScrolling() {
        return false
    }
    return g.scrollOffset < len(g.menuItems)-13
}
func NewGridMenu(gridRenderer *DualGridRenderer, menuItems []MenuItem) *GridMenu {
    topLeft, bottomRight := positionGridMenu(gridRenderer, menuItems, "")
    return &GridMenu{
        gridRenderer:  gridRenderer,
        topLeft:       topLeft,
        bottomRight:   bottomRight,
        menuItems:     menuItems,
        downIndicator: 5,
        upIndicator:   6,
    }
}

func positionGridMenu(gridRenderer *DualGridRenderer, menuItems []MenuItem, title string) (geometry.Point, geometry.Point) {
    height := min(len(menuItems)+2, 15)
    width := min(max(maxLenOfItems(menuItems)+2, len(title)+4), 36)
    screenGridSize := gridRenderer.GetSmallGridScreenSize()
    // center
    topLeft := geometry.Point{
        X: (screenGridSize.X - width) / 2,
        Y: (screenGridSize.Y - height) / 4,
    }
    bottomRight := geometry.Point{X: topLeft.X + width, Y: topLeft.Y + height}
    return topLeft, bottomRight
}
func NewGridMenuAtY(gridRenderer *DualGridRenderer, menuItems []MenuItem, yOffset int) *GridMenu {
    height := min(len(menuItems)+2, 15)
    width := min(maxLenOfItems(menuItems)+2, 36)
    screenGridSize := gridRenderer.GetSmallGridScreenSize()
    // center
    topLeft := geometry.Point{
        X: (screenGridSize.X - width) / 2,
        Y: yOffset,
    }
    bottomRight := geometry.Point{X: topLeft.X + width, Y: topLeft.Y + height}
    return &GridMenu{
        gridRenderer:  gridRenderer,
        topLeft:       topLeft,
        bottomRight:   bottomRight,
        menuItems:     menuItems,
        downIndicator: 5,
        upIndicator:   6,
    }
}
func (g *GridMenu) SetAutoClose() {
    g.autoCloseOnConfirm = true
}
func (g *GridMenu) Draw(screen *ebiten.Image) {
    var textColor color.Color
    g.gridRenderer.DrawFilledBorder(screen, g.topLeft, g.bottomRight, g.title)
    itemCount := min(len(g.menuItems), 13)
    for i := g.scrollOffset; i < g.scrollOffset+itemCount; i++ {
        item := g.menuItems[i]
        textColor = color.White
        if i == g.currentSelection {
            textColor = ega.BrightGreen
        } else if item.TextColor != nil {
            textColor = item.TextColor
        }
        drawOffset := i - g.scrollOffset
        if item.CharIcon > 0 {
            g.gridRenderer.DrawOnSmallGrid(screen, g.topLeft.X+1, g.topLeft.Y+1+drawOffset, item.CharIcon)
            g.gridRenderer.DrawColoredString(screen, g.topLeft.X+2, g.topLeft.Y+1+drawOffset, item.Text, textColor)
        } else {
            g.gridRenderer.DrawColoredString(screen, g.topLeft.X+1, g.topLeft.Y+1+drawOffset, item.Text, textColor)
        }
    }
    g.drawScrollIndicators(screen)
}

func (g *GridMenu) drawScrollIndicators(screen *ebiten.Image) {
    if g.NeedsScrolling() {
        if g.CanScrollUp() {
            g.gridRenderer.DrawOnSmallGrid(screen, g.bottomRight.X-2, g.topLeft.Y+2, g.upIndicator)
        }
        if g.CanScrollDown() {
            g.gridRenderer.DrawOnSmallGrid(screen, g.bottomRight.X-2, g.bottomRight.Y-3, g.downIndicator)
        }
    }
}

func (g *GridMenu) SetTitle(title string) {
    g.title = title
    g.topLeft, g.bottomRight = positionGridMenu(g.gridRenderer, g.menuItems, title)
}

func maxLenOfItems(items []MenuItem) int {
    maxLength := 0
    hasIcons := false
    for _, item := range items {
        if item.CharIcon > 0 {
            hasIcons = true
        }
        textAsRunes := []rune(item.Text)
        if len(textAsRunes) > maxLength {
            maxLength = len(textAsRunes)
        }
    }
    if hasIcons {
        maxLength += 1
    }
    return maxLength
}
