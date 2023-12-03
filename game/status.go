package game

import "fmt"

type CreatureType string

const (
    CreatureTypeHuman          CreatureType = "human"
    CreatureTypeAnimal         CreatureType = "animal"
    CreatureTypeOther          CreatureType = "other"
    CreatureTypeNonIntelligent CreatureType = "non-intelligent"
)

type StatusEffectName string

const (
    StatusEffectNameSleeping  StatusEffectName = "sleeping"
    StatusEffectNameUndead    StatusEffectName = "undead"
    StatusEffectNameWeak      StatusEffectName = "weak"
    StatusEffectNameHolyBonus StatusEffectName = "holy bonus"
    StatusEffectNameBlessed   StatusEffectName = "blessed"
)

// what do we really need?
// a list of distinct effects
// effects are always applied to an actor
// some only need checking, (isSleeping? isUndead?)
// some will need to apply damage (poison)
// some will raise stats (buffs)
type StatusEffect interface {
    IsExpired() bool
    Name() StatusEffectName
    OnRemove(engine Engine, actor *Actor)
    OnApply(engine Engine, actor *Actor)
    OnReapply(engine Engine, actor *Actor)
    Description() []string
}

type ModifierDefinition struct {
    Attribute AttributeName
    GetValue  func() int
}

type DerivedModifierDefinition struct {
    Attribute DerivedAttributeName
    GetValue  func() int
}

type AttributeModifier interface {
    StatusEffect
    GetModifiers() []ModifierDefinition
}

type DerivedAttributeModifier interface {
    StatusEffect
    GetDerivedModifiers() []DerivedModifierDefinition
}

type SkillModifier interface {
    StatusEffect
    GetSkill() SkillName
    GetSkillModifier() int
}

type RealTimeEffect interface {
    StatusEffect
    OnTick(engine Engine, actor *Actor, tick uint64)
}

type CombatEffect interface {
    StatusEffect
    OnNewTurn(engine Engine, actor *Actor)
}

type OnDamageEffect interface {
    StatusEffect
    OnDamageReceived(engine Engine, actor *Actor, amount int)
}

type OnRestEffect interface {
    StatusEffect
    OnRest(engine Engine, actor *Actor)
}

type BaseStatusEffect struct {
    isOver bool
}

func (e *BaseStatusEffect) OnRemove(engine Engine, actor *Actor)  {}
func (e *BaseStatusEffect) OnApply(engine Engine, actor *Actor)   {}
func (e *BaseStatusEffect) OnReapply(engine Engine, actor *Actor) {}
func (e *BaseStatusEffect) IsExpired() bool                       { return e.isOver }

type SleepingEffect struct {
    BaseStatusEffect
}

func (e *SleepingEffect) Description() []string {
    return []string{"Cannot move or act"}
}

func (e *SleepingEffect) OnDamageReceived(engine Engine, actor *Actor, amount int) {
    e.isOver = true
}

func (e *SleepingEffect) OnRest(engine Engine, actor *Actor) {
    e.isOver = true
}

func (e *SleepingEffect) Name() StatusEffectName {
    return StatusEffectNameSleeping
}

func StatusSleeping() *SleepingEffect {
    return &SleepingEffect{}
}

type UndeadEffect struct {
    BaseStatusEffect
}

func (e *UndeadEffect) Description() []string {
    return []string{"Is not living or breathing."}
}

func (e *UndeadEffect) Name() StatusEffectName {
    return StatusEffectNameUndead
}

func StatusUndead() *UndeadEffect {
    return &UndeadEffect{}
}

type WeakEffect struct {
    BaseStatusEffect
}

func (e *WeakEffect) Description() []string {
    return []string{"-2 Strength", "-2 Agility"}
}

func (e *WeakEffect) Name() StatusEffectName {
    return StatusEffectNameWeak
}

func StatusWeak() *WeakEffect {
    return &WeakEffect{}
}

func (e *WeakEffect) OnRest(engine Engine, actor *Actor) {
    e.isOver = true
}

func (e *WeakEffect) GetModifiers() []ModifierDefinition {
    return []ModifierDefinition{
        {Attribute: Strength, GetValue: func() int { return -2 }},
        {Attribute: Agility, GetValue: func() int { return -2 }},
    }
}

type HolyBonusEffect struct {
    BaseStatusEffect
    stacks int
}

func (e *HolyBonusEffect) Description() []string {
    return []string{fmt.Sprintf("+%d base melee damage", e.stacks)}
}

func (e *HolyBonusEffect) Name() StatusEffectName {
    return StatusEffectNameHolyBonus
}

func StatusHolyBonus() *HolyBonusEffect {
    return &HolyBonusEffect{}
}

func (e *HolyBonusEffect) OnApply(engine Engine, actor *Actor) {
    e.stacks++
}

func (e *HolyBonusEffect) OnReapply(engine Engine, actor *Actor) {
    e.stacks++
}

func (e *HolyBonusEffect) OnRest(engine Engine, actor *Actor) {
    e.isOver = true
}

func (e *HolyBonusEffect) GetDerivedModifiers() []DerivedModifierDefinition {
    return []DerivedModifierDefinition{
        {Attribute: DerivedAttributeBaseMeleeDamage, GetValue: func() int { return e.stacks }},
    }
}

type BlessedEffect struct {
    BaseStatusEffect
    stacks int
}

func (e *BlessedEffect) Description() []string {
    return []string{fmt.Sprintf("+%d base armor", e.stacks)}
}

func (e *BlessedEffect) Name() StatusEffectName {
    return StatusEffectNameBlessed
}

func StatusBlessed() *BlessedEffect {
    return &BlessedEffect{}
}

func (e *BlessedEffect) OnApply(engine Engine, actor *Actor) {
    e.stacks++
}

func (e *BlessedEffect) OnReapply(engine Engine, actor *Actor) {
    e.stacks++
}

func (e *BlessedEffect) OnRest(engine Engine, actor *Actor) {
    e.isOver = true
}

func (e *BlessedEffect) GetDerivedModifiers() []DerivedModifierDefinition {
    return []DerivedModifierDefinition{
        {Attribute: DerivedAttributeBaseArmor, GetValue: func() int { return e.stacks }},
    }
}

func StatusFromName(name string) StatusEffect {
    effectName := StatusEffectName(name)
    switch effectName {
    case StatusEffectNameSleeping:
        return StatusSleeping()
    case StatusEffectNameUndead:
        return StatusUndead()
    case StatusEffectNameWeak:
        return StatusWeak()
    case StatusEffectNameHolyBonus:
        return StatusHolyBonus()
    case StatusEffectNameBlessed:
        return StatusBlessed()
    }
    return nil
}
