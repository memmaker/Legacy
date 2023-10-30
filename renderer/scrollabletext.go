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
    title       string
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

func NewAutoTextWindow(gridRenderer *DualGridRenderer, text []string) *ScrollableTextWindow {
    text = AutoLayoutText(text, min(maxLen(text), 34))
    topLeft, bottomRight := gridRenderer.AutoPositionText(text)
    modal := NewScrollableTextWindow(gridRenderer, topLeft, bottomRight)
    modal.SetText(text)
    return modal
}
func NewFixedTextWindow(gridRenderer *DualGridRenderer, text []string) *ScrollableTextWindow {
    topLeft, bottomRight := gridRenderer.AutoPositionText(text)
    modal := NewScrollableTextWindow(gridRenderer, topLeft, bottomRight)
    modal.SetText(text)
    return modal
}
func NewScrollableTextWindow(gridRenderer *DualGridRenderer, topLeft, bottomRight geometry.Point) *ScrollableTextWindow {
    return &ScrollableTextWindow{
        gridRenderer: gridRenderer,
        textColor:    color.White,
        topLeft:      topLeft,
        bottomRight:  bottomRight,
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
    r.gridRenderer.DrawFilledBorder(screen, r.topLeft, r.bottomRight, r.title)
    for y := r.topLeft.Y + 2; y < r.bottomRight.Y-2; y++ {
        for x := r.topLeft.X + 2; x < r.bottomRight.X-2; x++ {
            currentLine := r.text[y-r.topLeft.Y-2+r.scrollOffset]
            horizontalIndex := x - r.topLeft.X - 2
            asRunes := []rune(currentLine)
            if horizontalIndex >= len(asRunes) {
                continue
            }
            currentChar := asRunes[horizontalIndex]
            r.gridRenderer.DrawColoredChar(screen, x, y, currentChar, r.textColor)
        }
    }
}
