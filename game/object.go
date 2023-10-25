package game

import (
    "Legacy/geometry"
    "Legacy/renderer"
    "fmt"
    "image/color"
)

type BaseObject struct {
    GameObject
    icon        int
    name        string
    description []string
}

func (a *BaseObject) Icon() int {
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
            Text: fmt.Sprintf("Look at \"%s\"", implObject.Name()),
            Action: func() {
                engine.ShowScrollableText(implObject.Description(), color.White)
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
    Icon() int
    TintColor() color.Color
    SetPos(geometry.Point)
    Name() string
    GetContextActions(engine Engine) []renderer.MenuItem
    IsWalkable(person *Actor) bool
    IsTransparent() bool
    IsPassableForProjectile() bool
    Description() []string
}
