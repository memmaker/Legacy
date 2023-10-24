package game

import "Legacy/geometry"

type Item interface {
    Pos() geometry.Point
    Icon() int
    SetPos(geometry.Point)
    ShortDescription() string
    Use()
}
