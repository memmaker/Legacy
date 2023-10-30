package game

import (
    "Legacy/renderer"
    "fmt"
    "math/rand"
    "regexp"
    "strconv"
)

type ArmorSlot string

const (
    ArmorSlotHead  ArmorSlot = "head"
    ArmorSlotTorso ArmorSlot = "torso"
)

type Armor struct {
    BaseItem
    protection int
    slot       ArmorSlot
    wearer     ItemWearer
}

func (a *Armor) GetWearer() ItemWearer {
    return a.wearer
}

func (a *Armor) Icon(u uint64) int {
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

func (a *Armor) Unequip(item Item) {
    if a.wearer != nil {
        a.wearer.Unequip(item)
    }
}

func NewArmor(name string, slot ArmorSlot, protection int) *Armor {
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
    protection := protectionFromSlot(slot, lootLevel)
    armorName := nameFromSlot(slot, lootLevel)
    return NewArmor(armorName, slot, protection)
}

func protectionFromSlot(slot ArmorSlot, level int) int {
    switch slot {
    case ArmorSlotHead:
        return 1 + level*2
    case ArmorSlotTorso:
        return 2 + level*3
    }
    return 0
}

func randomSlot() ArmorSlot {
    randomInt := rand.Intn(2)
    if randomInt == 0 {
        return ArmorSlotHead
    } else {
        return ArmorSlotTorso
    }
}

func nameFromSlot(slot ArmorSlot, level int) string {
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
func NewArmorFromString(encoded string) *Armor {
    // extract name, slot, and protection
    paramRegex := regexp.MustCompile(`^([^,]+), ?([^,]+), ?(\d+)$`)
    var name string
    var slot ArmorSlot
    var protection int
    if matches := paramRegex.FindStringSubmatch(encoded); matches != nil {
        name = matches[1]
        slot = ArmorSlot(matches[2])
        protection, _ = strconv.Atoi(matches[3])
        return NewArmor(name, slot, protection)
    }
    println("Invalid Armor: " + encoded)
    return NewArmor("Invalid Armor", ArmorSlotHead, 0)
}
