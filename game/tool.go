package game

import (
    "Legacy/recfile"
    "Legacy/util"
    "fmt"
    "image/color"
)

type ToolType string

const (
    ToolTypePickaxe      ToolType = "pickaxe"
    ToolTypeShovel       ToolType = "shovel"
    ToolTypeRope         ToolType = "rope"
    ToolTypeWoodenPlanks ToolType = "wooden planks"
)

type Tool struct {
    BaseItem
    kind ToolType
}

func (t *Tool) GetTooltipLines() []string {
    return []string{}
}

func (t *Tool) InventoryIcon() int32 {
    if t.kind == ToolTypePickaxe {
        return 176
    } else if t.kind == ToolTypeShovel {
        return 177
    }
    return 169
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

func (t *Tool) GetContextActions(engine Engine) []util.MenuItem {
    return inventoryItemActions(t, engine)
}

func (t *Tool) Icon(uint64) int32 {
    return int32(205)
}
func (t *Tool) Encode() string {
    return fmt.Sprintf("tool(%s, %s)", t.kind, t.name)
}

func NewTool(kind ToolType, name string) *Tool {
    return &Tool{
        BaseItem: BaseItem{
            name: name,
        },
        kind: kind,
    }
}
func NewToolFromPredicate(pred recfile.StringPredicate) *Tool {
    if pred.ParamCount() == 1 {
        return NewTool(ToolType(pred.GetString(0)), "")
    }
    return NewTool(ToolType(pred.GetString(0)), pred.GetString(1))
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
func (t *Tool) Name() string {
    if t.name == "" {
        return string(t.kind)
    }
    return t.name
}
