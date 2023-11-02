package game

import (
    "Legacy/renderer"
    "fmt"
    "image/color"
    "regexp"
)

type ToolType string

const (
    ToolTypePickaxe ToolType = "pickaxe"
    ToolTypeShovel  ToolType = "shovel"
)

type Tool struct {
    BaseItem
    kind ToolType
}

func (t *Tool) CanStackWith(other Item) bool {
    if otherTool, ok := other.(*Tool); ok {
        return t.kind == otherTool.kind
    }
    return false
}

func (t *Tool) TintColor() color.Color {
    return color.White
}

func (t *Tool) GetContextActions(engine Engine) []renderer.MenuItem {
    return inventoryItemActions(t, engine)
}

func (t *Tool) Icon(uint64) int32 {
    return 0
}
func (t *Tool) Encode() string {
    // encode name, key, and importance
    return fmt.Sprintf("%s: %s", t.kind, t.name)
}

func NewTool(kind ToolType, name string) *Tool {
    return &Tool{
        BaseItem: BaseItem{
            name: name,
        },
        kind: kind,
    }
}
func NewToolFromString(encoded string) *Tool {
    paramRegex := regexp.MustCompile(`^([^:]+): ?([^:]+)$`)
    // extract name, key, and importance
    var name string
    var kind ToolType

    if matches := paramRegex.FindStringSubmatch(encoded); matches != nil {
        kind = ToolType(matches[1])
        name = matches[2]
        return NewTool(kind, name)
    }
    return NewTool(ToolTypePickaxe, "Invalid Tool")
}

func (t *Tool) GetKind() ToolType {
    return t.kind
}

func (t *Tool) Use(engine Engine) {
    if t.kind == ToolTypePickaxe {
        t.usePickaxe(engine)
    } else if t.kind == ToolTypeShovel {
        t.useShovel(engine)
    }
}

func (t *Tool) usePickaxe(engine Engine) {
    // check if there is a breakable wall in front of the player
    // if so, break it
    // TODO
}

func (t *Tool) useShovel(engine Engine) {
    // TODO: dig here
}
