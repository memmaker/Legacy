package renderer

import (
    "Legacy/geometry"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
)

type ScrollableTextWindow struct {
    topLeft     geometry.Point
    bottomRight geometry.Point

    text []string

    textColor color.Color

    scrollOffset int
    gridRenderer *DualGridRenderer

    shouldClose bool
}

func (r *ScrollableTextWindow) ShouldClose() bool {
    return r.shouldClose
}

func (r *ScrollableTextWindow) ActionConfirm() {
    r.shouldClose = true
}

func maxLen(text []string) int {
    maxLength := 0
    for _, line := range text {
        if len(line) > maxLength {
            maxLength = len(line)
        }
    }
    return maxLength
}

func NewScrollableTextWindowWithAutomaticSize(gridRenderer *DualGridRenderer, text []string) *ScrollableTextWindow {
    height := min(len(text)+2, 15)
    width := min(maxLen(text)+2, 36)
    // in 8x8 cells
    // center
    screenSize := gridRenderer.GetSmallGridScreenSize()
    topLeft := geometry.Point{X: (screenSize.X - width) / 2, Y: (screenSize.Y - height) / 2}
    bottomRight := geometry.Point{X: topLeft.X + width, Y: topLeft.Y + height}
    modal := NewScrollableTextWindow(gridRenderer, topLeft, bottomRight)
    modal.SetText(text)
    return modal
}
func NewScrollableTextWindow(gridRenderer *DualGridRenderer, topLeft geometry.Point, bottomRight geometry.Point) *ScrollableTextWindow {
    return &ScrollableTextWindow{
        topLeft:      topLeft,
        bottomRight:  bottomRight,
        gridRenderer: gridRenderer,
        textColor:    color.White,
    }
}

func (r *ScrollableTextWindow) SetText(text []string) {
    r.text = text
}

func (r *ScrollableTextWindow) SetTextColor(color color.Color) {
    r.textColor = color
}

func (r *ScrollableTextWindow) ActionUp() {
    r.scrollOffset--
    if r.scrollOffset < 0 {
        r.scrollOffset = 0
    }
}

func (r *ScrollableTextWindow) ActionDown() {
    r.scrollOffset++
    if r.scrollOffset > len(r.text)-(r.bottomRight.Y-r.topLeft.Y)+2 {
        r.scrollOffset = len(r.text) - (r.bottomRight.Y - r.topLeft.Y) + 2
    }
}
func (r *ScrollableTextWindow) Draw(screen *ebiten.Image) {
    r.gridRenderer.DrawFilledBorder(screen, r.topLeft, r.bottomRight)
    for y := r.topLeft.Y + 1; y < r.bottomRight.Y-1; y++ {
        for x := r.topLeft.X + 1; x < r.bottomRight.X-1; x++ {
            currentLine := r.text[y-r.topLeft.Y-1+r.scrollOffset]
            horizontalIndex := x - r.topLeft.X - 1
            asRunes := []rune(currentLine)
            if horizontalIndex >= len(asRunes) {
                continue
            }
            currentChar := asRunes[horizontalIndex]
            r.gridRenderer.DrawColoredChar(screen, x, y, currentChar, r.textColor)
        }
    }
}
