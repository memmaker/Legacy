package game

import (
    "Legacy/ega"
    "Legacy/geometry"
    "Legacy/renderer"
    "image/color"
)

func NewSpellFromName(name string) *Spell {
    switch name {
    case "Nom De Plume":
        return NewSpell(name, 0, func(engine Engine, caster *Actor) {
            engine.AskUserForString("Now known as: ", 8, func(text string) {
                caster.SetName(text)
            })
        })
    case "Create Food":
        return NewSpell(name, 10, func(engine Engine, caster *Actor) {
            engine.AddFood(10)
        })
    case "Raise as Undead":
        targetedSpell := NewTargetedSpell(name, 10, func(engine Engine, caster *Actor, pos geometry.Point) {
            engine.RaiseAsUndeadForParty(pos)
        })
        targetedSpell.SetSpellColor(ega.BrightBlack)
        return targetedSpell
    case "Fireball":
        fireball := NewTargetedSpell(name, 10, func(engine Engine, caster *Actor, pos geometry.Point) {
            radius := 3
            explosionIcon := int32(28)
            fireballDamagePerTile := 10 * caster.GetLevel()
            hitPositions := engine.GetAoECircle(pos, radius)
            for _, p := range hitPositions {
                hitPos := p
                engine.HitAnimation(hitPos, renderer.AtlasEntitiesGrayscale, explosionIcon, ega.BrightRed, func() {
                    engine.SpellDamageAt(caster, hitPos, fireballDamagePerTile)
                })
            }
        })
        fireball.SetSpellColor(ega.BrightRed)
        return fireball
    case "Icebolt":
        icebolt := NewTargetedSpell(name, 10, func(engine Engine, caster *Actor, pos geometry.Point) {
            radius := 3
            explosionIcon := int32(28)
            iceboltDamage := 1 * caster.GetLevel()
            hitPositions := engine.GetAoECircle(pos, radius)
            for _, p := range hitPositions {
                hitPos := p
                engine.HitAnimation(hitPos, renderer.AtlasEntitiesGrayscale, explosionIcon, ega.BrightBlue, func() {
                    engine.SpellDamageAt(caster, hitPos, iceboltDamage)
                    engine.FreezeActorAt(hitPos, 3)
                })
            }
        })
        icebolt.SetSpellColor(ega.BrightBlue)
        return icebolt
    }

    return nil
}

type Spell struct {
    manaCost             int
    effect               func(engine Engine, caster *Actor)
    targetedEffect       func(engine Engine, caster *Actor, pos geometry.Point)
    name                 string
    color                color.Color
    closeModalsForEffect bool
}

func (s *Spell) SetSpellColor(c color.Color) {
    s.color = c
}

func NewSpell(name string, cost int, effect func(engine Engine, caster *Actor)) *Spell {
    return &Spell{
        name:     name,
        effect:   effect,
        manaCost: cost,
    }
}

func NewTargetedSpell(name string, cost int, effect func(engine Engine, caster *Actor, pos geometry.Point)) *Spell {
    return &Spell{
        name:                 name,
        targetedEffect:       effect,
        manaCost:             cost,
        closeModalsForEffect: true,
    }
}

func (s *Spell) Cast(engine Engine, caster *Actor) {
    if !caster.HasMana(s.manaCost) {
        engine.Print("Not enough mana!")
        return
    }
    engine.ManaSpent(caster, s.manaCost)
    if s.effect != nil {
        if s.closeModalsForEffect {
            engine.CloseAllModals()
        }
        s.effect(engine, caster)
    }
}

func (s *Spell) CastOnTarget(engine Engine, caster *Actor, pos geometry.Point) {
    if !caster.HasMana(s.manaCost) {
        engine.Print("Not enough mana!")
        return
    }
    caster.RemoveMana(s.manaCost)
    if s.targetedEffect != nil {
        if s.closeModalsForEffect {
            engine.CloseAllModals()
        }
        s.targetedEffect(engine, caster, pos)
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

func (s *Spell) GetValue() int {
    return s.manaCost * 10
}

func (s *Spell) Color() color.Color {
    return s.color
}
