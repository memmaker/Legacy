package game

import (
    "Legacy/geometry"
    "strconv"
)

type Actor struct {
    pos    geometry.Point
    icon   int
    name   string
    Health int
}

/*
Pos() geometry.Point
    Icon() int
    SetPos(geometry.Point)
*/

func NewActor(name string, pos geometry.Point, icon int) *Actor {
    return &Actor{
        name:   name,
        pos:    pos,
        icon:   icon,
        Health: 10,
    }
}

func (a *Actor) Pos() geometry.Point {
    return a.pos
}

func (a *Actor) Icon() int {
    return a.icon
}

func (a *Actor) SetPos(pos geometry.Point) {
    a.pos = pos
}

func (a *Actor) Name() string {
    return a.name
}

func (a *Actor) GetDetails() []string {
    return []string{
        a.name,
        "Health: " + strconv.Itoa(a.Health),
    }
}

func (a *Actor) LookDescription() []string {
    return []string{
        "A person is standing here.",
    }
}
