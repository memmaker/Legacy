package ui

import (
    "Legacy/ega"
    "Legacy/geometry"
    "Legacy/renderer"
    "Legacy/util"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
)

type GridMenu struct {
    ScrollingContent
    topLeft          geometry.Point
    menuItems        []util.MenuItem
    gridRenderer     *renderer.DualGridRenderer
    bottomRight      geometry.Point
    currentSelection int

    shouldClose        bool
    autoCloseOnConfirm bool

    lastAction    func()
    title         string
    drawCharIcons bool
}

func (g *GridMenu) OnMouseWheel(x int, y int, dy float64) bool {
    if g.ScrollingContent.OnMouseWheel(x, y, dy) {
        g.OnMouseMoved(x, y)
    }
    return true
}

func (g *GridMenu) OnCommand(command CommandType) bool {
    switch command {
    case PlayerCommandCancel:
        g.ActionCancel()
    case PlayerCommandConfirm:
        g.ActionConfirm()
    case PlayerCommandUp:
        g.ActionUp()
    case PlayerCommandDown:
        g.ActionDown()
    case PlayerCommandLeft:
        g.ActionLeft()
    case PlayerCommandRight:
        g.ActionRight()
    }
    return true
}

func (g *GridMenu) ActionCancel() {
    g.shouldClose = true
}

func (g *GridMenu) CanBeClosed() bool {
    return true
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

func (g *GridMenu) OnMouseMoved(x int, y int) (bool, Tooltip) {
    relativeLine := g.getLineFromScreenLine(y)
    if relativeLine < 0 || relativeLine >= len(g.menuItems) {
        return false, NoTooltip{}
    }
    if x < g.topLeft.X+1 || x >= g.bottomRight.X-1 {
        return false, NoTooltip{}
    }
    g.currentSelection = relativeLine
    if len(g.menuItems[g.currentSelection].TooltipText) > 0 {
        return true, NewTextTooltip(g.gridRenderer, g.menuItems[g.currentSelection].TooltipText, geometry.Point{X: x, Y: y})
    }
    return true, NoTooltip{}
}

func (g *GridMenu) OnMouseClicked(x int, y int) bool {
    relativeLine := g.getLineFromScreenLine(y)
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
        g.ScrollToBottom()
        return
    }
    g.MakeIndexVisible(g.currentSelection)
}

func (g *GridMenu) ActionDown() {
    g.currentSelection++
    if g.currentSelection >= len(g.menuItems) {
        g.currentSelection = 0
        g.ScrollToTop()
    }
    g.MakeIndexVisible(g.currentSelection)
}

func NewGridMenu(gridRenderer *renderer.DualGridRenderer, menuItems []util.MenuItem) *GridMenu {
    topLeft, bottomRight := positionGridMenu(gridRenderer, menuItems, "")
    g := &GridMenu{
        gridRenderer: gridRenderer,
        topLeft:      topLeft,
        bottomRight:  bottomRight,
        menuItems:    menuItems,
    }
    neededHeight := func() int { return len(menuItems) }
    availableSpace := func() geometry.Rect { return geometry.NewRect(g.topLeft.X+1, g.topLeft.Y+1, g.bottomRight.X-1, g.bottomRight.Y-1) }
    g.ScrollingContent = NewScrollingContentWithFunctions(neededHeight, availableSpace)
    return g
}

func positionGridMenu(gridRenderer *renderer.DualGridRenderer, menuItems []util.MenuItem, title string) (geometry.Point, geometry.Point) {
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
func NewGridMenuAtY(gridRenderer *renderer.DualGridRenderer, menuItems []util.MenuItem, yOffset int) *GridMenu {
    height := min(len(menuItems)+2, 15)
    width := min(maxLenOfItems(menuItems)+2, 36)
    screenGridSize := gridRenderer.GetSmallGridScreenSize()
    // center
    topLeft := geometry.Point{
        X: (screenGridSize.X - width) / 2,
        Y: yOffset,
    }
    bottomRight := geometry.Point{X: topLeft.X + width, Y: topLeft.Y + height}
    g := &GridMenu{
        gridRenderer: gridRenderer,
        topLeft:      topLeft,
        bottomRight:  bottomRight,
        menuItems:    menuItems,
    }
    neededHeight := func() int { return len(menuItems) }
    availableSpace := func() geometry.Rect { return geometry.NewRect(g.topLeft.X+1, g.topLeft.Y+1, g.bottomRight.X-1, g.bottomRight.Y-1) }
    g.ScrollingContent = NewScrollingContentWithFunctions(neededHeight, availableSpace)
    return g
}
func (g *GridMenu) SetAutoClose() {
    g.autoCloseOnConfirm = true
}
func (g *GridMenu) Draw(screen *ebiten.Image) {
    var textColor color.Color
    g.gridRenderer.DrawFilledBorder(screen, g.topLeft, g.bottomRight, g.title)
    for y := g.topLeft.Y + 1; y < g.bottomRight.Y-1; y++ {
        relativeItemOffset := g.getLineFromScreenLine(y)
        if relativeItemOffset >= len(g.menuItems) {
            break
        }
        item := g.menuItems[relativeItemOffset]
        textColor = color.White
        if relativeItemOffset == g.currentSelection {
            textColor = ega.BrightGreen
        } else if item.TextColor != nil {
            textColor = item.TextColor
        }
        xPos := g.topLeft.X + 1
        if item.CharIcon > 0 {
            g.gridRenderer.DrawOnSmallGrid(screen, xPos, y, item.CharIcon)
            g.gridRenderer.DrawColoredString(screen, xPos+1, y, item.Text, textColor)
        } else {
            g.gridRenderer.DrawColoredString(screen, xPos, y, item.Text, textColor)
        }
    }
    g.drawScrollIndicators(g.gridRenderer, screen)
}

func (g *GridMenu) SetTitle(title string) {
    g.title = title
    g.topLeft, g.bottomRight = positionGridMenu(g.gridRenderer, g.menuItems, title)
}

func (g *GridMenu) PositionNearMouse(x int, y int) {
    smallScreenSize := g.gridRenderer.GetSmallGridScreenSize()
    width := g.bottomRight.X - g.topLeft.X
    height := g.bottomRight.Y - g.topLeft.Y

    // try position the rect so that the mouse is at an 1,1 offset relative to the topleft
    newTopLeft := geometry.Point{
        X: x - 1,
        Y: y - 1,
    }
    newBottomRight := geometry.Point{
        X: newTopLeft.X + width,
        Y: newTopLeft.Y + height,
    }

    // make sure we are within the screen, if not adjust slightly

    if newBottomRight.X > smallScreenSize.X {
        newTopLeft.X = smallScreenSize.X - width
        newBottomRight.X = smallScreenSize.X
    }
    if newBottomRight.Y > smallScreenSize.Y {
        newTopLeft.Y = smallScreenSize.Y - height
        newBottomRight.Y = smallScreenSize.Y
    }

    if newTopLeft.X < 0 {
        newTopLeft.X = 0
        newBottomRight.X = width
    }
    if newTopLeft.Y < 0 {
        newTopLeft.Y = 0
        newBottomRight.Y = height
    }

    g.topLeft = newTopLeft
    g.bottomRight = newBottomRight
}

func (g *GridMenu) DisableAutoClose() {
    g.autoCloseOnConfirm = false
}

func maxLenOfItems(items []util.MenuItem) int {
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
