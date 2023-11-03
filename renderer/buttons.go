package renderer

import (
    "Legacy/geometry"
    "github.com/hajimehoshi/ebiten/v2"
)

type OneCellButton struct {
    icon   int32
    action func()
}

type ButtonHolder struct {
    buttons map[geometry.Point]OneCellButton
}

func NewButtonHolder() ButtonHolder {
    return ButtonHolder{
        buttons: make(map[geometry.Point]OneCellButton),
    }
}

func (b *ButtonHolder) AddButton(pos geometry.Point, icon int32, callback func()) {
    b.buttons[pos] = OneCellButton{
        icon:   icon,
        action: callback,
    }
}

func (b *ButtonHolder) Draw(gridRenderer *DualGridRenderer, screen *ebiten.Image) {
    for pos, button := range b.buttons {
        gridRenderer.DrawOnSmallGrid(screen, pos.X, pos.Y, button.icon)
    }
}

func (b *ButtonHolder) OnMouseClicked(x int, y int) bool {
    clickPos := geometry.Point{X: x, Y: y}
    if button, ok := b.buttons[clickPos]; ok {
        button.action()
        return true
    }
    return false
}
