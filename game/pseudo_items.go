package game

import (
    "Legacy/renderer"
    "fmt"
    "regexp"
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

func (p *PseudoItem) Icon(u uint64) int {
    return 0
}

func (p *PseudoItem) GetContextActions(engine Engine) []renderer.MenuItem {
    var actions []renderer.MenuItem
    if p.GetHolder() == nil {
        actions = append(actions, renderer.MenuItem{
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
            name: nameFromTypeAndAmount(itemType, amount),
        },
        itemType: itemType,
        amount:   amount,
    }
}

func NewPseudoItemFromString(encoded string) *PseudoItem {
    paramRegex := regexp.MustCompile(`^([^,]+), ?([^,]+), ?(\d+)$`)
    nameLessParamRegex := regexp.MustCompile(`^([^,]+), ?(\d+)$`)
    var name string
    var itemType PseudoItemType
    var amount int
    if matches := paramRegex.FindStringSubmatch(encoded); matches != nil && len(matches) == 4 {
        name = matches[1]
        itemType = PseudoItemType(matches[2])
        amount, _ = strconv.Atoi(matches[3])
        return NewPseudoItem(name, itemType, amount)
    } else if matches = nameLessParamRegex.FindStringSubmatch(encoded); matches != nil && len(matches) == 3 {
        itemType = PseudoItemType(matches[1])
        amount, _ = strconv.Atoi(matches[2])
        return NewPseudoItem(nameFromTypeAndAmount(itemType, amount), itemType, amount)
    }
    return NewPseudoItem("Invalid PseudoItem", PseudoItemTypeGold, 0)
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
