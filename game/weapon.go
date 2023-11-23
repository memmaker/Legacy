package game

import (
    "Legacy/recfile"
    "Legacy/util"
    "fmt"
    "image/color"
    "math/rand"
)

type WeaponType string

const (
    WeaponTypeSword      WeaponType = "sword"
    WeaponTypeShield     WeaponType = "shield"
    WeaponTypeGreatSword WeaponType = "great sword"
    WeaponTypeSpear      WeaponType = "spear"
    WeaponTypeStaff      WeaponType = "staff"
    WeaponTypeMace       WeaponType = "mace"
    WeaponTypeDagger     WeaponType = "dagger"
    WeaponTypeAxe        WeaponType = "axe"
    WeaponTypeBow        WeaponType = "bow"
    WeaponTypeCrossbow   WeaponType = "crossbow"
)

func GetAllWeaponTypes() []WeaponType {
    return []WeaponType{
        WeaponTypeSword,
        WeaponTypeGreatSword,
        WeaponTypeSpear,
        WeaponTypeShield,
        WeaponTypeStaff,
        WeaponTypeMace,
        WeaponTypeAxe,
        WeaponTypeDagger,
        WeaponTypeBow,
        WeaponTypeCrossbow,
    }
}

type OnHitEffect struct {
    condition    func(engine Engine, weapon *Weapon, attacker, victim *Actor) bool
    apply        func(engine Engine, weapon *Weapon, attacker, victim *Actor)
    toolTipLeft  string
    toolTipRight string
}

func (e OnHitEffect) toolTipDescription() util.TableRow {
    return util.TableRow{Label: " " + e.toolTipLeft, Columns: []string{e.toolTipRight}}
}

func NewOnHitEffect(name string) OnHitEffect {
    effect := OnHitEffect{}
    switch name {
    case "slime whisperer":
        effect.condition = func(engine Engine, weapon *Weapon, attacker, victim *Actor) bool {
            return victim.GetCombatFaction() == "slime"
        }
        effect.apply = func(engine Engine, weapon *Weapon, attacker, victim *Actor) {
            victim.AddStatusEffect(StatusEffectSleeping, 5)
        }
        effect.toolTipLeft = "slime"
        effect.toolTipRight = "sleeping (100%)"
    case "greed":
        effect.condition = func(engine Engine, weapon *Weapon, attacker, victim *Actor) bool {
            return rand.Float64() < 0.33 && !victim.IsAlive()
        }
        effect.apply = func(engine Engine, weapon *Weapon, attacker, victim *Actor) {
            amount := rand.Intn(victim.level*300) + 100
            victim.AddGold(amount)
        }
        effect.toolTipLeft = "greed"
        effect.toolTipRight = "loot more gold (33%)"
    }
    return effect
}

type Weapon struct {
    BaseItem
    wearer         ItemWearer
    isTwoHanded    bool
    weaponType     WeaponType
    material       WeaponMaterial
    level          ItemTier
    onHitEffects   []OnHitEffect
    useFixedDamage bool
    fixedDamage    int
}

func (a *Weapon) AddOnHitEffect(effect OnHitEffect) {
    a.onHitEffects = append(a.onHitEffects, effect)
}

