package game

import (
    "Legacy/recfile"
    "Legacy/renderer"
    "fmt"
    "math/rand"
    "strconv"
)

type Armor struct {
    BaseItem
    protection int
    slot       EquipmentSlot
    wearer     ItemWearer
}

func (a *Armor) GetWearer() ItemWearer {
    return a.wearer
}
func (a *Armor) GetValue() int {
    if a.slot == ArmorSlotHead {
        return a.protection * 7
    }
    return a.protection * 10
}
func (a *Armor) Icon(u uint64) int32 {
    return 0
}
func (a *Armor) GetContextActions(engine Engine) []renderer.MenuItem {
    // TODO: add equip
    equipAction := renderer.MenuItem{
        Text: "Equip",
        Action: func() {
            engine.ShowEquipMenu(a)
        },
    }
    equipActions := []renderer.MenuItem{equipAction}
    baseActions := inventoryItemActions(a, engine)
    return append(equipActions, baseActions...)
}

func (a *Armor) CanStackWith(other Item) bool {
    if otherArmor, ok := other.(*Armor); ok {
        return a.name == otherArmor.name && a.protection == otherArmor.protection && a.slot == otherArmor.slot && a.wearer == otherArmor.wearer
    } else {
        return false
    }
}

func (a *Armor) SetWearer(wearer ItemWearer) {
    a.wearer = wearer
}

func (a *Armor) Unequip() {
    if a.wearer != nil {
        a.wearer.Unequip(a)
    }
}

func (a *Armor) GetSlot() EquipmentSlot {
    return a.slot
}

func (a *Armor) IsBetterThan(other *Armor) bool {
    if other == nil {
        return true
    }
    return a.protection > other.protection
}

func (a *Armor) IsEquipped() bool {
    return a.wearer != nil
}

func NewArmor(name string, slot EquipmentSlot, protection int) *Armor {
    return &Armor{
        BaseItem: BaseItem{
            name: name,
        },
        slot:       slot,
        protection: protection,
    }
}
func NewRandomArmor(lootLevel int) *Armor {
    slot := randomSlot()
    protection := damage(slot, lootLevel)
    armorName := nameFromSlot(slot, lootLevel)
    return NewArmor(armorName, slot, protection)
}

func damage(slot EquipmentSlot, level int) int {
    switch slot {
    case ArmorSlotHead:
        return 1 + level*2
    case ArmorSlotTorso:
        return 2 + level*3
    }
    return 0
}

func randomSlot() EquipmentSlot {
    randomInt := rand.Intn(2)
    if randomInt == 0 {
        return ArmorSlotHead
    } else {
        return ArmorSlotTorso
    }
}

func nameFromSlot(slot EquipmentSlot, level int) string {
    materialName := materialNameFromLootLevel(level)
    switch slot {
    case ArmorSlotHead:
        return fmt.Sprintf("%s helmet", materialName)
    case ArmorSlotTorso:
        return fmt.Sprintf("%s armor", materialName)
    }
    return "unknown"
}

func materialNameFromLootLevel(level int) string {
    switch level {
    case 0:
        return "common"
    case 1:
        return "cloth"
    case 2:
        return "leather"
    case 3:
        return "chain"
    case 4:
        return "plate"
    case 5:
        return "mythical"
    }
    return "unknown"
}
func NewArmorFromPredicate(encoded recfile.StringPredicate) *Armor {
    // extract name, slot, and protection
    return NewArmor(
        encoded.GetString(0),
        EquipmentSlot(encoded.GetString(1)),
        encoded.GetInt(2),
    )
}

func (a *Armor) Encode() string {
    return recfile.ToPredicate("armor", a.name, string(a.slot), strconv.Itoa(a.protection))
}
