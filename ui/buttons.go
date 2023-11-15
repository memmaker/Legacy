package ui

import (
    "Legacy/ega"
    "Legacy/geometry"
    "Legacy/renderer"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
)

type NoTooltip struct{}

func (n NoTooltip) IsNull() bool {
    return true
}

func (n NoTooltip) Draw(screen *ebiten.Image) {}

type Tooltip interface {
    Draw(screen *ebiten.Image)
    IsNull() bool
}

type IconOnlyButton struct {
    icon   int32
    action func()
}

type IconAndTextButton struct {
    rect      geometry.Rect
    icon      int32
    text      string
    textColor color.Color
    action    func()
}

func NewIconAndTextButton(rect geometry.Rect, callback func()) *IconAndTextButton {
    return &IconAndTextButton{
        rect:      rect,
        action:    callback,
        textColor: color.White,
    }
}
func (i *IconAndTextButton) SetText(text string) {
    i.text = text
    i.rect.Max = i.rect.Min.Add(geometry.Point{X: len(text) + 1, Y: 1})
}

func (i *IconAndTextButton) SetIcon(icon int32) {
    i.icon = icon
}
func (i *IconAndTextButton) Draw(screen *ebiten.Image, gridRenderer *renderer.DualGridRenderer) {
    gridRenderer.DrawOnSmallGrid(screen, i.rect.Min.X, i.rect.Min.Y, i.icon)
    gridRenderer.DrawColoredString(screen, i.rect.Min.X+1, i.rect.Min.Y, i.text, i.textColor)
}

func (i *IconAndTextButton) DrawHover(screen *ebiten.Image, renderer *renderer.DualGridRenderer) {
    renderer.DrawOnSmallGrid(screen, i.rect.Min.X, i.rect.Min.Y, i.icon)
    renderer.DrawColoredString(screen, i.rect.Min.X+1, i.rect.Min.Y, i.text, ega.BrightGreen)
}

func (i *IconAndTextButton) DrawSelected(screen *ebiten.Image, renderer *renderer.DualGridRenderer) {
    renderer.DrawOnSmallGrid(screen, i.rect.Min.X, i.rect.Min.Y, i.icon)
    renderer.DrawColoredString(screen, i.rect.Min.X+1, i.rect.Min.Y, i.text, ega.BrightYellow)
}

type ButtonHolder struct {
    oneCellButtons map[geometry.Point]IconOnlyButton
    rectButtons    []*IconAndTextButton
    hoveredButton  *IconAndTextButton
    selectedButton *IconAndTextButton
}

func NewButtonHolder() ButtonHolder {
    return ButtonHolder{
        oneCellButtons: make(map[geometry.Point]IconOnlyButton),
    }
}

func (b *ButtonHolder) AddIconButton(pos geometry.Point, icon int32, callback func()) {
    b.oneCellButtons[pos] = IconOnlyButton{
        icon:   icon,
        action: callback,
    }
}

func (b *ButtonHolder) AddIconAndTextButton(rect geometry.Rect, callback func()) *IconAndTextButton {
    rectButton := NewIconAndTextButton(rect, callback)
    b.rectButtons = append(b.rectButtons, rectButton)
    return rectButton
}

func (b *ButtonHolder) SetSelectedButton(button *IconAndTextButton) {
    b.selectedButton = button
}
func (b *ButtonHolder) Draw(gridRenderer *renderer.DualGridRenderer, screen *ebiten.Image) {
    for pos, button := range b.oneCellButtons {
        gridRenderer.DrawOnSmallGrid(screen, pos.X, pos.Y, button.icon)
    }
    for _, button := range b.rectButtons {
        if button == b.selectedButton {
            button.DrawSelected(screen, gridRenderer)
        } else if button == b.hoveredButton {
            button.DrawHover(screen, gridRenderer)
        } else {
            button.Draw(screen, gridRenderer)
        }

    }
}

func (b *ButtonHolder) OnMouseClicked(x int, y int) bool {
    clickPos := geometry.Point{X: x, Y: y}
    if button, ok := b.oneCellButtons[clickPos]; ok {
        button.action()
        return true
    }
    for _, button := range b.rectButtons {
        if button.rect.Contains(clickPos) {
            button.action()
            b.selectedButton = button
            return true
        }
    }
    return false
}

func (b *ButtonHolder) OnMouseMoved(x int, y int) (bool, Tooltip) {
    b.hoveredButton = nil
    mousePos := geometry.Point{X: x, Y: y}
    for _, button := range b.rectButtons {
        if button.rect.Contains(mousePos) {
            b.hoveredButton = button
            return true, NoTooltip{}
        }
    }
    return false, NoTooltip{}
}
