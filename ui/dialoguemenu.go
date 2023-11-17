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
    ScrollingContent
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
    width := gridRenderer.GetSmallGridScreenSize().X - 4

    hotspotLayout := layoutMenuItems(menuItems, width)
    height := min(len(hotspotLayout)+4, 11)

    g := &GridDialogueMenu{
        gridRenderer:  gridRenderer,
        topLeft:       topLeft,
        bottomRight:   geometry.Point{X: topLeft.X + width, Y: topLeft.Y + height},
        hotspotLayout: hotspotLayout,
        lastIndex:     len(menuItems) - 1,
        canBeClosed:   true,
    }
    neededHeight := func() int {
        return len(g.hotspotLayout)
    }
    availableSpace := func() geometry.Rect {
        return geometry.NewRect(g.topLeft.X+2, g.topLeft.Y+2, g.bottomRight.X-2, g.bottomRight.Y-2)
    }
    g.ScrollingContent = NewScrollingContentWithFunctions(neededHeight, availableSpace)
    return g
}

func (g *GridDialogueMenu) Draw(screen *ebiten.Image) {
    var textColor color.Color
    g.gridRenderer.DrawFilledBorder(screen, g.topLeft, g.bottomRight, g.title)
    for y := g.topLeft.Y + 2; y < g.bottomRight.Y-2; y++ {
        relativeY := g.getLineFromScreenLine(y)
        if relativeY < 0 || relativeY >= len(g.hotspotLayout) {
            continue
        }
        itemsInLine := g.hotspotLayout[relativeY]
        for _, item := range itemsInLine {
            textColor = color.White
            if item.ItemIndex == g.currentSelection {
                textColor = ega.BrightGreen
            } else if item.TextColor != nil {
                textColor = item.TextColor
            }
            g.gridRenderer.DrawColoredString(screen, item.StartX, y, item.Label, textColor)
        }
    }
    g.drawScrollIndicators(g.gridRenderer, screen)
}

func (g *GridDialogueMenu) OnMouseClicked(x, y int) bool {
    relativeLine := g.getLineFromScreenLine(y)
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
    relativeLine := g.getLineFromScreenLine(y)
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
    hotspot := g.getCurrentButton()
    hotspot.Action()
}

func (g *GridDialogueMenu) getCurrentButton() ButtonHotspot {
    for _, line := range g.hotspotLayout {
        for _, hotspot := range line {
            if hotspot.ItemIndex == g.currentSelection {
                return hotspot
            }
        }
    }
    return ButtonHotspot{}
}

func (g *GridDialogueMenu) getFirstAndLastLineOfSelection() (int, int) {
    firstLine, lastLine := -1, -1
    for i, line := range g.hotspotLayout {
        for _, hotspot := range line {
            if hotspot.ItemIndex == g.currentSelection {
                if firstLine == -1 {
                    firstLine = i
                }
                lastLine = i
            }
        }
    }
    return firstLine, lastLine
}

func (g *GridDialogueMenu) ActionUp() {
    g.currentSelection--
    if g.currentSelection < 0 {
        g.currentSelection = g.lastIndex
        g.ScrollToBottom()
        return
    }
    first, _ := g.getFirstAndLastLineOfSelection()
    g.ScrollingContent.MakeIndexVisible(first)
}

func (g *GridDialogueMenu) ActionDown() {
    g.currentSelection++
    if g.currentSelection > g.lastIndex {
        g.currentSelection = 0
        g.ScrollToTop()
        return
    }
    _, last := g.getFirstAndLastLineOfSelection()
    g.ScrollingContent.MakeIndexVisible(last)
}

func (g *GridDialogueMenu) OnMouseWheel(x int, y int, dy float64) bool {
    if g.ScrollingContent.OnMouseWheel(x, y, dy) {
        g.OnMouseMoved(x, y)
    }
    return true
}

func (g *GridDialogueMenu) SetCannotBeClosed() {
    g.shouldClose = false
    g.canBeClosed = false
}

func layoutMenuItems(items []util.MenuItem, width int) [][]ButtonHotspot {
    result := make([][]ButtonHotspot, 0)
    currentLine := make([]ButtonHotspot, 0)
    xOffset := 4
    currentLineWidth := 0
    for i, item := range items {
        if currentLineWidth+len(item.Text) > width-4 {
            if len(currentLine) > 0 {
                result = append(result, currentLine)
            }
            currentLine = make([]ButtonHotspot, 0)
            currentLineWidth = 0
        }
        if len(item.Text) < width-4 {
            currentLine = append(currentLine, ButtonHotspot{
                ItemIndex: i,
                StartX:    xOffset + currentLineWidth,
                EndX:      xOffset + currentLineWidth + len(item.Text) + 1,
                Label:     "ө" + item.Text,
                Action:    item.Action,
                TextColor: item.TextColor,
            })
            currentLineWidth += len(item.Text) + 2
            continue
        }

        // option is too long..
        if currentLineWidth > 0 { // we are in the middle of a line, commit it first..
            if len(currentLine) > 0 {
                result = append(result, currentLine)
            }
            currentLine = make([]ButtonHotspot, 0)
            currentLineWidth = 0
        }
        broken := util.AutoLayoutWithBreakingPrefix(item.Text, width-4, " ")
        for splitIndex, line := range broken {
            if splitIndex == 0 {
                line = "ө" + line[1:]
            }
            result = append(result, []ButtonHotspot{
                {
                    ItemIndex: i,
                    StartX:    xOffset,
                    EndX:      xOffset + len(line),
                    Label:     line,
                    Action:    item.Action,
                    TextColor: item.TextColor,
                },
            })
        }
    }
    if len(currentLine) > 0 {
        result = append(result, currentLine)
    }
    return result
}
