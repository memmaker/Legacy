package game

import (
    "Legacy/geometry"
    "Legacy/recfile"
    "Legacy/util"
    "fmt"
    "image/color"
    "math/rand"
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
    emptyIcon      int32
    isLocked       bool
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
    if s.isLocked {
        return fmt.Sprintf("%s (locked)", s.name)
    } else if s.IsEmpty() {
        return fmt.Sprintf("%s (empty)", s.name)
    } else {
        return s.name
    }
}

func NewChest(lootLevel int, lootType []Loot) *Chest {
    return &Chest{
        BaseObject: BaseObject{
            icon: 25,
            name: "a chest",
        },
        emptyIcon: 204,
        lootLevel: lootLevel,
        lootType:  lootType,
    }
}
func NewFixedChest(contents []Item) *Chest {
    return &Chest{
        BaseObject: BaseObject{
            icon: 25,
            name: "a chest",
        },
        emptyIcon:      204,
        items:          contents,
        hasCreatedLoot: true,
    }
}

func NewContainer(lootLevel int, lootType []Loot, name string, icon int32) *Chest {
    return &Chest{
        BaseObject: BaseObject{
            icon: icon,
            name: name,
        },
        emptyIcon: icon,
        lootLevel: lootLevel,
        lootType:  lootType,
    }
}
func NewFixedContainer(contents []Item, name string, icon int32) *Chest {
    return &Chest{
        BaseObject: BaseObject{
            icon: icon,
            name: name,
        },
        emptyIcon:      icon,
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
        case "emptyIcon":
            chest.emptyIcon = field.AsInt32()
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
        {Name: "emptyIcon", Value: recfile.Int32Str(s.emptyIcon)},
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
    var result []string
    result = append(result, fmt.Sprintf("You see %s.", s.Name()))
    if s.isLocked {
        result = append(result, "It appears to be locked.")
    }
    return result
}

func (s *Chest) Icon(uint64) int32 {
    if s.IsEmpty() {
        return s.emptyIcon
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
func (s *Chest) GetContextActions(engine Engine) []util.MenuItem {
    party := engine.GetParty()
    actions := s.BaseObject.GetContextActions(engine, s)
    var additionalActions []util.MenuItem
    if s.needsKey == "" || !s.isLocked {
        additionalActions = append(additionalActions, util.MenuItem{
            Text: "Open",
            Action: func() {
                s.spawnLoot(engine)
                engine.ShowContainer(s)
            },
        })
    }
    if s.isLocked && party.HasKey(s.needsKey) {
        additionalActions = append(additionalActions, util.MenuItem{
            Text: "Unlock",
            Action: func() {
                s.isLocked = false
                party.UsedKey(s.needsKey)
                s.spawnLoot(engine)
                engine.ShowContainer(s)
            },
        })
    }
    if s.isLocked && party.GetLockpicks() > 0 {
        additionalActions = append(additionalActions, util.MenuItem{
            Text: "Pick lock",
            Action: func() {
                if rand.Float64() > (float64(s.lootLevel) / 10.0) {
                    s.isLocked = false
                    s.spawnLoot(engine)
                    engine.ShowContainer(s)
                } else {
                    // broke
                    engine.RemoveLockpick()
                    engine.Print("Your lockpick broke.")
                }
            },
        })
    }
    return append(additionalActions, actions...)
}

func (s *Chest) SetLockedWithKey(key string) {
    s.needsKey = key
    s.isLocked = true
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
