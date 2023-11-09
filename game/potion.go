package game

import (
    "Legacy/renderer"
    "image/color"
)

type Potion struct {
    BaseItem
    isEmpty bool
}

func (b *Potion) GetTooltipLines() []string {
    if b.isEmpty {
        return []string{"Empty potion"}
    } else {
        return []string{"Magic potion"}
    }
}

func (b *Potion) InventoryIcon() int32 {
    return 173
}

func (b *Potion) CanStackWith(other Item) bool {
    if _, ok := other.(*Potion); ok {
        return true
    } else {
        return false
    }
}

func (b *Potion) TintColor() color.Color {
    return color.White
}

func (b *Potion) GetContextActions(engine Engine) []renderer.MenuItem {
    actions := inventoryItemActions(b, engine)
    actions = append(actions, renderer.MenuItem{
        Text: "Quaff",
        Action: func() {
            if !b.isEmpty {
                engine.ShowDrinkPotionMenu(b)
            }
        },
    })
    return actions
}

func (b *Potion) Icon(uint64) int32 {
    return 191
}

func (b *Potion) SetEmpty() {
    b.isEmpty = true
}

func (b *Potion) IsEmpty() bool {
    return b.isEmpty
}
func NewPotion() *Potion {
    return &Potion{
        BaseItem: BaseItem{
            name:      "magic potion",
            baseValue: 250,
        },
    }
}

func (b *Potion) Encode() string {
    return "potion()"
}
