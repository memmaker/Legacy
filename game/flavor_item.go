package game

import (
    "Legacy/recfile"
    "Legacy/util"
    "fmt"
    "image/color"
)

type FlavorItem struct {
    BaseItem
    description []string
}

func (f *FlavorItem) GetTooltipLines() []string {
    return f.description
}

func (f *FlavorItem) InventoryIcon() int32 {
    return 169
}

func (f *FlavorItem) Icon(u uint64) int32 {
    return int32(205)
}

func (f *FlavorItem) GetContextActions(engine Engine) []util.MenuItem {
    actions := inventoryItemActions(f, engine)
    if len(f.description) > 0 {
        actions = append(actions, util.MenuItem{
            Text: "Examine",
            Action: func() {
                engine.ShowScrollableText(f.description, color.White, true)
            },
        })
    }
    return actions
}

func (f *FlavorItem) CanStackWith(other Item) bool {
    if otherFlavor, ok := other.(*FlavorItem); ok {
        return f.name == other.Name() && f.baseValue == otherFlavor.baseValue && len(f.description) == len(otherFlavor.description)
    } else {
        return false
    }
}

func NewFlavorItem(name string, value int) *FlavorItem {
    return &FlavorItem{
        BaseItem: BaseItem{
            name:      name,
            baseValue: value,
        },
    }
}

func (f *FlavorItem) Encode() string {
    //TODO
    //return recfile.ToPredicate("flavor", f.name, strconv.Itoa(f.baseValue), f.description)
    return fmt.Sprintf("flavor(%s, %d, %s)", f.name, f.baseValue, f.description) // TODO: escape commas in description.. or sth..
}
func NewFlavorItemFromPredicate(encoded recfile.StringPredicate) *FlavorItem {
    name := encoded.GetString(0)
    value := encoded.GetInt(1)
    description := encoded.GetString(2)
    return &FlavorItem{
        BaseItem: BaseItem{
            name:      name,
            baseValue: value,
        },
        description: []string{description},
    }
}

func (f *FlavorItem) SetDescription(description []string) {
    f.description = description
}
