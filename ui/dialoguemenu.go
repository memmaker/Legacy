package ui

import (
    "Legacy/ega"
    "Legacy/geometry"
    "Legacy/renderer"
    "Legacy/util"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
)

type GridDialogueMenu struct {
    topLeft      geometry.Point
    gridRenderer *renderer.DualGridRenderer
    bottomRight  geometry.Point

    currentSelection int
    hotspotLayout    [][]ButtonHotspot
    lastIndex        int

    shouldClose bool
    title       string
    canBeClosed bool
}

func (g *GridDialogueMenu) OnCommand(command CommandType) bool {
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

func (g *GridDialogueMenu) ActionCancel() {

}

func (g *GridDialogueMenu) CanBeClosed() bool {
    return g.canBeClosed
}

func (g *GridDialogueMenu) OnAvatarSwitched() {

}

func (g *GridDialogueMenu) ActionLeft() {
    g.ActionUp()
}

func (g *GridDialogueMenu) ActionRight() {
    g.ActionDown()
}

func (g *GridDialogueMenu) ShouldClose() bool {
    return g.shouldClose
}

type ButtonHotspot struct {
    ItemIndex int
    StartX    int
    EndX      int
    Label     string
    Action    func()
    TextColor color.Color
}

func NewGridDialogueMenu(gridRenderer *renderer.DualGridRenderer, topLeft geometry.Point, menuItems []util.MenuItem) *GridDialogueMenu {
    width := gridRenderer.GetSmallGridScreenSize().X - 6

    hotspotLayout := layoutMenuItems(menuItems, width)
    height := min(len(hotspotLayout)+4, 15)

    return &GridDialogueMenu{
        gridRenderer:  gridRenderer,
        topLeft:       topLeft,
        bottomRight:   geometry.Point{X: topLeft.X + width, Y: topLeft.Y + height},
        hotspotLayout: hotspotLayout,
        lastIndex:     len(menuItems) - 1,
        canBeClosed:   true,
    }
}

func (g *GridDialogueMenu) Draw(screen *ebiten.Image) {
    var textColor color.Color
    g.gridRenderer.DrawFilledBorder(screen, g.topLeft, g.bottomRight, g.title)
    for line, items := range g.hotspotLayout {
        for _, item := range items {
            textColor = color.White
            if item.ItemIndex == g.currentSelection {
                textColor = ega.BrightGreen
            } else if item.TextColor != nil {
                textColor = item.TextColor
            }
            g.gridRenderer.DrawColoredString(screen, item.StartX, g.topLeft.Y+2+line, item.Label, textColor)
        }
    }
}

func (g *GridDialogueMenu) OnMouseClicked(x, y int) bool {
    relativeLine := y - g.topLeft.Y - 2
    if relativeLine < 0 || relativeLine >= len(g.hotspotLayout) {
        return false
    }
    line := g.hotspotLayout[relativeLine]

    for _, hotspot := range line {
        if x >= hotspot.StartX && x < hotspot.EndX {
            hotspot.Action()
            return true
        }
    }
    return false
}

func (g *GridDialogueMenu) OnMouseMoved(x, y int) (bool, Tooltip) {
    relativeLine := y - g.topLeft.Y - 2
    if relativeLine < 0 || relativeLine >= len(g.hotspotLayout) {
        return false, NoTooltip{}
    }
    line := g.hotspotLayout[relativeLine]

    for _, hotspot := range line {
        if x >= hotspot.StartX && x < hotspot.EndX {
            g.currentSelection = hotspot.ItemIndex
            return true, NoTooltip{}
        }
    }
    return false, NoTooltip{}
}

func (g *GridDialogueMenu) ActionConfirm() {
    relativeIndex := g.currentSelection
    currentLine := 0
    for relativeIndex >= len(g.hotspotLayout[currentLine]) {
        relativeIndex -= len(g.hotspotLayout[currentLine])
        currentLine++
    }
    g.hotspotLayout[currentLine][relativeIndex].Action()
}

func (g *GridDialogueMenu) ActionUp() {
    g.currentSelection--
    if g.currentSelection < 0 {
        g.currentSelection = g.lastIndex
    }
}

func (g *GridDialogueMenu) ActionDown() {
    g.currentSelection++
    if g.currentSelection > g.lastIndex {
        g.currentSelection = 0
    }
}

func (g *GridDialogueMenu) SetCannotBeClosed() {
    g.shouldClose = false
    g.canBeClosed = false
}

func layoutMenuItems(items []util.MenuItem, width int) [][]ButtonHotspot {
    result := make([][]ButtonHotspot, 0)
    currentLine := make([]ButtonHotspot, 0)
    xOffset := 5
    currentLineWidth := 0
    for i, item := range items {
        if currentLineWidth+len(item.Text) > width-4 {
            result = append(result, currentLine)
            currentLine = make([]ButtonHotspot, 0)
            currentLineWidth = 0
        }
        currentLine = append(currentLine, ButtonHotspot{
            ItemIndex: i,
            StartX:    xOffset + currentLineWidth,
            EndX:      xOffset + currentLineWidth + len(item.Text) + 1,
            Label:     string(append([]rune{'ө'}, []rune(item.Text)...)),
            Action:    item.Action,
            TextColor: item.TextColor,
        })
        currentLineWidth += len(item.Text) + 2
    }
    if len(currentLine) > 0 {
        result = append(result, currentLine)
    }
    return result
}
