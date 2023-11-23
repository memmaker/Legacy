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

// derived
const (
    Initiative        AttributeName = "Initiative"
    MovementAllowance AttributeName = "MovementAllowance"
    MeleeDamage       AttributeName = "MeleeDamage"
    MeleeHitChance    AttributeName = "MeleeHitChance"
    RangedDamage      AttributeName = "RangedDamage"
    RangedHitChance   AttributeName = "RangedHitChance"
    MeleeDefense      AttributeName = "MeleeDefense"
    RangedDefense     AttributeName = "RangedDefense"
    MagicDefense      AttributeName = "MagicDefense"
    MagicAffinity     AttributeName = "MagicAffinity"
)

type Modifier struct {
    Name     string
    Value    int
    IsActive func() bool
}
type AttributeHolder struct {
    attributes map[AttributeName]int
    modifiers  map[AttributeName][]Modifier
}

func NewAttributeHolder() AttributeHolder {
    return AttributeHolder{
        attributes: map[AttributeName]int{},
    }
}

func (a *AttributeHolder) GetAttributeBaseValue(name AttributeName) int {
    return a.attributes[name]
}

func (a *AttributeHolder) GetAttribute(name AttributeName) int {
    value := a.GetAttributeBaseValue(name)
    for i := len(a.modifiers[name]) - 1; i >= 0; i-- {
        mod := a.modifiers[name][i]
        if mod.IsActive == nil || !mod.IsActive() {
            a.modifiers[name] = append(a.modifiers[name][:i], a.modifiers[name][i+1:]...)
        } else {
            value += mod.Value
        }
    }
    return value
}

func (a *AttributeHolder) AddModifier(name AttributeName, modifierName string, value int, active func() bool) {
    a.modifiers[name] = append(a.modifiers[name], Modifier{
        Name:     modifierName,
        Value:    value,
        IsActive: active,
    })
}

func (a *AttributeHolder) GetMovementAllowance(armorEncumbrance int) int { // max 10
    agility := a.GetAttribute(Agility)
    return (agility - armorEncumbrance) + 1
}

func (a *AttributeHolder) GetInitiative() int { // 0-400
    agility := a.GetAttribute(Agility)
    perception := a.GetAttribute(Perception)
    level := a.GetAttribute(Level)
    return (agility + perception) * level
}

func (a *AttributeHolder) GetMeleeDamage() int { // 0-10
    strength := a.GetAttribute(Strength)
    return strength
}
