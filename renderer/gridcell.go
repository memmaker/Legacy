package renderer

import (
    "Legacy/geometry"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
)

type CellDrawInfo struct {
    Icon     int32
    Tint     color.Color
    Atlas    *ebiten.Image
    GridPos  geometry.Point
    GridSize int
}
type Drawable interface {
    GetIcon() int32
    GetTint() color.Color
    GetAtlas() *ebiten.Image
}
