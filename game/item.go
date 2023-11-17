package game

import (
    "Legacy/geometry"
    "Legacy/recfile"
    "Legacy/util"
    "fmt"
    "image/color"
)

type ItemHolder interface {
    Pos() geometry.Point
    RemoveItem(item Item) bool
    Name() string
}
type ItemWearer interface {
    Pos() geometry.Point
    Unequip(item Item)
    Name() string
}
type Item interface {
    Pos() geometry.Point
    Icon(uint64) int32
    InventoryIcon() int32
    TintColor() color.Color
    SetPos(geometry.Point)
    Name() string
    GetContextActions(engine Engine) []util.MenuItem
    SetHolder(owner ItemHolder)
    GetHolder() ItemHolder
    IsHidden() bool
    Discover() []string
    CanStackWith(other Item) bool
    IsHeld() bool
    GetValue() int
    Encode() string
    SetName(value string)
    SetHidden(asBool bool)
    SetValue(asInt int)
    SetPickupEvent(name string)
    GetPickupEvent() string
    GetTooltipLines() []string
}

type BaseItem struct {
    GameObject
    holder          ItemHolder
    name            string
    baseValue       int
    pickupEventName string
}

func (i *BaseItem) SetPickupEvent(name string) {
    i.pickupEventName = name
}
func (i *BaseItem) GetPickupEvent() string {
    return i.pickupEventName
}
func (i *BaseItem) SetName(value string) {
    i.name = value
}

func (i *BaseItem) Name() string {
    return i.name
}

func (i *BaseItem) SetValue(value int) {
    i.baseValue = value
}

func (i *BaseItem) GetValue() int {
    return i.baseValue
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
func inventoryItemActions(item Item, engine Engine) []util.MenuItem {
    var actions []util.MenuItem
    if item.GetHolder() == nil {
        actions = append(actions, util.MenuItem{
            Text: "Take",
            Action: func() {
                if item.GetHolder() == nil {
                    engine.PickUpItem(item)
                }
            },
        })
    } else if engine.IsPlayerControlled(item.GetHolder()) {
        actions = append(actions, util.MenuItem{
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

func ItemToRecord(item Item) recfile.Record {
    return recfile.Record{
        recfile.Field{Name: "item", Value: item.Encode()},
        recfile.Field{Name: "name", Value: item.Name()},
        recfile.Field{Name: "position", Value: item.Pos().Encode()},
        recfile.Field{Name: "isHidden", Value: recfile.BoolStr(item.IsHidden())},
        recfile.Field{Name: "value", Value: recfile.IntStr(item.GetValue())},
    }
}
func NewItemFromRecord(record recfile.Record) Item {
    theItem := NewItemFromString(record[0].Value)
    for _, field := range record {
        switch field.Name {
        case "name":
            theItem.SetName(field.Value)
        case "position":
            theItem.SetPos(geometry.MustDecodePoint(field.Value))
        case "isHidden":
            theItem.SetHidden(field.AsBool())
        case "value":
            theItem.SetValue(field.AsInt())
        }
    }
    return theItem
}
func NewItemFromString(encoded string) Item {
    predicate := recfile.StrPredicate(encoded)
    if predicate == nil {
        println(fmt.Sprintf("Unknown item type and params: %s", encoded))
        return NewKeyFromImportance("unknown item", "unknown item", 1)
    }

    switch predicate.Name() {
    case "key":
        return NewKeyFromPredicate(predicate)
    case "potion":
        return NewPotion()
    case "candle":
        return NewCandle(false)
    case "scroll":
        scroll := NewScrollFromPredicate(predicate)
        scroll.SetAutoLayoutText(true)
        return scroll
    case "fixedScroll":
        scroll := NewScrollFromPredicate(predicate)
        scroll.SetAutoLayoutText(false)
        return scroll
    case "armor":
        return NewArmorFromPredicate(predicate)
    case "noitem":
        return NewPseudoItemFromPredicate(predicate)
    case "flavor": // example flavor(a teddy, 20, a nice cozy teddy bear)
        return NewFlavorItemFromPredicate(predicate)
    case "weapon":
        return NewWeaponFromPredicate(predicate)
    case "namedWeapon":
        return NewNamedWeapon(predicate.GetString(0))
    case "tool":
        return NewToolFromPredicate(predicate)
    }
    println(fmt.Sprintf("Unknown item type and params: %s", encoded))
    return NewKeyFromImportance("unknown item", "unknown item", 1)
}
