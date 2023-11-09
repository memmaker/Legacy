package ui

import (
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/renderer"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
)

type ItemTooltip struct {
    item         game.Item
    gridRenderer *renderer.DualGridRenderer

    topLeft     geometry.Point
    bottomRight geometry.Point
    nameLines   []string
    infoLines   []string
}

func (i *ItemTooltip) Draw(screen *ebiten.Image) {
    i.gridRenderer.DrawFilledBorder(screen, i.topLeft, i.bottomRight, "")
    i.drawStringArray(screen, i.nameLines, i.topLeft.Y+2)

    if len(i.infoLines) == 0 {
        return
    }
    firstLineAfterName := i.topLeft.Y + 3 + len(i.nameLines)
    i.drawStringArray(screen, i.infoLines, firstLineAfterName)
}

func (i *ItemTooltip) IsNull() bool {
    return false
}

func (i *ItemTooltip) drawStringArray(screen *ebiten.Image, lines []string, startAtY int) {
    startPos := geometry.Point{X: i.topLeft.X + 2, Y: startAtY}
    for y, line := range lines {
        i.gridRenderer.DrawColoredString(screen, startPos.X, startPos.Y+y, line, color.White)
    }
}

func NewItemTooltip(gridRenderer *renderer.DualGridRenderer, item game.Item, mousePos geometry.Point) *ItemTooltip {
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
    width := widthForContents + 4
    wrappedName := renderer.AutoLayout(item.Name(), widthForContents)

    height := 4 + len(wrappedName) + otherStatsNeededLines

    // can we position the tooltip to the top right of the mouse cursor?
    var topLeftX, topLeftY int
    // decide if above or below the mouse cursor
    if mousePos.Y < height {
        // above
        topLeftY = mousePos.Y + 1
    } else {
        // below
        topLeftY = mousePos.Y - height
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

    return &ItemTooltip{
        item:         item,
        gridRenderer: gridRenderer,
        topLeft:      geometry.Point{X: topLeftX, Y: topLeftY},
        bottomRight:  geometry.Point{X: bottomRightX, Y: bottomRightY},
        nameLines:    wrappedName,
        infoLines:    infoLines,
    }
}
