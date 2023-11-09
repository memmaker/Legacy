package game

import (
    "Legacy/recfile"
    "Legacy/renderer"
    "Legacy/util"
    "fmt"
    "image/color"
    "math/rand"
)

type WeaponType string

const (
    WeaponTypeSword      WeaponType = "sword"
    WeaponTypeGreatSword WeaponType = "great sword"
    WeaponTypeSpear      WeaponType = "spear"
    WeaponTypeStaff      WeaponType = "staff"
    WeaponTypeDagger     WeaponType = "dagger"

    WeaponTypeBow      WeaponType = "bow"
    WeaponTypeCrossbow WeaponType = "crossbow"
)

type Weapon struct {
    BaseItem
    wearer      ItemWearer
    isTwoHanded bool
    weaponType  WeaponType
    material    WeaponMaterial
    level       ItemTier
}

func (a *Weapon) GetTooltipLines() []string {
    rows := []util.TableRow{
        {"Level", []string{string(a.level)}},
        {"Damage", []string{fmt.Sprintf("%d", a.GetBaseDamage())}},
        {"Value", []string{fmt.Sprintf("%dg", a.GetValue())}},
    }
    return util.TableLayout(rows)
}

func (a *Weapon) TintColor() color.Color {
    return a.level.Color()
}

func (a *Weapon) GetSlot() ItemSlot {
    return ItemSlotRightHand
}

func (a *Weapon) InventoryIcon() int32 {
    switch a.weaponType {
    case WeaponTypeGreatSword:
        return 186
    case WeaponTypeSword:
        return 164
    case WeaponTypeSpear:
        return 166
    case WeaponTypeStaff:
        return 167
    case WeaponTypeDagger:
        return 187
    case WeaponTypeBow:
        return 170
    case WeaponTypeCrossbow:
        return 171
    }
    return 164
}

func (a *Weapon) GetWearer() ItemWearer {
    return a.wearer
}
func (a *Weapon) GetValue() int {
    damageValue := max(a.GetBaseDamage(), 10)
    return damageValue * 10
}
func (a *Weapon) Icon(u uint64) int32 {
    return int32(205)
}
func (a *Weapon) GetContextActions(engine Engine) []renderer.MenuItem {
    equipAction := renderer.MenuItem{
        Text: "Equip",
        Action: func() {
            engine.ShowEquipMenu(a)
        },
    }
    baseActions := inventoryItemActions(a, engine)
    if !a.IsHeldByPlayer(engine) {
        return baseActions
    }
    equipActions := []renderer.MenuItem{equipAction}
    return append(equipActions, baseActions...)
}

func (a *Weapon) CanStackWith(other Item) bool {
    if otherWeapon, ok := other.(*Weapon); ok {
        return a.name == otherWeapon.name && a.GetBaseDamage() == otherWeapon.GetBaseDamage() && a.wearer == otherWeapon.wearer && a.weaponType == otherWeapon.weaponType && a.material == otherWeapon.material
    } else {
        return false
    }
}

func (a *Weapon) SetWearer(wearer ItemWearer) {
    a.wearer = wearer
}

func (a *Weapon) Unequip() {
    if a.wearer != nil {
        a.wearer.Unequip(a)
    }
}

func (a *Weapon) IsBetterThan(other Handheld) bool {
    if other == nil {
        return true
    }
    otherWeapon, isWeapon := other.(*Weapon)
    if !isWeapon {
        return true
    }
    return a.GetBaseDamage() > otherWeapon.GetBaseDamage()
}

func (a *Weapon) IsEquipped() bool {
    return a.wearer != nil
}

func (a *Weapon) IsTwoHanded() bool {
    return a.isTwoHanded
}

func (a *Weapon) GetDamage(actorBaseDamage int) int {
    return a.GetBaseDamage() + actorBaseDamage
}

type WeaponMaterial string

const (
    WeaponMaterialIron     WeaponMaterial = "iron"
    WeaponMaterialBronze   WeaponMaterial = "bronze"
    WeaponMaterialSteel    WeaponMaterial = "steel"
    WeaponMaterialDiamond  WeaponMaterial = "diamond"
    WeaponMaterialObsidian WeaponMaterial = "obsidian"
)

