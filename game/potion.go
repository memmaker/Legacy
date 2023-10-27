package game

import (
    "Legacy/renderer"
    "fmt"
    "image/color"
)

type Potion struct {
    BaseItem
    isEmpty bool
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
        Text: fmt.Sprintf("Quaff \"%s\"", b.name),
        Action: func() {
            if !b.isEmpty {
                engine.ShowDrinkPotionMenu(b)
            }
        },
    })
    return actions
}

func (b *Potion) Icon(uint64) int {
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
            name: "magic potion",
        },
    }
}
