package game

import (
    "Legacy/ega"
    "Legacy/geometry"
    "Legacy/renderer"
    "image/color"
)

func NewActiveSkillFromName(name SkillName) *BaseAction {
    switch name {
    case CombatSkillJellyJab:
        return NewMeleeAttackWithFixedDamage(string(name), 7)
    }
    return nil
}

func NewMeleeAttackWithFixedDamage(name string, damage int) *BaseAction {
    jab := NewTargetedCombatSkill(name, func(engine Engine, caster *Actor, pos geometry.Point) {
        bloodIcon := int32(104)
        engine.CombatHitAnimation(pos, renderer.AtlasWorld, bloodIcon, ega.BrightWhite, func() {
            engine.FixedDamageAt(caster, pos, damage)
        })
    })
    jab.SetValidTargets(func(engine Engine, caster *Actor, usePosition geometry.Point) map[geometry.Point]bool {
        gridMap := engine.GetGridMap()
        result := make(map[geometry.Point]bool)
        neighbors := gridMap.GetAllCardinalNeighbors(usePosition)
        for i := len(neighbors) - 1; i >= 0; i-- {
            pos := neighbors[i]
            if gridMap.IsActorAt(pos) {
                result[pos] = true
            }
        }
        return result
    })
    jab.SetDescription([]string{
        "Jab an adjacent enemy with your tentacle.",
    })
    jab.SetCombatUtilityForTargetedUseOnLocation(func(engine Engine, caster *Actor, target geometry.Point, allyPositions, enemyPositions map[geometry.Point]*Actor) int {
        if _, ok := enemyPositions[target]; ok {
            return damage * 10
        } else {
            return 0
        }
    })
    jab.SetActionColor(ega.BrightGreen)
    return jab
}
func NewCombatSkill(name string, effect func(engine Engine, user *Actor)) *BaseAction {
    return &BaseAction{
        name:   name,
        effect: effect,
        canPayCost: func(engine Engine, user *Actor) bool {
            return true
        },
        payCost: func(engine Engine, user *Actor) {
            //engine.ManaSpent(user, costToPay)
        },
        color: color.White,
    }
}

func NewTargetedCombatSkill(name string, effect func(engine Engine, user *Actor, pos geometry.Point)) *BaseAction {
    return &BaseAction{
        name:                 name,
        targetedEffect:       effect,
        closeModalsForEffect: true,
        canPayCost: func(engine Engine, user *Actor) bool {
            return true
        },
        payCost: func(engine Engine, user *Actor) {
            //engine.ManaSpent(user, costToPay)
        },
        color: color.White,
    }
}
