package game

import (
    "Legacy/geometry"
    "image/color"
)

type Action struct {
    cost                 int
    effect               func(engine Engine, user *Actor)
    targetedEffect       func(engine Engine, user *Actor, pos geometry.Point)
    name                 string
    color                color.Color
    closeModalsForEffect bool

    canPayCost        func(engine Engine, user *Actor, cost int) bool
    payCost           func(engine Engine, user *Actor, cost int)
    getValidPositions func(engine Engine, user *Actor, usePosition geometry.Point) []geometry.Point
    description       []string

    // AI stuff
    targetedCombatUtility func(target geometry.Point, enemyPositions map[geometry.Point]*Actor) int
    combatUtility         func(userPosition geometry.Point, enemyPositions map[geometry.Point]*Actor) int
}

func (s *Action) SetActionColor(c color.Color) {
    s.color = c
}

func (s *Action) Execute(engine Engine, caster *Actor) {
    if !s.canPayCost(engine, caster, s.cost) {
        return
    }
    s.payCost(engine, caster, s.cost)
    if s.effect != nil {
        if s.closeModalsForEffect {
            engine.CloseAllModals()
        }
        s.effect(engine, caster)
    }
}

func (s *Action) ExecuteOnTarget(engine Engine, caster *Actor, pos geometry.Point) {
    if !s.canPayCost(engine, caster, s.cost) {
        return
    }
    s.payCost(engine, caster, s.cost)
    if s.targetedEffect != nil {
        if s.closeModalsForEffect {
            engine.CloseAllModals()
        }
        s.targetedEffect(engine, caster, pos)
    }
}

func (s *Action) GetValidTargets(engine Engine, caster *Actor, usePosition geometry.Point) []geometry.Point {
    return s.getValidPositions(engine, caster, usePosition)
}
func (s *Action) IsTargeted() bool {
    return s.targetedEffect != nil && s.effect == nil
}

func (s *Action) ManaCost() int {
    return s.cost
}
func (s *Action) Name() string {
    return s.name
}

func (s *Action) GetValue() int {
    return s.cost * 10
}

func (s *Action) Color() color.Color {
    return s.color
}

func (s *Action) SetValidTargets(filter func(engine Engine, user *Actor, usePosition geometry.Point) []geometry.Point) {
    s.getValidPositions = filter
}

func (s *Action) SetDescription(text []string) {
    s.description = text
}

func (s *Action) GetDescription() []string {
    return s.description
}

func (s *Action) GetCombatUtilityForTargetedUseOnLocation(target geometry.Point, enemyPositions map[geometry.Point]*Actor) int {
    return s.targetedCombatUtility(target, enemyPositions)
}

func (s *Action) GetCombatUtilityForUseAtLocation(userPosition geometry.Point, enemyPositions map[geometry.Point]*Actor) int {
    return s.combatUtility(userPosition, enemyPositions)
}
func (s *Action) SetCombatUtilityForTargetedUseOnLocation(utility func(target geometry.Point, enemyPositions map[geometry.Point]*Actor) int) {
    s.targetedCombatUtility = utility
}

func (s *Action) SetCombatUtilityForUseAtLocation(utility func(userPosition geometry.Point, enemyPositions map[geometry.Point]*Actor) int) {
    s.combatUtility = utility
}
