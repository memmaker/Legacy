package game

type AttributeName string

const (
    Strength     AttributeName = "Strength"
    Perception   AttributeName = "Perception"
    Endurance    AttributeName = "Endurance"
    Charisma     AttributeName = "Charisma"
    Intelligence AttributeName = "Intelligence"
    Agility      AttributeName = "Agility"
    Luck         AttributeName = "Luck"
    Level        AttributeName = "Level"
)

func GetAllAttributeNames() []AttributeName {
    return []AttributeName{
        Level,
        Strength,
        Perception,
        Endurance,
        Charisma,
        Intelligence,
        Agility,
        Luck,
    }
}

func GetAllDerivedAttributeNames() []DerivedAttributeName {
    return []DerivedAttributeName{
        DerivedAttributeMovementAllowance,
        DerivedAttributeInitiative,
        DerivedAttributeBaseArmor,
        DerivedAttributeBaseMeleeDamage,
        DerivedAttributeBaseRangedDamage,
    }
}

type Modifier struct {
    Name      string
    Value     func() int
    IsExpired func() bool
}
type DerivedAttributeName string

const (
    DerivedAttributeMovementAllowance DerivedAttributeName = "MovementAllowance"
    DerivedAttributeInitiative        DerivedAttributeName = "Initiative"
    DerivedAttributeBaseArmor         DerivedAttributeName = "BaseArmor"
    DerivedAttributeBaseMeleeDamage   DerivedAttributeName = "BaseMeleeDamage"
    DerivedAttributeBaseRangedDamage  DerivedAttributeName = "BaseRangedDamage"
)

type AttributeHolder struct {
    attributes map[AttributeName]int
    modifiers  map[AttributeName][]Modifier

    // derived
    derivedModifiers map[DerivedAttributeName][]Modifier
}

func NewAttributeHolder() AttributeHolder {
    return AttributeHolder{
        attributes: map[AttributeName]int{
            Level:        1,
            Strength:     5,
            Perception:   5,
            Endurance:    5,
            Charisma:     5,
            Intelligence: 5,
            Agility:      5,
            Luck:         5,
        },
        modifiers:        make(map[AttributeName][]Modifier),
        derivedModifiers: make(map[DerivedAttributeName][]Modifier),
    }
}

func (a *AttributeHolder) GetAttributeBaseValue(name AttributeName) int {
    return a.attributes[name]
}

func (a *AttributeHolder) GetAttribute(name AttributeName) int {
    value := a.GetAttributeBaseValue(name) + a.getTotalModifier(name)
    return value
}

func (a *AttributeHolder) getTotalModifier(name AttributeName) int {
    modifierSum := 0
    for i := len(a.modifiers[name]) - 1; i >= 0; i-- {
        mod := a.modifiers[name][i]
        if mod.IsExpired == nil || mod.IsExpired() {
            a.modifiers[name] = append(a.modifiers[name][:i], a.modifiers[name][i+1:]...)
        } else {
            modifierSum += mod.Value()
        }
    }
    return modifierSum
}

func (a *AttributeHolder) getTotalDerivedModifier(name DerivedAttributeName) int {
    modifierSum := 0
    for i := len(a.derivedModifiers[name]) - 1; i >= 0; i-- {
        mod := a.derivedModifiers[name][i]
        if mod.IsExpired == nil || mod.IsExpired() {
            a.derivedModifiers[name] = append(a.derivedModifiers[name][:i], a.derivedModifiers[name][i+1:]...)
        } else {
            modifierSum += mod.Value()
        }
    }
    return modifierSum
}

func (a *AttributeHolder) GetAllModifiersForAttribute(name AttributeName) []Modifier {
    modifiers := a.modifiers[name]
    for i := len(modifiers) - 1; i >= 0; i-- {
        mod := modifiers[i]
        if mod.IsExpired == nil || mod.IsExpired() {
            modifiers = append(modifiers[:i], modifiers[i+1:]...)
        }
    }
    a.modifiers[name] = modifiers
    return modifiers
}

func (a *AttributeHolder) AddModifier(name AttributeName, modifierName string, value func() int, isExpired func() bool) {
    a.modifiers[name] = append(a.modifiers[name], Modifier{
        Name:      modifierName,
        Value:     value,
        IsExpired: isExpired,
    })
}

func (a *AttributeHolder) AddDerivedModifier(name DerivedAttributeName, modifierName string, value func() int, isExpired func() bool) {
    a.derivedModifiers[name] = append(a.derivedModifiers[name], Modifier{
        Name:      modifierName,
        Value:     value,
        IsExpired: isExpired,
    })
}

func (a *AttributeHolder) GetBaseArmor() int {
    agility := a.GetAttribute(Agility)
    modifier := a.getTotalDerivedModifier(DerivedAttributeBaseArmor)
    return agility + modifier
}

func (a *AttributeHolder) GetBaseMeleeDamage() int {
    strength := a.GetAttribute(Strength)
    modifier := a.getTotalDerivedModifier(DerivedAttributeBaseMeleeDamage)
    return strength + modifier
}

func (a *AttributeHolder) GetBaseRangedDamage() int {
    perception := a.GetAttribute(Perception)
    modifier := a.getTotalDerivedModifier(DerivedAttributeBaseRangedDamage)
    return perception + modifier
}
func (a *AttributeHolder) GetMovementAllowance(armorEncumbrance int) int { // max 10
    agilityFactor := int(float64(a.GetAttribute(Agility)) * 1.5) // 2-15
    modifier := a.getTotalDerivedModifier(DerivedAttributeMovementAllowance)
    // 2..15 - 1..7 = 1..8
    baseValue := (agilityFactor - armorEncumbrance) + modifier
    return max(1, baseValue+3)
}

func (a *AttributeHolder) GetInitiative() int { // 0-400
    agility := a.GetAttribute(Agility)
    perception := a.GetAttribute(Perception)
    level := a.GetAttribute(Level)
    modifier := a.getTotalDerivedModifier(DerivedAttributeInitiative)
    return ((agility + perception) * level) + modifier
}

func (a *AttributeHolder) Increment(attributeName AttributeName) {
    a.attributes[attributeName]++
}

func (a *AttributeHolder) Decrement(name AttributeName) {
    a.attributes[name]--
}

func (a *AttributeHolder) SetAttribute(attributeName AttributeName, value int) {
    a.attributes[attributeName] = value
}
