package game

type CreatureType string

const (
    CreatureTypeHuman          CreatureType = "human"
    CreatureTypeAnimal         CreatureType = "animal"
    CreatureTypeUndead         CreatureType = "undead"
    CreatureTypeNonIntelligent CreatureType = "non-intelligent"
)

type StatusEffect string

const (
    StatusEffectSleeping StatusEffect = "sleeping"
)
