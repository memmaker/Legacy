package game

import (
    "Legacy/geometry"
    "Legacy/renderer"
    "image/color"
)

type BaseObject struct {
    GameObject
    icon        int
    name        string
    description []string
}

func (a *BaseObject) Icon(uint64) int {
    return a.icon
}

func (a *BaseObject) IsWalkable(person *Actor) bool {
    return false
}

func (a *BaseObject) IsTransparent() bool {
    return false
}

func (a *BaseObject) IsPassableForProjectile() bool {
    return false
}

func (a *BaseObject) GetContextActions(engine Engine, implObject Object) []renderer.MenuItem {
    return []renderer.MenuItem{
        {
            Text: "Examine",
            Action: func() {
                engine.ShowColoredText(implObject.Description(), color.White, true)
            },
        },
    }
}

func (a *BaseObject) Name() string {
    return a.name
}

func (a *BaseObject) Description() []string {
    return a.description
}

type Object interface {
    Pos() geometry.Point
    Icon(uint64) int
    TintColor() color.Color
    SetPos(geometry.Point)
    Name() string
    GetContextActions(engine Engine) []renderer.MenuItem
    IsWalkable(person *Actor) bool
    IsTransparent() bool
    IsPassableForProjectile() bool
    Description() []string
    IsHidden() bool
    SetHidden(hidden bool, discoveryMessage []string)
    Discover() []string
}
