package game

import (
    "Legacy/renderer"
    "fmt"
    "image/color"
)

type LightSource struct {
    BaseItem
    isLit             bool
    unlitIcon         int32
    litIcon           int32
    isAttached        bool
    tickCount         int
    litAttachedFrames []int32
    unlitAttachedIcon int32
}

func (b *LightSource) InventoryIcon() int32 {
    return 174
}

func (b *LightSource) SetHolder(holder ItemHolder) {
    b.BaseItem.SetHolder(holder)
    b.isAttached = false
}

func (b *LightSource) Encode() string {
    return "candle()"
}

func (b *LightSource) Name() string {
    if b.isLit {
        return fmt.Sprintf("a lit %s", b.name)
    } else {
        return fmt.Sprintf("a %s", b.name)
    }
}
func (b *LightSource) CanStackWith(other Item) bool {
    if otherCandle, ok := other.(*LightSource); ok {
        return b.isLit == otherCandle.isLit && b.name == otherCandle.name && b.isAttached == otherCandle.isAttached && b.holder == otherCandle.holder
    } else {
        return false
    }
}

func (b *LightSource) TintColor() color.Color {
    return color.White
}

func (b *LightSource) GetContextActions(engine Engine) []renderer.MenuItem {
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

func (b *LightSource) Icon(tick uint64) int32 {
    if b.isAttached {
        return b.attachedIcon(tick)
    }
    if b.isLit {
        return b.litIcon
    } else {
        return b.unlitIcon
    }
}

func (b *LightSource) attachedIcon(tick uint64) int32 {
    if !b.isLit {
        return b.unlitAttachedIcon
    }
    b.tickCount++
    if b.tickCount > 30 {
        b.tickCount = 0
    }

    if b.tickCount < 15 {
        return b.litAttachedFrames[0]
    } else {
        return b.litAttachedFrames[1]
    }
}
func NewCandle(isLit bool) *LightSource {
    return &LightSource{
        BaseItem: BaseItem{
            name: "candle",
        },
        unlitIcon: 189,
        litIcon:   190,
        isLit:     isLit,
    }
}

func NewLeftTorch(isLit bool) *LightSource {
    return &LightSource{
        BaseItem: BaseItem{
            name: "torch",
        },
        isLit:             isLit,
        litAttachedFrames: []int32{209, 210},
        unlitAttachedIcon: 217,
        unlitIcon:         214,
        litIcon:           213,
        isAttached:        true,
    }
}

func NewRightTorch(isLit bool) *LightSource {
    return &LightSource{
        BaseItem: BaseItem{
            name: "torch",
        },
        isLit:             isLit,
        litAttachedFrames: []int32{211, 212},
        unlitAttachedIcon: 218,
        unlitIcon:         214,
        litIcon:           213,
        isAttached:        true,
    }
}
