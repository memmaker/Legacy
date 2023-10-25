package game

import (
    "Legacy/renderer"
    "image/color"
)

type Key struct {
    BaseItem
    icon     int
    key      string
    keyColor color.Color
}

func (k *Key) TintColor() color.Color {
    return k.keyColor
}

func (k *Key) GetContextActions(engine Engine) []renderer.MenuItem {
    return inventoryItemActions(k, engine)
}

func (k *Key) Icon() int {
    return k.icon
}
func NewKey(name, key string, keyColor color.Color) *Key {
    return &Key{
        BaseItem: BaseItem{
            name: name,
        },
        icon:     182,
        keyColor: keyColor,
        key:      key,
    }
}
