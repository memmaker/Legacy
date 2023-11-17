package game

import (
    "Legacy/ega"
    "Legacy/recfile"
    "Legacy/util"
    "fmt"
    "image/color"
    "math/rand"
)

type ItemTier string

func (t ItemTier) Multiplier() float64 {
    switch t {
    case ItemTierCommon:
        return 1
    case ItemTierUncommon:
        return 1.2
    case ItemTierRare:
        return 1.5
    case ItemTierLegendary:
        return 3
    }
    return 0
}

func (t ItemTier) Color() color.Color {
    switch t {
    case ItemTierCommon:
        return color.White
    case ItemTierUncommon:
        return ega.BrightCyan
    case ItemTierRare:
        return ega.BrightYellow
    case ItemTierLegendary:
        return ega.BrightMagenta
    }
    return color.White
}

func (t ItemTier) AsInt() int {
    switch t {
    case ItemTierCommon:
        return 1
    case ItemTierUncommon:
        return 2
    case ItemTierRare:
        return 3
    case ItemTierLegendary:
        return 4
    }
    return 0
}

const (
    ItemTierCommon    ItemTier = "common"
    ItemTierUncommon  ItemTier = "uncommon"
    ItemTierRare      ItemTier = "rare"
    ItemTierLegendary ItemTier = "legendary"
)

func GetAllTiers() []ItemTier {
    return []ItemTier{
        ItemTierCommon,
        ItemTierUncommon,
        ItemTierRare,
        ItemTierLegendary,
    }
}

type ArmorModifier string

func (m ArmorModifier) ValueMultiplier() float64 {
    switch m {
    case ArmorMaterialCloth:
        return 0.7
    case ArmorMaterialLeather:
        return 1
    case ArmorMaterialChain:
        return 1.5
    case ArmorMaterialPlate:
        return 2
    case ArmorMaterialMagical:
        return 3
    }
    return 0
}

const (
    ArmorMaterialCloth   ArmorModifier = "cloth"
    ArmorMaterialLeather ArmorModifier = "leather"
    ArmorMaterialChain   ArmorModifier = "chain"
    ArmorMaterialPlate   ArmorModifier = "plate"
    ArmorMaterialMagical ArmorModifier = "magical"
)

func GetAllArmorModifiers() []ArmorModifier {
    return []ArmorModifier{
        ArmorMaterialCloth,
        ArmorMaterialLeather,
        ArmorMaterialChain,
        ArmorMaterialPlate,
        ArmorMaterialMagical,
    }
}

type Armor struct {
    BaseItem
    slot     ArmorSlot
    wearer   ItemWearer
    material ArmorModifier
    level    ItemTier
}

func (a *Armor) GetTooltipLines() []string {
    rows := []util.TableRow{
        {Label: "Level", Columns: []string{string(a.level)}},
        {Label: "Type", Columns: []string{string(a.slot)}},
        {Label: "Material", Columns: []string{string(a.material)}},
        {Label: "Protection", Columns: []string{fmt.Sprintf("%d", a.GetProtection())}},
        {Label: "Value", Columns: []string{fmt.Sprintf("%dg", a.GetValue())},
        },
    }
    return util.TableLayout(rows)
}

func (a *Armor) InventoryIcon() int32 {
    if a.slot == AccessorySlotRobe {
        return 188
    }
    if a.slot == ArmorSlotShoes {
        return 189
    }
    if a.slot == AccessorySlotAmuletOne || a.slot == AccessorySlotAmuletTwo {
        return 190
    }
    if a.slot == AccessorySlotRingLeft || a.slot == AccessorySlotRingRight {
        return 168
    }
    materialOffsets := map[ArmorModifier]int32{
        ArmorMaterialLeather: 180,
        ArmorMaterialChain:   178,
        ArmorMaterialPlate:   160,
        ArmorMaterialMagical: 182,
    }
    if material, ok := materialOffsets[a.material]; ok {
        if a.slot == ArmorSlotHelmet {
            return material + 1
        }
        return material
    }
    if a.slot == ArmorSlotHelmet {
        return 181
    } else {
        return 180
    }
}

