package game

import "Legacy/geometry"

func NewSpellFromName(name string) *Spell {
    switch name {
    case "Create Food":
        return NewSpell(name, 10, func(engine Engine) {
            engine.AddFood(10)
        })
    }
    return nil
}

type Spell struct {
    manaCost       int
    effect         func(engine Engine)
    targetedEffect func(engine Engine, pos geometry.Point)
    name           string
}

func NewSpell(name string, cost int, effect func(engine Engine)) *Spell {
    return &Spell{
        name:     name,
        effect:   effect,
        manaCost: cost,
    }
}

func NewTargetedSpell(name string, cost int, effect func(engine Engine, pos geometry.Point)) *Spell {
    return &Spell{
        name:           name,
        targetedEffect: effect,
        manaCost:       cost,
    }
}

func (s *Spell) Cast(engine Engine, caster *Actor) {
    if !caster.HasMana(s.manaCost) {
        engine.Print("Not enough mana!")
        return
    }
    engine.ManaSpent(caster, s.manaCost)
    if s.effect != nil {
        s.effect(engine)
    }
}

func (s *Spell) CastOnTarget(engine Engine, caster *Actor, pos geometry.Point) {
    if !caster.HasMana(s.manaCost) {
        engine.Print("Not enough mana!")
        return
    }
    caster.RemoveMana(s.manaCost)
    if s.targetedEffect != nil {
        s.targetedEffect(engine, pos)
    }
}
func (s *Spell) IsTargeted() bool {
    return s.targetedEffect != nil && s.effect == nil
}

func (s *Spell) ManaCost() int {
    return s.manaCost
}
func (s *Spell) Name() string {
    return s.name
}