func (a *Weapon) AddOnHitEffectByName(effectName string) {
    a.AddOnHitEffect(NewOnHitEffect(effectName))
}
func (a *Weapon) GetTooltipLines() []string {
    rows := []util.TableRow{
        {"Level", []string{string(a.level)}},
        {"Type", []string{string(a.weaponType)}},
        {"Material", []string{string(a.material)}},
        {"Damage", []string{fmt.Sprintf("%d", a.GetBaseDamage())}},
        {"Value", []string{fmt.Sprintf("%dg", a.GetValue())}},
    }
    if len(a.onHitEffects) > 0 {
        rows = append(rows, util.TableRow{})
        rows = append(rows, util.TableRow{Label: "On Hit", Columns: []string{}})
        for _, effect := range a.onHitEffects {
            rows = append(rows, effect.toolTipDescription())
        }
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
    levelMultiplier := a.level.Multiplier()
    materialMultiplier := a.materialMultiplier(a.material)
    var baseValue int
    switch a.weaponType {
    case WeaponTypeGreatSword:
        baseValue = 10000
    case WeaponTypeSword:
        baseValue = 5000
    case WeaponTypeShield:
        baseValue = 2000
    case WeaponTypeSpear:
        baseValue = 900
    case WeaponTypeStaff:
        baseValue = 200
    case WeaponTypeMace:
        baseValue = 800
    case WeaponTypeAxe:
        baseValue = 500
    case WeaponTypeDagger:
        baseValue = 400
    case WeaponTypeBow:
        baseValue = 1000
    case WeaponTypeCrossbow:
        baseValue = 4000
    }
    return int(float64(baseValue) * (levelMultiplier * materialMultiplier))
}

func (a *Weapon) Name() string {
    if a.name != "" {
        return a.name
    }
    return fmt.Sprintf("%s %s", a.material, a.weaponType)
}
func (a *Weapon) Icon(u uint64) int32 {
    if a.weaponType == WeaponTypeBow || a.weaponType == WeaponTypeCrossbow {
        return int32(223)
    }
    return int32(220)
}
func (a *Weapon) GetContextActions(engine Engine) []util.MenuItem {
    baseActions := inventoryItemActions(a, engine)
    if !a.IsHeldByPlayer(engine) {
        return baseActions
    }

    if a.IsEquipped() {
        return append([]util.MenuItem{{
            Text: "Unequip",
            Action: func() {
                a.Unequip()
            }}}, baseActions...)
    } else {
        return append([]util.MenuItem{{
            Text: "Equip",
            Action: func() {
                engine.ShowEquipMenu(a)
            }}}, baseActions...)
    }
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
func GetAllWeaponMaterials() []WeaponMaterial {
    return []WeaponMaterial{
        WeaponMaterialIron,
        WeaponMaterialBronze,
        WeaponMaterialSteel,
        WeaponMaterialGold,
        WeaponMaterialDiamond,
        WeaponMaterialObsidian,
    }
}

type WeaponMaterial string

const (
    WeaponMaterialIron     WeaponMaterial = "iron"
    WeaponMaterialBronze   WeaponMaterial = "bronze"
    WeaponMaterialSteel    WeaponMaterial = "steel"
    WeaponMaterialGold     WeaponMaterial = "gold"
    WeaponMaterialDiamond  WeaponMaterial = "diamond"
    WeaponMaterialObsidian WeaponMaterial = "obsidian"
)

func getRandomMaterial(lootLevel int) WeaponMaterial {
    mod := rand.Intn(4) - 2
    weaponMaterial := min(6, max(1, lootLevel+mod))

    switch weaponMaterial {
    case 1:
        return WeaponMaterialIron
    case 2:
        return WeaponMaterialBronze
    case 3:
        return WeaponMaterialSteel
    case 4:
        return WeaponMaterialGold
    case 5:
        return WeaponMaterialDiamond
    case 6:
        return WeaponMaterialObsidian
    }
    return WeaponMaterialIron
}
func getRandomWeaponType() WeaponType {
    allWeaponTypes := GetAllWeaponTypes()
    randomIndex := rand.Intn(len(allWeaponTypes))
    return allWeaponTypes[randomIndex]
}

func NewRandomWeapon(lootLevel int) *Weapon {
    weaponType := getRandomWeaponType()
    material := getRandomMaterial(lootLevel)
    level := tierFromLootLevel(lootLevel)
    return NewWeapon(level, weaponType, material)
}

func NewRandomWeaponForVendor(lootLevel int) *Weapon {
    weaponType := getRandomWeaponType()
    material := getRandomMaterial(lootLevel)
    level := "common"
    if lootLevel > 1 {
        level = "uncommon"
    }
    return NewWeapon(ItemTier(level), weaponType, material)
}

func (a *Weapon) IsRanged() bool {
    return a.weaponType == WeaponTypeBow || a.weaponType == WeaponTypeCrossbow
}
func (a *Weapon) Encode() string {
    return recfile.ToPredicate("weapon", string(a.level), string(a.weaponType), string(a.material), a.name)
}

func (a *Weapon) GetBaseDamage() int {
    if a.useFixedDamage {
        return a.fixedDamage
    }
    return int(float64(a.damageByType(a.weaponType, a.level)) * a.materialMultiplier(a.material))
}

func (a *Weapon) damageByType(weaponType WeaponType, level ItemTier) int {
    switch weaponType {
    case WeaponTypeDagger:
        return int(float64(17) * level.Multiplier())
    case WeaponTypeSword:
        return int(float64(19) * level.Multiplier())
    case WeaponTypeStaff:
        return int(float64(15) * level.Multiplier())
    case WeaponTypeMace:
        return int(float64(17) * level.Multiplier())
    case WeaponTypeSpear:
        return int(float64(18) * level.Multiplier())
    case WeaponTypeGreatSword:
        return int(float64(20) * level.Multiplier())
    case WeaponTypeBow:
        return int(float64(15) * level.Multiplier())
    case WeaponTypeCrossbow:
        return int(float64(17) * level.Multiplier())
    }
    return 1
}

func (a *Weapon) materialMultiplier(material WeaponMaterial) float64 {
    switch material {
    case WeaponMaterialIron:
        return 0.75
    case WeaponMaterialBronze:
        return float64(1)
    case WeaponMaterialSteel:
        return 1.1
    case WeaponMaterialGold:
        return 1.25
    case WeaponMaterialDiamond:
        return 1.5
    case WeaponMaterialObsidian:
        return float64(3)
    }
    return 1
}

func (a *Weapon) IsHeldByPlayer(engine Engine) bool {
    if !a.IsHeld() {
        return false
    }
    return engine.IsPlayerControlled(a.GetHolder())
}

// goal:
// we want a weapon, that has a 100% chance to apply the sleeping status effect
// to a creature, if the combat faction is set to "slime"
// onHit:
func (a *Weapon) OnHitProc(engine Engine, attacker, victim *Actor) {
    for _, effect := range a.onHitEffects {
        if effect.condition(engine, a, attacker, victim) {
            effect.apply(engine, a, attacker, victim)
        }
    }
}

func (a *Weapon) SetFixedDamage(newValue int) {
    a.useFixedDamage = true
    a.fixedDamage = newValue
}

func (a *Weapon) IsDagger() bool {
    return a.weaponType == WeaponTypeDagger
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
    return &Weapon{
        weaponType: weaponType,
        material:   material,
        BaseItem: BaseItem{
            name: "",
        },
        level: level,
    }
}

func NewNamedWeapon(weaponName string) *Weapon {
    switch weaponName {
    case "slime whisperer":
        weapon := NewWeapon(ItemTierCommon, WeaponTypeBow, WeaponMaterialIron)
        weapon.SetFixedDamage(0)
        weapon.SetName("slime whisperer")
        weapon.AddOnHitEffectByName("slime whisperer")
        return weapon
    case "robbers_bow":
        weapon := NewWeapon(ItemTierCommon, WeaponTypeBow, WeaponMaterialIron)
        weapon.SetName("robber's bow")
        weapon.AddOnHitEffectByName("greed")
        return weapon
    case "robbers_dagger":
        weapon := NewWeapon(ItemTierCommon, WeaponTypeDagger, WeaponMaterialIron)
        weapon.SetName("robber's dagger")
        weapon.AddOnHitEffectByName("greed")
        return weapon
    }
    println("ERR: unknown weapon name:", weaponName)
    return nil
}
