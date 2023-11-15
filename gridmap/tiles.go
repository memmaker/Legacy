package gridmap

import (
    "encoding/binary"
    "fmt"
    "io"
)

type SpecialTileType uint64

// These are markers, so we can identify special types of tiles programmatically.
const (
    SpecialTileNone SpecialTileType = iota
    SpecialTileLethal
    SpecialTileBreakable
    SpecialTileBreakableGold
    SpecialTileBreakableGems
    SpecialTileBreakableGlass
    SpecialTileTrap
    SpecialTileForest
    SpecialTileMountain
    SpecialTileWater
    SpecialTileSwamp
    SpecialTileBed
)

type Tile struct {
    DefinedIcon        int32 // we need this
    DefinedDescription string
    IsWalkable         bool            // this
    IsTransparent      bool            // this
    Special            SpecialTileType // and this
}

func (t Tile) IsBreakable() bool {
    return t.Special == SpecialTileBreakable || t.Special == SpecialTileBreakableGold || t.Special == SpecialTileBreakableGems || t.Special == SpecialTileBreakableGlass
}

func (t Tile) IsBed() bool {
    return t.Special == SpecialTileBed
}

func (t Tile) ToBinary(out io.Writer) {
    // we want to serialize the tile
    // icon, iswalkable, istransparent, special

    must(binary.Write(out, binary.LittleEndian, t.DefinedIcon))
    must(binary.Write(out, binary.LittleEndian, t.IsWalkable))
    must(binary.Write(out, binary.LittleEndian, t.IsTransparent))
    must(binary.Write(out, binary.LittleEndian, t.Special))
}

func NewTileFromBinary(in io.Reader) Tile {
    var icon int32
    var isWalkable bool
    var isTransparent bool
    var special SpecialTileType

    must(binary.Read(in, binary.LittleEndian, &icon))
    must(binary.Read(in, binary.LittleEndian, &isWalkable))
    must(binary.Read(in, binary.LittleEndian, &isTransparent))
    must(binary.Read(in, binary.LittleEndian, &special))

    return Tile{
        DefinedIcon:   icon,
        IsWalkable:    isWalkable,
        IsTransparent: isTransparent,
        Special:       special,
    }
}

func must(err error) {
    if err != nil {
        panic(err)
    }
}
func (t Tile) Icon() int32 {
    return t.DefinedIcon
}

func (t Tile) Description() string {
    return t.DefinedDescription
}

func (t Tile) EncodeAsString() string {
    return fmt.Sprintf("%c: %s", t.DefinedIcon, t.DefinedDescription)
}

func (t Tile) WithIsWalkable(isWalkable bool) Tile {
    t.IsWalkable = isWalkable
    return t
}

func (t Tile) WithIcon(icon int32) Tile {
    t.DefinedIcon = icon
    return t
}

func (t Tile) WithIsTransparent(value bool) Tile {
    t.IsTransparent = value
    return t
}

func (t Tile) WithSpecial(special SpecialTileType) Tile {
    t.Special = special
    return t
}

func (t Tile) GetDebrisTile() int32 {
    switch t.Special {
    case SpecialTileBreakableGlass:
        return 36
    default:
        return 32
    }
}

func (t Tile) IsForest() bool {
    return t.Special == SpecialTileForest
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
func (c MapCell[ActorType, ItemType, ObjectType]) WithDownedActorHereRemoved(actorHere ActorType) MapCell[ActorType, ItemType, ObjectType] {
    if c.DownedActor != nil && *c.DownedActor == actorHere {
        c.DownedActor = nil
    }
    return c
}

func (c MapCell[ActorType, ItemType, ObjectType]) WithItem(item ItemType) MapCell[ActorType, ItemType, ObjectType] {
    c.Item = &item
    return c
}