func (a *Armor) GetWearer() ItemWearer {
    return a.wearer
}
func (a *Armor) GetValue() int {
    //protectionValue := max(a.GetProtection(), 10)
    baseValue := float64(a.slot.BaseValue()) * a.level.Multiplier() * a.material.ValueMultiplier()
    return int(baseValue)
}
func (a *Armor) Icon(u uint64) int32 {
    if a.slot == ArmorSlotHelmet {
        return int32(222)
    }
    return int32(221)
}
func (a *Armor) GetContextActions(engine Engine) []util.MenuItem {
    baseActions := inventoryItemActions(a, engine)
    if !engine.IsPlayerControlled(a.GetHolder()) {
        return baseActions
    }
    equipAction := util.MenuItem{
        Text: "Equip",
        Action: func() {
            engine.ShowEquipMenu(a)
        },
    }
    equipActions := []util.MenuItem{equipAction}
    return append(equipActions, baseActions...)
}

func (a *Armor) CanStackWith(other Item) bool {
    if otherArmor, ok := other.(*Armor); ok {
        return a.name == otherArmor.name && a.GetLevel() == otherArmor.GetLevel() && a.GetProtection() == otherArmor.GetProtection() && a.GetSlot().IsEqualTo(otherArmor.slot) && a.wearer == otherArmor.wearer
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

func (a *Armor) GetSlot() ArmorSlot {
    return a.slot
}

func (a *Armor) IsBetterThan(other Wearable) bool {
    if other == nil {
        return true
    }
    armor, isArmor := other.(*Armor)
    if !isArmor || armor == nil {
        return true
    }
    return a.GetProtection() > armor.GetProtection()
}

func (a *Armor) IsEquipped() bool {
    return a.wearer != nil
}

func NewArmor(level ItemTier, slot ArmorSlot, material ArmorModifier) *Armor {
    return &Armor{
        BaseItem: BaseItem{
            name: nameFromSlotAndMaterial(slot, material),
        },
        material: material,
        slot:     slot,
        level:    level,
    }
}
func NewRandomArmor(lootLevel int) *Armor {
    slot := randomSlot()
    material := materialFromLootLevel(lootLevel)
    randomTier := tierFromLootLevel(lootLevel)
    return NewArmor(randomTier, slot, material)
}

func NewRandomArmorForVendor(lootLevel int) *Armor {
    slot := randomSlot()
    material := materialFromLootLevel(lootLevel)
    randomTier := "common"
    if lootLevel > 1 {
        randomTier = "uncommon"
    }
    return NewArmor(ItemTier(randomTier), slot, material)
}

func tierFromLootLevel(lootLevel int) ItemTier {
    chanceForLegendary := 0.01
    chanceForRare := 0.04
    chanceForUncommon := 0.25
    randFloat := rand.Float64() - (float64(lootLevel) * 0.05)
    if randFloat < chanceForLegendary {
        return ItemTierLegendary
    }
    if randFloat < chanceForLegendary+chanceForRare {
        return ItemTierRare
    }
    if randFloat < chanceForLegendary+chanceForRare+chanceForUncommon {
        return ItemTierUncommon
    }
    return ItemTierCommon
}

func protectionForArmor(slot ArmorSlot, material ArmorModifier, level ItemTier) int {
    return baseProtectionForArmor(slot, level) + materialProtection(material, level)
}

func materialProtection(material ArmorModifier, level ItemTier) int {
    switch material {
    case ArmorMaterialCloth:
        return 1
    case ArmorMaterialLeather:
        return int(float64(5) * level.Multiplier())
    case ArmorMaterialChain:
        return int(float64(10) * level.Multiplier())
    case ArmorMaterialPlate:
        return int(float64(15) * level.Multiplier())
    case ArmorMaterialMagical:
        return int(float64(20) * level.Multiplier())
    }
    return 0
}

func baseProtectionForArmor(slot ArmorSlot, level ItemTier) int {
    switch slot {
    case ArmorSlotHelmet:
        return int(float64(10) * level.Multiplier())
    case ArmorSlotBreastPlate:
        return int(float64(20) * level.Multiplier())
    case ArmorSlotShoes:
        return int(float64(5) * level.Multiplier())
    case AccessorySlotRobe:
        return 1
    case AccessorySlotAmuletOne:
        fallthrough
    case AccessorySlotAmuletTwo:
        return 0
    case AccessorySlotRingLeft:
        fallthrough
    case AccessorySlotRingRight:
        return 0
    }
    return 0
}

func randomSlot() ArmorSlot {
    randomInt := rand.Intn(8)
    switch randomInt {
    case 0:
        return ArmorSlotHelmet
    case 1:
        return ArmorSlotBreastPlate
    case 2:
        return ArmorSlotShoes
    case 3:
        return AccessorySlotRobe
    case 4:
        return AccessorySlotAmuletOne
    case 5:
        return AccessorySlotAmuletTwo
    case 6:
        return AccessorySlotRingLeft
    case 7:
        return AccessorySlotRingRight
    }
    return ArmorSlotBreastPlate
}
func (a *Armor) Name() string {
    if a.name != "" {
        return a.name
    }
    return nameFromSlotAndMaterial(a.slot, a.material)
}

func (a *Armor) TintColor() color.Color {
    return a.level.Color()
}
func nameFromSlotAndMaterial(slot ArmorSlot, material ArmorModifier) string {
    switch slot {
    case ArmorSlotHelmet:
        return helmetName(material)
    case ArmorSlotBreastPlate:
        return armorName(material)
    case ArmorSlotShoes:
        return shoeName(material)
    case AccessorySlotRobe:
        return robeName(material)
    case AccessorySlotAmuletOne:
        fallthrough
    case AccessorySlotAmuletTwo:
        return amuletName(material)
    case AccessorySlotRingLeft:
        fallthrough
    case AccessorySlotRingRight:
        return ringName(material)
    }
    return "unknown"
}

func ringName(material ArmorModifier) string {
    switch material {
    case ArmorMaterialCloth:
        return "plain ring"
    case ArmorMaterialLeather:
        return "braided ring"
    case ArmorMaterialChain:
        return "silver ring"
    case ArmorMaterialPlate:
        return "gold ring"
    case ArmorMaterialMagical:
        return "magical ring"
    }
    return fmt.Sprintf("%s ring", material)
}

func amuletName(material ArmorModifier) string {
    switch material {
    case ArmorMaterialCloth:
        return "plain amulet"
    case ArmorMaterialLeather:
        return "braided amulet"
    case ArmorMaterialChain:
        return "silver amulet"
    case ArmorMaterialPlate:
        return "gold amulet"
    case ArmorMaterialMagical:
        return "magical amulet"
    }
    return fmt.Sprintf("%s amulet", material)
}

func robeName(material ArmorModifier) string {
    switch material {
    case ArmorMaterialCloth:
        return "cloth robe"
    case ArmorMaterialLeather:
        return "fur cloak"
    case ArmorMaterialChain:
        return "red hooded cloak"
    case ArmorMaterialPlate:
        return "royal mantle"
    case ArmorMaterialMagical:
        return "magical cloak"
    }
    return fmt.Sprintf("%s robe", material)
}

func shoeName(material ArmorModifier) string {
    switch material {
    case ArmorMaterialCloth:
        return "cloth sandals"
    case ArmorMaterialLeather:
        return "leather boots"
    case ArmorMaterialChain:
        return "chain boots"
    case ArmorMaterialPlate:
        return "plate boots"
    case ArmorMaterialMagical:
        return "magical plate boots"
    }
    return fmt.Sprintf("%s boots", material)
}

func armorName(material ArmorModifier) string {
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

func helmetName(material ArmorModifier) string {
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

func materialFromLootLevel(level int) ArmorModifier {
    randLevel := min(rand.Intn(level)+1, 5)
    switch randLevel {
    case 1:
        return ArmorMaterialCloth
    case 2:
        return ArmorMaterialLeather
    case 3:
        return ArmorMaterialChain
    case 4:
        return ArmorMaterialPlate
    case 5:
        return ArmorMaterialMagical
    }
    return ArmorMaterialCloth
}
func NewArmorFromPredicate(encoded recfile.StringPredicate) *Armor {
    // extract name, slot, and protection
    // example: armor(common, helmet, cloth)
    armor := NewArmor(ItemTier(encoded.GetString(0)), ArmorSlot(encoded.GetString(1)), ArmorModifier(encoded.GetString(2)))
    if encoded.ParamCount() > 3 {
        armor.SetName(encoded.GetString(3))
    }
    return armor
}

func (a *Armor) Encode() string {
    return recfile.ToPredicate("armor", string(a.level), string(a.slot), string(a.material))
}

func (a *Armor) IsAccessory() bool {
    return a.slot == AccessorySlotAmuletOne || a.slot == AccessorySlotAmuletTwo || a.slot == AccessorySlotRingLeft || a.slot == AccessorySlotRingRight || a.slot == AccessorySlotRobe
}

func (a *Armor) SetSlotUsed(slot ArmorSlot) {
    a.slot = slot
}

func (a *Armor) GetLevel() ItemTier {
    return a.level
}

func (a *Armor) GetProtection() int {
    return protectionForArmor(a.slot, a.material, a.level)
}
