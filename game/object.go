package game

import "Legacy/geometry"

type Object struct {
    pos  geometry.Point
    icon int
}

/*
Pos() geometry.Point
    Icon() int
    SetPos(geometry.Point)
*/

func (a *Object) Pos() geometry.Point {
    return a.pos
}

func (a *Object) Icon() int {
    return a.icon
}

func (a *Object) SetPos(pos geometry.Point) {
    a.pos = pos
}

/*
IsWalkable(person ActorType) bool
    IsTransparent() bool
    IsPassableForProjectile() bool
*/

func (a *Object) IsWalkable(person *Actor) bool {
    return false
}

func (a *Object) IsTransparent() bool {
    return false
}

func (a *Object) IsPassableForProjectile() bool {
    return false
}
