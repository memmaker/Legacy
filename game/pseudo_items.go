package game

import (
    "Legacy/recfile"
    "Legacy/util"
    "fmt"
    "strconv"
)

type PseudoItemType string

const (
    PseudoItemTypeGold     PseudoItemType = "gold"
    PseudoItemTypeFood     PseudoItemType = "food"
    PseudoItemTypeLockpick PseudoItemType = "lockpick"
)

type PseudoItem struct {
    BaseItem
    itemType PseudoItemType
    amount   int
}

func (p *PseudoItem) GetTooltipLines() []string {
    return []string{}
}

func (p *PseudoItem) InventoryIcon() int32 {
    return 169
}

func (p *PseudoItem) Icon(u uint64) int32 {
    return int32(205)
}

func (p *PseudoItem) GetContextActions(engine Engine) []util.MenuItem {
    var actions []util.MenuItem
    if p.GetHolder() == nil {
        actions = append(actions, util.MenuItem{
            Text: fmt.Sprintf("Take \"%s\"", p.Name()),
            Action: func() {
                if p.GetHolder() == nil {
                    p.Take(engine)
                }
            },
        })
    }
    return actions
}

func (p *PseudoItem) Take(engine Engine) {
    switch p.itemType {
    case PseudoItemTypeGold:
        engine.AddGold(p.amount)
        engine.RemoveItem(p)
    case PseudoItemTypeFood:
        engine.AddFood(p.amount)
        engine.RemoveItem(p)
    case PseudoItemTypeLockpick:
        engine.AddLockpicks(p.amount)
        engine.RemoveItem(p)
    }
}

func (p *PseudoItem) CanStackWith(other Item) bool {
    if otherPseudoItem, ok := other.(*PseudoItem); ok {
        return p.itemType == otherPseudoItem.itemType && p.amount == otherPseudoItem.amount && p.name == otherPseudoItem.name
    } else {
        return false
    }
}
func (p *PseudoItem) Name() string {
    if p.name == "" {
        return nameFromTypeAndAmount(p.itemType, p.amount)
    }
    return p.name
}
func NewPseudoItem(name string, itemType PseudoItemType, amount int) *PseudoItem {
    return &PseudoItem{
        BaseItem: BaseItem{
            name: name,
        },
        itemType: itemType,
        amount:   amount,
    }
}

func NewPseudoItemFromTypeAndAmount(itemType PseudoItemType, amount int) *PseudoItem {
    return &PseudoItem{
        BaseItem: BaseItem{
            name: "",
        },
        itemType: itemType,
        amount:   amount,
    }
}

func NewPseudoItemFromPredicate(encoded recfile.StringPredicate) *PseudoItem {
    if encoded.ParamCount() == 3 {
        return NewPseudoItem(
            encoded.GetString(0),
            PseudoItemType(encoded.GetString(1)),
            encoded.GetInt(2),
        )
    }
    itemType := PseudoItemType(encoded.GetString(0))
    amount := encoded.GetInt(1)
    return NewPseudoItemFromTypeAndAmount(itemType, amount)

}
func (p *PseudoItem) Encode() string {
    if p.name == "" {
        return recfile.ToPredicateSep("noitem", string(p.itemType), strconv.Itoa(p.amount))
    }
    return recfile.ToPredicate("noitem", p.name, string(p.itemType), strconv.Itoa(p.amount))
}
func nameFromTypeAndAmount(itemType PseudoItemType, amount int) string {
    if amount == 1 {
        switch itemType {
        case PseudoItemTypeGold:
            return "a Gold Coin"
        case PseudoItemTypeFood:
            return "a ration of Food"
        case PseudoItemTypeLockpick:
            return "a Lockpick"
        }
    }

    switch itemType {
    case PseudoItemTypeGold:
        return fmt.Sprintf("%d Gold Coins", amount)
    case PseudoItemTypeFood:
        return fmt.Sprintf("%d rations of Food", amount)
    case PseudoItemTypeLockpick:
        return fmt.Sprintf("%d Lockpicks", amount)
    }
    return "Invalid PseudoItem"
}
