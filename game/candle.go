package game

import (
    "Legacy/renderer"
    "image/color"
)

type Candle struct {
    BaseItem
    isLit bool
}

func (b *Candle) Encode() string {
    return "candle()"
}

func (b *Candle) Name() string {
    if b.isLit {
        return "lit candle"
    } else {
        return "candle"
    }
}
func (b *Candle) CanStackWith(other Item) bool {
    if otherCandle, ok := other.(*Candle); ok {
        return b.isLit == otherCandle.isLit
    } else {
        return false
    }
}

func (b *Candle) TintColor() color.Color {
    return color.White
}

func (b *Candle) GetContextActions(engine Engine) []renderer.MenuItem {
    actions := inventoryItemActions(b, engine)
    if b.isLit {
        actions = append(actions, renderer.MenuItem{
            Text:   "Extinguish",
            Action: func() { b.isLit = false },
        })
    } else {
        actions = append(actions, renderer.MenuItem{
            Text:   "Light",
            Action: func() { b.isLit = true },
        })
    }
    return actions
}

func (b *Candle) Icon(uint64) int32 {
    if b.isLit {
        return 190
    } else {
        return 189
    }
}
func NewCandle(isLit bool) *Candle {
    return &Candle{
        BaseItem: BaseItem{
            name: "candle",
        },
        isLit: isLit,
    }
}
