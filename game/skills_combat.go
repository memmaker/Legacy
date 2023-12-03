package game

import (
    "Legacy/ega"
    "Legacy/geometry"
    "Legacy/renderer"
    "image/color"
)

func NewActiveSkillFromName(name string) *Action {
    switch name {
    case "Jelly Jab":
        return NewMeleeAttackWithFixedDamage(name, 7)
    }
    return nil
}

func NewMeleeAttackWithFixedDamage(name string, damage int) *Action {
    jab := NewTargetedCombatSkill(name, 0, func(engine Engine, caster *Actor, pos geometry.Point) {
        bloodIcon := int32(104)
        engine.CombatHitAnimation(pos, renderer.AtlasWorld, bloodIcon, ega.BrightWhite, func() {
            engine.FixedDamageAt(caster, pos, damage)
        })
    })
    jab.SetValidTargets(func(engine Engine, caster *Actor, usePosition geometry.Point) []geometry.Point {
        gridMap := engine.GetGridMap()
        neighbors := gridMap.GetAllCardinalNeighbors(usePosition)
        for i := len(neighbors) - 1; i >= 0; i-- {
            if !gridMap.IsActorAt(neighbors[i]) {
                neighbors = append(neighbors[:i], neighbors[i+1:]...)
            }
        }
        return neighbors
    })
    jab.SetDescription([]string{
        "Jab an adjacent enemy with your tentacle.",
    })
    jab.SetCombatUtilityForTargetedUseOnLocation(func(target geometry.Point, enemyPositions map[geometry.Point]*Actor) int {
        if _, ok := enemyPositions[target]; ok {
            return damage * 10
        } else {
            return 0
        }
    })
    jab.SetActionColor(ega.BrightGreen)
    return jab
}
func NewCombatSkill(name string, cost int, effect func(engine Engine, user *Actor)) *Action {
    return &Action{
        name:   name,
        effect: effect,
        cost:   cost,
        canPayCost: func(engine Engine, user *Actor, costToPay int) bool {
            return true
        },
        payCost: func(engine Engine, user *Actor, costToPay int) {
            //engine.ManaSpent(user, costToPay)
        },
        color: color.White,
    }
}

func NewTargetedCombatSkill(name string, cost int, effect func(engine Engine, user *Actor, pos geometry.Point)) *Action {
    return &Action{
        name:                 name,
        targetedEffect:       effect,
        cost:                 cost,
        closeModalsForEffect: true,
        canPayCost: func(engine Engine, user *Actor, costToPay int) bool {
            return true
        },
        payCost: func(engine Engine, user *Actor, costToPay int) {
            //engine.ManaSpent(user, costToPay)
        },
        color: color.White,
    }
}
