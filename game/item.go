package game

import (
    "Legacy/geometry"
    "Legacy/renderer"
    "fmt"
    "image/color"
)

type ItemHolder interface {
    Pos() geometry.Point
}
type Item interface {
    Pos() geometry.Point
    Icon(uint64) int
    TintColor() color.Color
    SetPos(geometry.Point)
    Name() string
    GetContextActions(engine Engine) []renderer.MenuItem
    SetHolder(owner ItemHolder)
    GetHolder() ItemHolder
    IsHidden() bool
    SetHidden(hidden bool)
    CanStackWith(other Item) bool
}

type BaseItem struct {
    GameObject
    holder ItemHolder
    name   string
}

func (i *BaseItem) Name() string {
    return i.name
}

func (i *BaseItem) SetHolder(holder ItemHolder) {
    i.holder = holder
}

func (i *BaseItem) GetHolder() ItemHolder {
    return i.holder
}

func (i *BaseItem) IsHeld() bool {
    return i.holder != nil
}
func inventoryItemActions(item Item, engine Engine) []renderer.MenuItem {
    var actions []renderer.MenuItem
    if item.GetHolder() == nil {
        actions = append(actions, renderer.MenuItem{
            Text: fmt.Sprintf("Take \"%s\"", item.Name()),
            Action: func() {
                if item.GetHolder() == nil {
                    engine.PickUpItem(item)
                }
            },
        })
    } else if engine.IsPlayerControlled(item.GetHolder()) {
        actions = append(actions, renderer.MenuItem{
            Text: fmt.Sprintf("Drop \"%s\"", item.Name()),
            Action: func() {
                if engine.IsPlayerControlled(item.GetHolder()) {
                    engine.DropItem(item)
                }
            },
        })
    }
    return actions
}
