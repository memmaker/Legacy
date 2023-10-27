package game

import (
    "Legacy/renderer"
    "fmt"
    "image/color"
)

type Loot string

const (
    Common  Loot = "common"
    Healer  Loot = "healer"
    Food    Loot = "food"
    Weapon  Loot = "weapon"
    Armor   Loot = "armor"
    Scrolls Loot = "scrolls"
)

type Chest struct {
    BaseObject
    needsKey  string
    lootLevel int
    lootType  Loot
    items     []Item
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
    return "chest"
}

func NewChest(lootLevel int, lootType Loot) *Chest {
    return &Chest{
        BaseObject: BaseObject{
            icon: 25,
        },
        lootLevel: lootLevel,
        lootType:  lootType,
    }
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

func (s *Chest) Icon(uint64) int {
    return s.icon
}
func (s *Chest) IsWalkable(person *Actor) bool {
    return true
}

func (s *Chest) IsTransparent() bool {
    return true
}

func (s *Chest) open(engine Engine) {
    s.items = engine.CreateLoot(s.lootLevel, s.lootType)
}
func (s *Chest) GetContextActions(engine Engine) []renderer.MenuItem {
    actions := s.BaseObject.GetContextActions(engine, s)
    if s.needsKey == "" || engine.PartyHasKey(s.needsKey) {
        actions = append(actions, renderer.MenuItem{
            Text:   fmt.Sprintf("Open \"%s\"", s.Name()),
            Action: func() { s.open(engine); engine.ShowContainer(s) },
        })
    }
    return actions
}

func (s *Chest) SetLockedWithKey(key string) {
    s.needsKey = key
}
