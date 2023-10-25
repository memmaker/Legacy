package game

import "Legacy/geometry"

type GameObject struct {
    pos geometry.Point
}

func (a *GameObject) Pos() geometry.Point {
    return a.pos
}

func (a *GameObject) SetPos(pos geometry.Point) {
    a.pos = pos
}
