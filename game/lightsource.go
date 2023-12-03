package game

import (
    "Legacy/util"
    "fmt"
    "image/color"
    "math/rand"
)

type LightSource struct {
    BaseItem
    isLit             bool
    unlitIcon         int32
    litIcon           int32
    isAttached        bool
    litAttachedFrames []int32
    unlitAttachedIcon int32
    randomOffset      uint64
}

func (b *LightSource) GetTooltipLines() []string {
    return []string{}
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

func (b *LightSource) GetContextActions(engine Engine) []util.MenuItem {
    actions := inventoryItemActions(b, engine)
    if b.isLit {
        actions = append(actions, util.MenuItem{
            Text:   "Extinguish",
            Action: func() { b.isLit = false },
        })
    } else {
        actions = append(actions, util.MenuItem{
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
    frameIndex := util.GetLoopingFrameFromTick(tick+b.randomOffset, 0.2, len(b.litAttachedFrames))
    return b.litAttachedFrames[frameIndex]
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
        randomOffset:      uint64(rand.Intn(15)),
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
        randomOffset:      uint64(rand.Intn(15)),
    }
}
