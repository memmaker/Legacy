package renderer

import (
    "Legacy/geometry"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
)

type ScrollableTextWindow struct {
    ButtonHolder
    topLeft     geometry.Point
    bottomRight geometry.Point

    text []string

    textColor color.Color

    scrollOffset int
    gridRenderer *DualGridRenderer

    shouldClose bool
    title       string

    upIndicator, downIndicator int32
    upDownIndicator            int32
}

func (r *ScrollableTextWindow) OnAvatarSwitched() {

}

func (r *ScrollableTextWindow) OnMouseClicked(x int, y int) bool {
    if x < r.topLeft.X || x >= r.bottomRight.X {
        return false
    }
    if y < r.topLeft.Y || y >= r.bottomRight.Y {
        return false
    }
    r.ActionConfirm()
    return true
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
    screenSize := gridRenderer.GetSmallGridScreenSize()
    borderNeeded := 4 * 2
    widthAvailable := screenSize.X - borderNeeded
    text = AutoLayoutArray(text, min(maxLen(text), widthAvailable))
    topLeft, bottomRight := gridRenderer.GetAutoFitRect(text)
    modal := NewScrollableTextWindow(gridRenderer, topLeft, bottomRight)
    modal.SetText(text)
    return modal
}
func NewFixedTextWindow(gridRenderer *DualGridRenderer, text []string) *ScrollableTextWindow {
    topLeft, bottomRight := gridRenderer.GetAutoFitRect(text)
    modal := NewScrollableTextWindow(gridRenderer, topLeft, bottomRight)
    modal.SetText(text)
    return modal
}
func NewScrollableTextWindow(gridRenderer *DualGridRenderer, topLeft, bottomRight geometry.Point) *ScrollableTextWindow {
    return &ScrollableTextWindow{
        ButtonHolder:    NewButtonHolder(),
        gridRenderer:    gridRenderer,
        textColor:       color.White,
        topLeft:         topLeft,
        bottomRight:     bottomRight,
        upDownIndicator: 4,
        downIndicator:   5,
        upIndicator:     6,
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
    if r.scrollOffset > len(r.text)-(r.bottomRight.Y-r.topLeft.Y)+4 {
        r.scrollOffset = len(r.text) - (r.bottomRight.Y - r.topLeft.Y) + 4
    }
}

func (r *ScrollableTextWindow) CanScroll() bool {
    return len(r.text) > (r.bottomRight.Y-r.topLeft.Y)-4
}
func (r *ScrollableTextWindow) Draw(screen *ebiten.Image) {
    brForBorder := r.bottomRight
    if r.CanScroll() {
        brForBorder.X += 1
    }
    r.gridRenderer.DrawFilledBorder(screen, r.topLeft, brForBorder, r.title)
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
    if r.CanScroll() {
        if r.scrollOffset > 0 {
            r.gridRenderer.DrawOnSmallGrid(screen, r.bottomRight.X-1, r.topLeft.Y+2, r.upIndicator)
        }
        if r.scrollOffset < len(r.text)-(r.bottomRight.Y-r.topLeft.Y)+4 {
            r.gridRenderer.DrawOnSmallGrid(screen, r.bottomRight.X-1, r.bottomRight.Y-3, r.downIndicator)
        }
    }
}

func (r *ScrollableTextWindow) SetTitle(name string) {
    r.title = name
}
