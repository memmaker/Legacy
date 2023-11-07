package game

import (
    "Legacy/recfile"
    "Legacy/renderer"
    "fmt"
    "math/rand"
    "strconv"
)

type WeaponType string

const (
    WeaponTypeSword      WeaponType = "sword"
    WeaponTypeGreatSword WeaponType = "great sword"
    WeaponTypeSpear      WeaponType = "spear"
    WeaponTypeStaff      WeaponType = "staff"
    WeaponTypeDagger     WeaponType = "dagger"
)

type Weapon struct {
    BaseItem
    baseDamage  int
    wearer      ItemWearer
    isTwoHanded bool
    weaponType  WeaponType
    material    WeaponMaterial
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
    }
    return 164
}

func (a *Weapon) GetWearer() ItemWearer {
    return a.wearer
}
func (a *Weapon) GetValue() int {
    return a.baseDamage * 10
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
    equipActions := []renderer.MenuItem{equipAction}
    baseActions := inventoryItemActions(a, engine)
    return append(equipActions, baseActions...)
}

func (a *Weapon) CanStackWith(other Item) bool {
    if otherWeapon, ok := other.(*Weapon); ok {
        return a.name == otherWeapon.name && a.baseDamage == otherWeapon.baseDamage && a.wearer == otherWeapon.wearer
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

func (a *Weapon) IsBetterThan(other *Weapon) bool {
    if other == nil {
        return true
    }
    return a.baseDamage > other.baseDamage
}

func (a *Weapon) IsEquipped() bool {
    return a.wearer != nil
}

func (a *Weapon) IsTwoHanded() bool {
    return a.isTwoHanded
}

func (a *Weapon) GetDamage(actorBaseDamage int) int {
    return a.baseDamage + actorBaseDamage
}

type WeaponMaterial string

func getRandomMaterial(lootLevel int) WeaponMaterial {
    mod := rand.Intn(4) - 2
    weaponLevel := min(5, max(1, lootLevel+mod))

    switch weaponLevel {
    case 1:
        return "wooden"
    case 2:
        return "iron"
    case 3:
        return "bronze"
    case 4:
        return "steel"
    case 5:
        return "diamond"
    }
    return "wooden"
}
func getRandomWeaponType(lootLevel int) WeaponType {
    weaponLevel := rand.Intn(5) + 1

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
    }
    return WeaponTypeDagger
}
func NewWeapon(weaponType WeaponType, material WeaponMaterial) *Weapon {
    weaponName := fmt.Sprintf("%s %s", material, weaponType)
    return &Weapon{
        weaponType: weaponType,
        material:   material,
        BaseItem: BaseItem{
            name: weaponName,
        },
        baseDamage: 1,
    }
}

func NewRandomWeapon(lootLevel int) *Weapon {
    weaponType := getRandomWeaponType(lootLevel)
    material := getRandomMaterial(lootLevel)

    return NewWeapon(weaponType, material)
}
func (a *Weapon) Encode() string {
    return recfile.ToPredicate("weapon", a.name, strconv.Itoa(a.baseDamage))
}
func NewWeaponFromPredicate(encoded recfile.StringPredicate) *Weapon {
    return NewWeapon(
        WeaponType(encoded.GetString(0)),
        WeaponMaterial(encoded.GetString(1)),
    )
}
