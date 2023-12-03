package ui

import (
    "Legacy/geometry"
    "Legacy/renderer"
    "Legacy/util"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
)

type ScrollableTextWindow struct {
    ButtonHolder
    ScrollingContent
    topLeft     geometry.Point
    bottomRight geometry.Point

    text []string

    textColor color.Color

    gridRenderer *renderer.DualGridRenderer

    shouldClose bool
    title       string
    rightAction func()
    leftAction  func()
}

func (r *ScrollableTextWindow) OnMouseWheel(x int, y int, dy float64) bool {
    if r.ScrollingContent.OnMouseWheel(x, y, dy) {
        r.OnMouseMoved(x, y)
    }
    return true
}

func (r *ScrollableTextWindow) OnCommand(command CommandType) bool {
    switch command {
    case PlayerCommandCancel:
        r.ActionCancel()
    case PlayerCommandConfirm:
        r.ActionConfirm()
    case PlayerCommandUp:
        r.ActionUp()
    case PlayerCommandDown:
        r.ActionDown()
    case PlayerCommandLeft:
        r.ActionLeft()
    case PlayerCommandRight:
        r.ActionRight()
    }
    return true
}

func (r *ScrollableTextWindow) ActionLeft() {
    if r.leftAction != nil {
        r.leftAction()
        return
    }
    // page up
    if r.needsScroll() {
        pageSize := r.bottomRight.Y - r.topLeft.Y - 4
        r.ScrollUp(pageSize)
    }
}

func (r *ScrollableTextWindow) ActionRight() {
    if r.rightAction != nil {
        r.rightAction()
        return
    }
    // page down
    if r.needsScroll() {
        pageSize := r.bottomRight.Y - r.topLeft.Y - 4
        r.ScrollDown(pageSize)
    }
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

func NewAutoTextWindow(gridRenderer *renderer.DualGridRenderer, text []string) *ScrollableTextWindow {
    screenSize := gridRenderer.GetSmallGridScreenSize()
    borderNeeded := 4 * 2
    widthAvailable := screenSize.X - borderNeeded
    widthWanted := widthAvailable
    text = util.AutoLayoutArray(text, widthWanted)
    addedSpace := geometry.Point{}
    topLeft, bottomRight := gridRenderer.GetAutoFitRectWithExtraSpace(text, addedSpace)
    modal := NewScrollableTextWindow(gridRenderer, topLeft, bottomRight)
    modal.SetText(text)
    return modal
}
func NewFixedTextWindow(gridRenderer *renderer.DualGridRenderer, text []string) *ScrollableTextWindow {
    addedSpace := geometry.Point{}
    topLeft, bottomRight := gridRenderer.GetAutoFitRectWithExtraSpace(text, addedSpace)
    modal := NewScrollableTextWindow(gridRenderer, topLeft, bottomRight)
    modal.SetText(text)
    return modal
}
func NewScrollableTextWindow(gridRenderer *renderer.DualGridRenderer, topLeft, bottomRight geometry.Point) *ScrollableTextWindow {
    s := &ScrollableTextWindow{
        ButtonHolder: NewButtonHolder(),
        gridRenderer: gridRenderer,
        textColor:    color.White,
        topLeft:      topLeft,
        bottomRight:  bottomRight,
    }
    neededHeight := func() int { return len(s.text) }
    availableSpace := func() geometry.Rect { return geometry.NewRect(s.topLeft.X+2, s.topLeft.Y+2, s.bottomRight.X-2, s.bottomRight.Y-2) }
    s.ScrollingContent = NewScrollingContentWithFunctions(neededHeight, availableSpace)
    return s
}

func (r *ScrollableTextWindow) SetText(text []string) {
    r.text = text
}

func (r *ScrollableTextWindow) SetTextColor(color color.Color) {
    r.textColor = color
}

func (r *ScrollableTextWindow) ActionUp() {
    r.ScrollUp(1)
}

func (r *ScrollableTextWindow) ActionDown() {
    r.ScrollDown(1)
}

func (r *ScrollableTextWindow) Draw(screen *ebiten.Image) {
    brForBorder := r.bottomRight
    r.gridRenderer.DrawFilledBorder(screen, r.topLeft, brForBorder, r.title)
    for y := r.topLeft.Y + 2; y < r.bottomRight.Y-2; y++ {
        for x := r.topLeft.X + 2; x <= r.bottomRight.X-2; x++ {
            currentLine := r.text[r.getLineFromScreenLine(y)]
            horizontalIndex := x - r.topLeft.X - 2
            asRunes := []rune(currentLine)
            if horizontalIndex >= len(asRunes) {
                continue
            }
            currentChar := asRunes[horizontalIndex]
            r.gridRenderer.DrawColoredChar(screen, x, y, currentChar, r.textColor)
        }
    }
    r.drawScrollIndicators(r.gridRenderer, screen)
}

func (r *ScrollableTextWindow) SetTitle(name string) {
    r.title = name
}

func (r *ScrollableTextWindow) ActionCancel() {
    r.shouldClose = true
}

func (r *ScrollableTextWindow) SetRightAction(rightAction func()) {
    r.rightAction = rightAction
}

func (r *ScrollableTextWindow) SetLeftAction(leftAction func()) {
    r.leftAction = leftAction
}
