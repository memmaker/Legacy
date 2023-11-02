package game

import (
    "Legacy/geometry"
    "Legacy/recfile"
    "Legacy/renderer"
    "image/color"
)

type Loot string

const (
    LootCommon    Loot = "common"
    LootHealer    Loot = "healer"
    LootFood      Loot = "food"
    LootPotions   Loot = "potions"
    LootGold      Loot = "gold"
    LootWeapon    Loot = "weapon"
    LootArmor     Loot = "armor"
    LootScrolls   Loot = "scrolls"
    LootLockpicks Loot = "lockpicks"
)

type Chest struct {
    BaseObject
    needsKey       string
    lootLevel      int
    lootType       []Loot
    items          []Item
    hasCreatedLoot bool
}

func (s *Chest) GetItems() []Item {
    return s.items
}
func (s *Chest) RemoveItem(item Item) {
    for i, v := range s.items {
        if v == item {
            s.items = append(s.items[:i], s.items[i+1:]...)
            return
        }
    }
}

func (s *Chest) AddItem(item Item) {
    s.items = append(s.items, item)
}

func (s *Chest) TintColor() color.Color {
    return color.White
}

func (s *Chest) Name() string {
    if s.needsKey != "" {
        return "a locked chest"
    } else if s.IsEmpty() {
        return "a chest (empty)"
    } else {
        return "a chest"
    }
}

func NewChest(lootLevel int, lootType []Loot) *Chest {
    return &Chest{
        BaseObject: BaseObject{
            icon: 25,
        },
        lootLevel: lootLevel,
        lootType:  lootType,
    }
}
func NewFixedChest(contents []Item) *Chest {
    return &Chest{
        BaseObject: BaseObject{
            icon: 25,
        },
        items:          contents,
        hasCreatedLoot: true,
    }
}

func NewChestFromRecord(record recfile.Record) *Chest {
    chest := NewChest(0, []Loot{})
    for _, field := range record {
        switch field.Name {
        case "name":
            chest.name = field.Value
        case "isHidden":
            chest.isHidden = field.AsBool()
        case "icon":
            chest.icon = field.AsInt32()
        case "position":
            chest.SetPos(geometry.MustDecodePoint(field.Value))
        case "needsKey":
            chest.needsKey = field.Value
        case "lootLevel":
            chest.lootLevel = field.AsInt()
        case "hasCreatedLoot":
            chest.hasCreatedLoot = field.AsBool()
        case "lootType":
            chest.lootType = append(chest.lootType, Loot(field.Value))
        case "item":
            chest.items = append(chest.items, NewItemFromString(field.Value))
        }
    }
    return chest
}
func (s *Chest) ToRecordAndType() (recfile.Record, string) {
    var record recfile.Record
    record = recfile.Record{
        {Name: "name", Value: s.name},
        {Name: "isHidden", Value: recfile.BoolStr(s.isHidden)},
        {Name: "icon", Value: recfile.Int32Str(s.icon)},
        {Name: "position", Value: s.Pos().Encode()},
        {"needsKey", s.needsKey},
        {"lootLevel", recfile.IntStr(s.lootLevel)},
        {"hasCreatedLoot", recfile.BoolStr(s.hasCreatedLoot)},
    }
    for _, lootType := range s.lootType {
        record = append(record, recfile.Field{Name: "lootType", Value: string(lootType)})
    }
    for _, item := range s.items {
        record = append(record, recfile.Field{Name: "item", Value: item.Encode()})
    }

    return record, "chest"
}

func (s *Chest) Description() []string {
    if s.needsKey != "" {
        return []string{
            "You see a chest.",
            "It appears to be locked.",
        }
    } else {
        return []string{
            "You see a chest.",
            "It is unlocked.",
        }
    }
}

func (s *Chest) Icon(uint64) int32 {
    if s.IsEmpty() {
        return 204
    }
    return s.icon
}
func (s *Chest) IsWalkable(person *Actor) bool {
    return true
}

func (s *Chest) IsTransparent() bool {
    return true
}

func (s *Chest) spawnLoot(engine Engine) {
    if s.hasCreatedLoot {
        return
    }
    s.items = engine.CreateLootForContainer(s.lootLevel, s.lootType)
    s.hasCreatedLoot = true
}
func (s *Chest) GetContextActions(engine Engine) []renderer.MenuItem {
    actions := s.BaseObject.GetContextActions(engine, s)
    var additionalActions []renderer.MenuItem
    if s.needsKey == "" || engine.PartyHasKey(s.needsKey) {
        additionalActions = append(additionalActions, renderer.MenuItem{
            Text: "Open",
            Action: func() {
                s.spawnLoot(engine)
                engine.ShowContainer(s)
            },
        })
    }
    return append(additionalActions, actions...)
}

func (s *Chest) SetLockedWithKey(key string) {
    s.needsKey = key
}

func (s *Chest) SetFixedLoot(loot []Item) {
    s.items = loot
    s.hasCreatedLoot = true
}

func (s *Chest) IsEmpty() bool {
    if s.hasCreatedLoot {
        return len(s.items) == 0
    }
    return false
}
