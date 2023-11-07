package game

import (
    "Legacy/geometry"
    "Legacy/recfile"
    "Legacy/renderer"
    "image/color"
)

type BaseObject struct {
    GameObject
    icon        int32
    name        string
    description []string
}

func (a *BaseObject) Icon(uint64) int32 {
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
func (a *BaseObject) SetDescription(description []string) {
    a.description = description
}
func (a *BaseObject) GetContextActions(engine Engine, implObject Object) []renderer.MenuItem {
    return []renderer.MenuItem{
        {
            Text: "Examine",
            Action: func() {
                engine.ShowScrollableText(implObject.Description(), color.White, true)
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
    Icon(uint64) int32
    TintColor() color.Color
    SetPos(geometry.Point)
    Name() string
    GetContextActions(engine Engine) []renderer.MenuItem
    IsWalkable(person *Actor) bool
    IsTransparent() bool
    IsPassableForProjectile() bool
    Description() []string
    IsHidden() bool
    Discover() []string
    ToRecordAndType() (recfile.Record, string)
}

func NewObjectFromRecord(record recfile.Record, objectTypeName string) Object {
    switch objectTypeName {
    case "chest":
        return NewChestFromRecord(record)
    case "door":
        return NewDoorFromRecord(record)
    case "shrine":
        return NewShrineFromRecord(record)
    case "fireplace":
        return NewFireplaceFromRecord(record)
    }
    panic("unknown object type")
}
