package game

import (
    "Legacy/ega"
    "Legacy/geometry"
    "Legacy/renderer"
    "image/color"
)

type OngoingSpellEffect string

const (
    OngoingSpellEffectBirdsEye OngoingSpellEffect = "birds_eye"
)

func GetAllSpellScrolls() []*Scroll {
    names := []string{
        "Nom De Plume",
        "Create Food",
        "Bird's Eye",
        "Raise as StatusUndead",
        "Fireball",
        "Icebolt",
        "Healing word of Tauci",
    }
    var scrolls []*Scroll
    for _, name := range names {
        scrolls = append(scrolls, NewSpellScroll(name, "dummy", NewSpellFromName(name)))
    }

    return scrolls
}

func NewSpellFromName(name string) *Action {
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
    case "Healing word of Tauci":
        return NewSpell(name, 10, func(engine Engine, caster *Actor) {
            healthIncrease := 10 * caster.GetLevel()
            newHealth := min(caster.GetHealth()+healthIncrease, caster.GetMaxHealth())
            caster.SetHealth(newHealth)
            engine.Print("You feel better!")
        })
    case "Bird's Eye":
        return NewSpell(name, 10, func(engine Engine, caster *Actor) {
            engine.GetParty().AddSpellEffect(OngoingSpellEffectBirdsEye, 5)
        })
    case "Raise as StatusUndead":
        targetedSpell := NewTargetedSpell(name, 10, func(engine Engine, caster *Actor, pos geometry.Point) {
            engine.RaiseAsUndeadForParty(pos)
        })
        targetedSpell.SetActionColor(ega.BrightBlack)
        targetedSpell.SetValidTargets(func(engine Engine, caster *Actor, usePos geometry.Point) []geometry.Point {
            gridMap := engine.GetGridMap()
            potentialPositions := gridMap.GetDijkstraMapWithActorsNotBlocking(usePos, 4)
            var validPositions []geometry.Point
            for pos, _ := range potentialPositions {
                if geometry.Distance(caster.Pos(), pos) > 4 {
                    continue
                }
                isDownedActorAt := gridMap.IsDownedActorAt(pos)
                if isDownedActorAt {
                    downedActorAt := gridMap.DownedActorAt(pos)
                    if !downedActorAt.IsAlive() {
                        validPositions = append(validPositions, pos)
                    }
                }
            }

            return validPositions
        })
        targetedSpell.SetDescription([]string{
            "Raise a dead person as undead.",
            "Range: 4 tiles.",
        })
        return targetedSpell
    case "Fireball":
        fireball := NewTargetedSpell(name, 10, func(engine Engine, caster *Actor, pos geometry.Point) {
            radius := 3
            explosionIcon := int32(28)
            fireballDamagePerTile := 10 * caster.GetLevel()
            currentMap := engine.GetGridMap()
            hitPositions := currentMap.GetDijkstraMapWithActorsNotBlocking(pos, radius)
            for p, _ := range hitPositions {
                hitPos := p
                engine.CombatHitAnimation(hitPos, renderer.AtlasEntitiesGrayscale, explosionIcon, ega.BrightRed, func() {
                    engine.FixedDamageAt(caster, hitPos, fireballDamagePerTile)
                })
            }
        })
        fireball.SetActionColor(ega.BrightRed)
        return fireball
    case "Icebolt":
        icebolt := NewTargetedSpell(name, 10, func(engine Engine, caster *Actor, pos geometry.Point) {
            radius := 3
            explosionIcon := int32(28)
            iceboltDamage := 1 * caster.GetLevel()
            currentMap := engine.GetGridMap()
            hitPositions := currentMap.GetDijkstraMapWithActorsNotBlocking(pos, radius)
            for p, _ := range hitPositions {
                hitPos := p
                engine.CombatHitAnimation(hitPos, renderer.AtlasEntitiesGrayscale, explosionIcon, ega.BrightBlue, func() {
                    engine.FixedDamageAt(caster, hitPos, iceboltDamage)
                    engine.FreezeActorAt(hitPos, 3)
                })
            }
        })
        icebolt.SetActionColor(ega.BrightBlue)
        return icebolt
    }

    return nil
}

func NewSpell(name string, cost int, effect func(engine Engine, caster *Actor)) *Action {
    return &Action{
        name:   name,
        effect: effect,
        cost:   cost,
        canPayCost: func(engine Engine, user *Actor, costToPay int) bool {
            if !user.HasMana(costToPay) {
                engine.Print("Not enough mana!")
                return false
            }
            return true
        },
        payCost: func(engine Engine, user *Actor, costToPay int) {
            engine.ManaSpent(user, costToPay)
        },
        color: color.White,
    }
}

func NewTargetedSpell(name string, cost int, effect func(engine Engine, caster *Actor, pos geometry.Point)) *Action {
    return &Action{
        name:                 name,
        targetedEffect:       effect,
        cost:                 cost,
        closeModalsForEffect: true,
        canPayCost: func(engine Engine, user *Actor, costToPay int) bool {
            if !user.HasMana(costToPay) {
                engine.Print("Not enough mana!")
                return false
            }
            return true
        },
        payCost: func(engine Engine, user *Actor, costToPay int) {
            engine.ManaSpent(user, costToPay)
        },
        color: color.White,
    }
}
