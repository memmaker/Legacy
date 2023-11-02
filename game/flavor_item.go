package game

import (
    "Legacy/recfile"
    "Legacy/renderer"
    "fmt"
    "image/color"
)

type FlavorItem struct {
    BaseItem
    description []string
}

func (f *FlavorItem) Icon(u uint64) int32 {
    return 0
}

func (f *FlavorItem) GetContextActions(engine Engine) []renderer.MenuItem {
    actions := inventoryItemActions(f, engine)
    if len(f.description) > 0 {
        actions = append(actions, renderer.MenuItem{
            Text: "Examine",
            Action: func() {
                engine.ShowColoredText(f.description, color.White, true)
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
    return fmt.Sprintf("flavor(%s|%d, %s)", f.name, f.baseValue, f.description) // TODO: escape commas in description.. or sth..
}
func NewFlavorItemFromPredicate(encoded recfile.StringPredicate) *FlavorItem {
    var name string
    var value int
    // TODO
    return &FlavorItem{
        BaseItem: BaseItem{
            name:      name,
            baseValue: value,
        },
    }
}

func (f *FlavorItem) SetDescription(description []string) {
    f.description = description
}
