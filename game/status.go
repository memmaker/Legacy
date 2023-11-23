package game

type CreatureType string

const (
    CreatureTypeHuman          CreatureType = "human"
    CreatureTypeAnimal         CreatureType = "animal"
    CreatureTypeUndead         CreatureType = "undead"
    CreatureTypeNonIntelligent CreatureType = "non-intelligent"
)

type StatusEffect string

func (e StatusEffect) IsRemovedOnDamage() bool {
    switch e {
    case StatusEffectSleeping:
        return true
    }
    return false
}

const (
    StatusEffectSleeping StatusEffect = "sleeping"
)
