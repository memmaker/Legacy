package gridmap

import (
    "fmt"
)

type SpecialTileType uint64

func (t SpecialTileType) ToString() string {
    switch t {
    case SpecialTileDefaultFloor:
        return "defaultFloor"
    case SpecialTileToilet:
        return "toilet"
    case SpecialTilePlayerSpawn:
        return "playerSpawn"
    case SpecialTilePlayerExit:
        return "playerExit"
    case SpecialTileTreeLike:
        return "treeLike"
    case SpecialTileTypeFood:
        return "food"
    case SpecialTileTypePowerOutlet:
        return "powerOutlet"
    case SpecialTileLethal:
        return "lethal"
    default:
        return "none"
    }
}

func NewSpecialTileTypeFromString(text string) SpecialTileType {
    switch text {
    case "defaultFloor":
        return SpecialTileDefaultFloor
    case "toilet":
        return SpecialTileToilet
    case "playerSpawn":
        return SpecialTilePlayerSpawn
    case "playerExit":
        return SpecialTilePlayerExit
    case "treeLike":
        return SpecialTileTreeLike
    case "food":
        return SpecialTileTypeFood
    case "powerOutlet":
        return SpecialTileTypePowerOutlet
    case "lethal":
        return SpecialTileLethal
    default:
        return SpecialTileNone
    }
}

// These are markers, so we can identify special types of tiles programmatically.
const (
    SpecialTileNone SpecialTileType = iota
    SpecialTileDefaultFloor
    SpecialTileToilet
    SpecialTilePlayerSpawn
    SpecialTilePlayerExit
    SpecialTileTreeLike
    SpecialTileTypeFood
    SpecialTileTypePowerOutlet
    SpecialTileLethal
)

type Tile struct {
    DefinedIcon        int
    DefinedDescription string
    IsWalkable         bool
    IsTransparent      bool
    Special            SpecialTileType
}

func (t Tile) Icon() int {
    return t.DefinedIcon
}

func (t Tile) Description() string {
    return t.DefinedDescription
}

func (t Tile) EncodeAsString() string {
    return fmt.Sprintf("%c: %s", t.DefinedIcon, t.DefinedDescription)
}

func (t Tile) IsLethal() bool {
    return t.Special == SpecialTileLethal
}

func (t Tile) ToString() string {
    if t.Special != SpecialTileNone {
        return fmt.Sprintf("%s (%s)", t.DefinedDescription, t.Special.ToString())
    }
    return t.DefinedDescription
}

type MapCell[ActorType interface {
    comparable
    MapActor
}, ItemType interface {
    comparable
    MapObject
}, ObjectType interface {
    comparable
    MapObjectWithProperties[ActorType]
}] struct {
    TileType    Tile
    IsExplored  bool
    Actor       *ActorType
    DownedActor *ActorType
    Item        *ItemType
    Object      *ObjectType
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithItemHereRemoved(itemHere ItemType) MapCell[ActorType, ItemType, ObjectType] {
    if c.Item != nil && *c.Item == itemHere {
        c.Item = nil
    }
    return c
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithObjectHereRemoved(obj ObjectType) MapCell[ActorType, ItemType, ObjectType] {
    if c.Object != nil && *c.Object == obj {
        c.Object = nil
    }
    return c
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithItemRemoved() MapCell[ActorType, ItemType, ObjectType] {
    c.Item = nil
    return c
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithObjectRemoved() MapCell[ActorType, ItemType, ObjectType] {
    c.Object = nil
    return c
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithDownedActor(a ActorType) MapCell[ActorType, ItemType, ObjectType] {
    c.DownedActor = &a
    return c
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithActor(actor ActorType) MapCell[ActorType, ItemType, ObjectType] {
    c.Actor = &actor
    return c
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithObject(obj ObjectType) MapCell[ActorType, ItemType, ObjectType] {
    c.Object = &obj
    return c
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithActorHereRemoved(actorHere ActorType) MapCell[ActorType, ItemType, ObjectType] {
    if c.Actor != nil && *c.Actor == actorHere {
        c.Actor = nil
    }
    return c
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithItem(item ItemType) MapCell[ActorType, ItemType, ObjectType] {
    c.Item = &item
    return c
}
