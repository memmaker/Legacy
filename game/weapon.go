package game

import (
    "Legacy/recfile"
    "Legacy/renderer"
    "fmt"
    "math/rand"
    "strconv"
)

type Weapon struct {
    BaseItem
    baseDamage  int
    wearer      ItemWearer
    isTwoHanded bool
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
    if otherArmor, ok := other.(*Armor); ok {
        return a.name == otherArmor.name && a.baseDamage == otherArmor.protection && a.wearer == otherArmor.wearer
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

func getRandomMaterial(lootLevel int) (string, int) {
    mod := rand.Intn(4) - 2
    weaponLevel := min(5, max(1, lootLevel+mod))

    switch weaponLevel {
    case 1:
        return "wooden", 2
    case 2:
        return "iron", 5
    case 3:
        return "bronze", 7
    case 4:
        return "steel", 12
    case 5:
        return "diamond", 20
    }
    return "wooden", 1
}
func getRandomWeaponType(lootLevel int) (string, int) {
    weaponLevel := rand.Intn(5) + 1

    switch weaponLevel {
    case 1:
        return "dagger", 2
    case 2:
        return "sword", 5
    case 3:
        return "spear", 7
    case 4:
        return "hammer", 12
    case 5:
        return "greatsword", 20
    }
    return "dagger", 1
}
func NewWeapon(name string, damage int) *Weapon {
    return &Weapon{
        BaseItem: BaseItem{
            name: name,
        },
        baseDamage: damage,
    }
}

func NewRandomWeapon(lootLevel int) *Weapon {
    weaponType, baseDamage := getRandomWeaponType(lootLevel)
    material, damageModifier := getRandomMaterial(lootLevel)
    totalDamage := baseDamage + damageModifier
    weaponName := fmt.Sprintf("%s %s", material, weaponType)
    return NewWeapon(weaponName, totalDamage)
}
func (a *Weapon) Encode() string {
    return recfile.ToPredicate("weapon", a.name, strconv.Itoa(a.baseDamage))
}
func NewWeaponFromPredicate(encoded recfile.StringPredicate) *Weapon {
    return NewWeapon(
        encoded.GetString(0),
        encoded.GetInt(1),
    )
}
