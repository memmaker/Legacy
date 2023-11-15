package game

import (
    "Legacy/ega"
    "Legacy/recfile"
    "Legacy/util"
    "image/color"
    "strconv"
)

type Key struct {
    BaseItem
    icon           int32
    key            string
    importance     int
    FoundInMap     string
    firstOwnerName string
}

func (k *Key) GetTooltipLines() []string {
    return []string{}
}

func (k *Key) InventoryIcon() int32 {
    return 172
}

func (k *Key) CanStackWith(other Item) bool {
    return false
}

func (k *Key) TintColor() color.Color {
    return keyColor(k.importance)
}

func (k *Key) GetContextActions(engine Engine) []util.MenuItem {
    return inventoryItemActions(k, engine)
}

func (k *Key) Icon(uint64) int32 {
    return k.icon
}
func (k *Key) Encode() string {
    // encode name, key, and importance
    return recfile.ToPredicate("key", k.name, k.key, strconv.Itoa(k.importance))
}
func (k *Key) SetHolder(holder ItemHolder) {
    k.BaseItem.SetHolder(holder)
    if k.firstOwnerName == "" && holder.Name() != "Party" {
        k.firstOwnerName = holder.Name()
    }
}

func (k *Key) KeySource() string {
    if k.firstOwnerName != "" {
        return k.firstOwnerName
    }
    if k.FoundInMap != "" {
        return k.FoundInMap
    }
    return "unknown"
}

func (k *Key) GetKeyString() string {
    return k.key
}
func NewKeyFromPredicate(encoded recfile.StringPredicate) *Key {
    return NewKeyFromImportance(
        encoded.GetString(0),
        encoded.GetString(1),
        encoded.GetInt(2),
    )
}
func NewKeyFromImportance(name, key string, importance int) *Key {
    return &Key{
        BaseItem: BaseItem{
            name: name,
        },
        icon:       182,
        importance: importance,
        key:        key,
    }
}
func keyColor(importance int) color.Color {
    switch importance {
    case 1:
        return ega.BrightBlack
    case 2:
        return ega.White
    case 3:
        return ega.BrightWhite
    case 4:
        return ega.Yellow
    case 5:
        return ega.BrightYellow
    }
    return ega.White
}
