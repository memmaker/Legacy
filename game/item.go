package game

import (
    "Legacy/geometry"
    "Legacy/renderer"
    "image/color"
    "regexp"
)

type ItemHolder interface {
    Pos() geometry.Point
    RemoveItem(item Item)
}
type ItemWearer interface {
    Pos() geometry.Point
    Unequip(item Item)
    Name() string
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
    SetHidden(hidden bool, discoveryMessage []string)
    Discover() []string
    CanStackWith(other Item) bool
    IsHeld() bool
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

func (i *BaseItem) TintColor() color.Color {
    return color.White
}

func (i *BaseItem) IsHeld() bool {
    return i.holder != nil
}
func inventoryItemActions(item Item, engine Engine) []renderer.MenuItem {
    var actions []renderer.MenuItem
    if item.GetHolder() == nil {
        actions = append(actions, renderer.MenuItem{
            Text: "Take",
            Action: func() {
                if item.GetHolder() == nil {
                    engine.PickUpItem(item)
                }
            },
        })
    } else if engine.IsPlayerControlled(item.GetHolder()) {
        actions = append(actions, renderer.MenuItem{
            Text: "Drop",
            Action: func() {
                if engine.IsPlayerControlled(item.GetHolder()) {
                    engine.DropItem(item)
                }
            },
        })
    }
    return actions
}

func NewItemFromString(encoded string) Item {
    // format:
    // {ItemType}:({Param1},{Param2},...)

    regex := regexp.MustCompile(`([a-zA-Z]+)\(([^)]*)\)`)

    matches := regex.FindStringSubmatch(encoded)
    if len(matches) != 3 {
        return NewKeyFromImportance("unknown item", "unknown item", 1)
    }

    itemType := matches[1]
    params := matches[2]

    switch itemType {
    case "key":
        return NewKeyFromString(params)
    case "potion":
        return NewPotion()
    case "candle":
        return NewCandle(false)
    case "scroll":
        return NewScrollFromString(params)
    case "armor":
        return NewArmorFromString(params)
    case "noitem":
        return NewPseudoItemFromString(params)
    }
    return NewKeyFromImportance("unknown item", "unknown item", 1)
}
