package game

import "Legacy/geometry"

type GameObject struct {
    pos      geometry.Point
    isHidden bool
}

func (a *GameObject) Pos() geometry.Point {
    return a.pos
}

func (a *GameObject) SetPos(pos geometry.Point) {
    a.pos = pos
}

func (a *GameObject) IsHidden() bool {
    return a.isHidden
}

func (a *GameObject) SetHidden(hidden bool) {
    a.isHidden = hidden
}