func getRandomMaterial(lootLevel int) WeaponMaterial {
    mod := rand.Intn(4) - 2
    weaponMaterial := min(5, max(1, lootLevel+mod))

    switch weaponMaterial {
    case 1:
        return WeaponMaterialIron
    case 2:
        return WeaponMaterialBronze
    case 3:
        return WeaponMaterialSteel
    case 4:
        return WeaponMaterialDiamond
    case 5:
        return WeaponMaterialObsidian
    }
    return WeaponMaterialIron
}
func getRandomWeaponType(lootLevel int) WeaponType {
    weaponLevel := rand.Intn(7) + 1

    switch weaponLevel {
    case 1:
        return WeaponTypeDagger
    case 2:
        return WeaponTypeSword
    case 3:
        return WeaponTypeStaff
    case 4:
        return WeaponTypeSpear
    case 5:
        return WeaponTypeGreatSword
    case 6:
        return WeaponTypeBow
    case 7:
        return WeaponTypeCrossbow
    }
    return WeaponTypeDagger
}

func NewRandomWeapon(lootLevel int) *Weapon {
    weaponType := getRandomWeaponType(lootLevel)
    material := getRandomMaterial(lootLevel)
    level := tierFromLootLevel(lootLevel)
    return NewWeapon(level, weaponType, material)
}

func (a *Weapon) IsRanged() bool {
    return a.weaponType == WeaponTypeBow || a.weaponType == WeaponTypeCrossbow
}
func (a *Weapon) Encode() string {
    return recfile.ToPredicate("weapon", string(a.level), string(a.weaponType), string(a.material), a.name)
}

func (a *Weapon) GetBaseDamage() int {
    return int(float64(a.damageByType(a.weaponType, a.level)) * a.materialMultiplier(a.material, a.level))
}

func (a *Weapon) damageByType(weaponType WeaponType, level ItemTier) int {
    switch weaponType {
    case WeaponTypeDagger:
        return int(float64(12) * level.Multiplier())
    case WeaponTypeSword:
        return int(float64(15) * level.Multiplier())
    case WeaponTypeStaff:
        return int(float64(8) * level.Multiplier())
    case WeaponTypeSpear:
        return int(float64(18) * level.Multiplier())
    case WeaponTypeGreatSword:
        return int(float64(25) * level.Multiplier())
    case WeaponTypeBow:
        return int(float64(12) * level.Multiplier())
    case WeaponTypeCrossbow:
        return int(float64(17) * level.Multiplier())
    }
    return 1
}

func (a *Weapon) materialMultiplier(material WeaponMaterial, level ItemTier) float64 {
    switch material {
    case WeaponMaterialIron:
        return 0.75 * level.Multiplier()
    case WeaponMaterialBronze:
        return float64(1) * level.Multiplier()
    case WeaponMaterialSteel:
        return 1.1 * level.Multiplier()
    case WeaponMaterialDiamond:
        return 1.5 * level.Multiplier()
    case WeaponMaterialObsidian:
        return float64(3) * level.Multiplier()
    }
    return 1
}

func (a *Weapon) IsHeldByPlayer(engine Engine) bool {
    if !a.IsHeld() {
        return false
    }
    return engine.IsPlayerControlled(a.GetHolder())
}
func NewWeaponFromPredicate(encoded recfile.StringPredicate) *Weapon {
    weapon := NewWeapon(
        ItemTier(encoded.GetString(0)),
        WeaponType(encoded.GetString(1)),
        WeaponMaterial(encoded.GetString(2)),
    )
    if encoded.ParamCount() > 3 {
        weapon.SetName(encoded.GetString(3))
    }
    return weapon
}
func NewWeapon(level ItemTier, weaponType WeaponType, material WeaponMaterial) *Weapon {
    weaponName := fmt.Sprintf("%s %s", material, weaponType)
    return &Weapon{
        weaponType: weaponType,
        material:   material,
        BaseItem: BaseItem{
            name: weaponName,
        },
        level: level,
    }
}
