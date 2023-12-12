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
    actions     []Action
    wearer      ItemWearer
}

func (f *FlavorItem) GetWearer() ItemWearer {
    return f.wearer
}

func (f *FlavorItem) SetWearer(wearer ItemWearer) {
    f.wearer = wearer
}

func (f *FlavorItem) Unequip() {
    if f.wearer != nil {
        f.wearer.Unequip(f)
    }
}

func (f *FlavorItem) IsEquipped() bool {
    return f.wearer != nil
}

func (f *FlavorItem) IsBetterThan(other Handheld) bool {
    return false
}

func (f *FlavorItem) GetTooltipLines() []string {
    baseDescription := f.description
    if len(f.actions) == 0 {
        return baseDescription
    }
    baseDescription = append(baseDescription, "")
    baseDescription = append(baseDescription, "Actions:")
    for _, action := range f.actions {
        baseDescription = append(baseDescription, fmt.Sprintf("  %s", action.Name()))
    }

    return baseDescription
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
    if len(f.actions) > 0 && engine.IsPlayerControlled(f.GetHolder()) {
        if f.IsEquipped() {
            return append([]util.MenuItem{{
                Text: "Unequip",
                Action: func() {
                    f.Unequip()
                }}}, actions...)
        } else {
            return append([]util.MenuItem{{
                Text: "Equip",
                Action: func() {
                    engine.ShowEquipMenu(f)
                }}}, actions...)
        }
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

func (f *FlavorItem) GetEmbeddedActions() []Action {
    return f.actions
}

func (f *FlavorItem) SetActions(actions []Action) {
    f.actions = actions
}

func NewNamedFlavorItem(itemName string) *FlavorItem {
    switch itemName {
    case "slime stick":
        item := NewFlavorItem("slime stick", 10)
        item.SetActions([]Action{
            NewActiveSkillFromName(CombatSkillJellyJab),
        })
        item.SetDescription([]string{
            "A stick covered in slime.",
        })

        return item

    }
    println("ERR: unknown item name:", itemName)
    return nil

}
