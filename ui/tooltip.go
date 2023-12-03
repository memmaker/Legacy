package ui

import (
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/renderer"
    "Legacy/util"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
)

type BasicTooltip struct {
    gridRenderer *renderer.DualGridRenderer
    topLeft      geometry.Point
    bottomRight  geometry.Point
    upperText    []string
    lowerText    []string
}

func (i *BasicTooltip) Draw(screen *ebiten.Image) {
    i.gridRenderer.DrawFilledBorder(screen, i.topLeft, i.bottomRight, "")
    i.drawStringArray(screen, i.upperText, i.topLeft.Y+2)

    if len(i.lowerText) == 0 {
        return
    }
    firstLineAfterName := i.topLeft.Y + 3 + len(i.upperText)
    i.drawStringArray(screen, i.lowerText, firstLineAfterName)
}

func (i *BasicTooltip) IsNull() bool {
    return false
}

func (i *BasicTooltip) drawStringArray(screen *ebiten.Image, lines []string, startAtY int) {
    startPos := geometry.Point{X: i.topLeft.X + 2, Y: startAtY}
    for y, line := range lines {
        i.gridRenderer.DrawColoredString(screen, startPos.X, startPos.Y+y, line, color.White)
    }
}

func (i *BasicTooltip) SetUpperText(name []string) {
    i.upperText = name
}

func (i *BasicTooltip) SetLowerText(infoLines []string) {
    i.lowerText = infoLines
}

func NewTextTooltip(gridRenderer *renderer.DualGridRenderer, text []string, mousePos geometry.Point) *BasicTooltip {
    widthNeeded := util.MaxLen(text)
    heightNeeded := len(text)

    width := widthNeeded + 4
    height := heightNeeded + 4

    tooltip := newTooltip(gridRenderer, mousePos, height, width)
    tooltip.SetUpperText(text)
    return tooltip
}

func NewItemTooltip(gridRenderer *renderer.DualGridRenderer, item game.Item, mousePos geometry.Point) *BasicTooltip {
    // let's start with a simple tooltip that just shows the item name
    // and some stats on the second line.
    // so height would be, 2x border, 2x spacing, 5x lines = 9
    // width would be, 2x border, 2x spacing + max(len(name), len(stats))
    infoLines := item.GetTooltipLines()
    screenSize := gridRenderer.GetSmallGridScreenSize()
    widthForContents := int(float64(screenSize.X) * 0.4)
    otherStatsNeededLines := 0

    if len(infoLines) > 0 {
        widthForContents = len(infoLines[0])
        otherStatsNeededLines = len(infoLines) + 1
    }

    wrappedName := util.AutoLayout(item.Name(), widthForContents)

    width := widthForContents + 4
    height := 4 + len(wrappedName) + otherStatsNeededLines

    // can we position the tooltip to the top right of the mouse cursor?
    tooltip := newTooltip(gridRenderer, mousePos, height, width)
    tooltip.SetUpperText(wrappedName)
    tooltip.SetLowerText(infoLines)
    return tooltip
}

func newTooltip(gridRenderer *renderer.DualGridRenderer, mousePos geometry.Point, height int, width int) *BasicTooltip {
    var topLeftX, topLeftY int
    // decide if above or below the mouse cursor
    screenSize := gridRenderer.GetSmallGridScreenSize()
    if mousePos.Y < height {
        // above
        topLeftY = min(mousePos.Y+1, screenSize.Y-height)
    } else {
        // below
        topLeftY = max(mousePos.Y-height, 0)
    }
    // decide if left or right of the mouse cursor
    if mousePos.X < width {
        // left
        topLeftX = min(mousePos.X+1, screenSize.X-width)
    } else {
        // right
        topLeftX = max(mousePos.X-width, 0)
    }

    bottomRightX := topLeftX + width
    bottomRightY := topLeftY + height

    return &BasicTooltip{
        gridRenderer: gridRenderer,
        topLeft:      geometry.Point{X: topLeftX, Y: topLeftY},
        bottomRight:  geometry.Point{X: bottomRightX, Y: bottomRightY},
    }
}
