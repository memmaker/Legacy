package game

import (
    "Legacy/recfile"
    "Legacy/renderer"
    "fmt"
    "math/rand"
    "strconv"
)

type ArmorKind string

const (
    ArmorMaterialCloth   ArmorKind = "cloth"
    ArmorMaterialLeather ArmorKind = "leather"
    ArmorMaterialChain   ArmorKind = "chain"
    ArmorMaterialPlate   ArmorKind = "plate"
    ArmorMaterialMagical ArmorKind = "magical"
)

type Armor struct {
    BaseItem
    protection int
    slot       EquipmentSlot
    wearer     ItemWearer
    material   ArmorKind
}

func (a *Armor) InventoryIcon() int32 {
    materialOffsets := map[ArmorKind]int32{
        ArmorMaterialLeather: 180,
        ArmorMaterialChain:   178,
        ArmorMaterialPlate:   160,
        ArmorMaterialMagical: 182,
    }
    if material, ok := materialOffsets[a.material]; ok {
        if a.slot == ArmorSlotHead {
            return material + 1
        }
        return material
    }
    if a.slot == ArmorSlotHead {
        return 181
    } else {
        return 180
    }
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
    return int32(205)
}
func (a *Armor) GetContextActions(engine Engine) []renderer.MenuItem {
    baseActions := inventoryItemActions(a, engine)
    if !engine.IsPlayerControlled(a.GetHolder()) {
        return baseActions
    }
    equipAction := renderer.MenuItem{
        Text: "Equip",
        Action: func() {
            engine.ShowEquipMenu(a)
        },
    }
    equipActions := []renderer.MenuItem{equipAction}
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

func NewArmor(material ArmorKind, slot EquipmentSlot, protection int) *Armor {
    return &Armor{
        BaseItem: BaseItem{
            name: nameFromSlotAndMaterial(slot, material),
        },
        material:   material,
        slot:       slot,
        protection: protection,
    }
}
func NewRandomArmor(lootLevel int) *Armor {
    slot := randomSlot()
    protection := damage(slot, lootLevel)
    material := materialFromLootLevel(lootLevel)
    return NewArmor(material, slot, protection)
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

func nameFromSlotAndMaterial(slot EquipmentSlot, material ArmorKind) string {
    switch slot {
    case ArmorSlotHead:
        return helmetName(material)
    case ArmorSlotTorso:
        return armorName(material)
    }
    return "unknown"
}

func armorName(material ArmorKind) string {
    switch material {
    case ArmorMaterialCloth:
        return "cloth doublet"
    case ArmorMaterialLeather:
        return "leather jerkin"
    case ArmorMaterialChain:
        return "chain mail"
    case ArmorMaterialPlate:
        return "plate armor"
    case ArmorMaterialMagical:
        return "magical plate armor"
    }
    return fmt.Sprintf("%s armor", material)
}

func helmetName(material ArmorKind) string {
    switch material {
    case ArmorMaterialCloth:
        return "cloth cap"
    case ArmorMaterialLeather:
        return "leather cap"
    case ArmorMaterialChain:
        return "chain coif"
    case ArmorMaterialPlate:
        return "plate helmet"
    case ArmorMaterialMagical:
        return "magical plate helmet"
    }
    return fmt.Sprintf("%s helmet", material)
}

func materialFromLootLevel(level int) ArmorKind {
    switch level {
    case 1:
        return ArmorMaterialLeather
    case 2:
        return ArmorMaterialChain
    case 3:
        return ArmorMaterialPlate
    case 4:
        return ArmorMaterialMagical
    }
    return "unknown"
}
func NewArmorFromPredicate(encoded recfile.StringPredicate) *Armor {
    // extract name, slot, and protection
    return NewArmor(
        ArmorKind(encoded.GetString(0)),
        EquipmentSlot(encoded.GetString(1)),
        encoded.GetInt(2),
    )
}

func (a *Armor) Encode() string {
    return recfile.ToPredicate("armor", string(a.material), string(a.slot), strconv.Itoa(a.protection))
}
