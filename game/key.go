package game

import (
    "Legacy/ega"
    "Legacy/renderer"
    "fmt"
    "image/color"
    "regexp"
    "strconv"
)

type Key struct {
    BaseItem
    icon       int
    key        string
    importance int
}

func (k *Key) CanStackWith(other Item) bool {
    return false
}

func (k *Key) TintColor() color.Color {
    return keyColor(k.importance)
}

func (k *Key) GetContextActions(engine Engine) []renderer.MenuItem {
    return inventoryItemActions(k, engine)
}

func (k *Key) Icon(uint64) int {
    return k.icon
}
func (k *Key) Encode() string {
    // encode name, key, and importance
    return fmt.Sprintf("%s, %s, %d", k.name, k.key, k.importance)
}

func NewKeyFromString(encoded string) *Key {
    paramRegex := regexp.MustCompile(`^([^,]+), ?([^,]+), ?(\d+)$`)
    // extract name, key, and importance
    var name, key string
    var importance int
    if matches := paramRegex.FindStringSubmatch(encoded); matches != nil {
        name = matches[1]
        key = matches[2]
        importance, _ = strconv.Atoi(matches[3])
        return NewKeyFromImportance(name, key, importance)
    }
    return NewKeyFromImportance("Invalid Key", "invalid_key", 1)
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
