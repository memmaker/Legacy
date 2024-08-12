package game

import (
    "Legacy/geometry"
    "image/color"
)

func (s *Spell) GetScrollTitle() string {
    return s.scrollTitle
}

func (s *Spell) GetScrollFile() string {
    return s.scrollFile
}

func (s *Spell) SetScrollTitle(title string) {
    s.scrollTitle = title
}

func (s *Spell) SetScrollFile(file string) {
    s.scrollFile = file
}

type BaseAction struct {
    effect               func(engine Engine, user *Actor)
    targetedEffect       func(engine Engine, user *Actor, pos geometry.Point)
    name                 string
    color                color.Color
    closeModalsForEffect bool

    canPayCost        func(engine Engine, user *Actor) bool
    payCost           func(engine Engine, user *Actor)
    getValidPositions func(engine Engine, user *Actor, usePosition geometry.Point) map[geometry.Point]bool
    description       []string

    // AI stuff
    targetedCombatUtility func(engine Engine, caster *Actor, target geometry.Point, allyPositions, enemyPositions map[geometry.Point]*Actor) int
    combatUtility         func(engine Engine, caster *Actor, userPosition geometry.Point, allyPositions, enemyPositions map[geometry.Point]*Actor) int
    monetaryValue         int
}

func (s *BaseAction) LabelWithCost() string {
    return s.name
}

func (s *BaseAction) SetActionColor(c color.Color) {
    s.color = c
}

func (s *BaseAction) Execute(engine Engine, caster *Actor) {
    if !s.canPayCost(engine, caster) {
        return
    }
    s.payCost(engine, caster)
    if s.effect != nil {
        if s.closeModalsForEffect {
            engine.CloseAllModals()
        }
        s.effect(engine, caster)
    }
}

func (s *BaseAction) ExecuteOnTarget(engine Engine, caster *Actor, pos geometry.Point) {
    if !s.canPayCost(engine, caster) {
        return
    }
    s.payCost(engine, caster)
    if s.targetedEffect != nil {
        if s.closeModalsForEffect {
            engine.CloseAllModals()
        }
        s.targetedEffect(engine, caster, pos)
    }
}

func (s *BaseAction) GetValidTargets(engine Engine, caster *Actor, usePosition geometry.Point) map[geometry.Point]bool {
    return s.getValidPositions(engine, caster, usePosition)
}
func (s *BaseAction) IsTargeted() bool {
    return s.targetedEffect != nil && s.effect == nil
}

func (s *BaseAction) Name() string {
    return s.name
}

func (s *BaseAction) GetValue() int {
    return s.monetaryValue
}

func (s *BaseAction) GetColor() color.Color {
    return s.color
}

func (s *BaseAction) SetValidTargets(filter func(engine Engine, user *Actor, usePosition geometry.Point) map[geometry.Point]bool) {
    s.getValidPositions = filter
}

func (s *BaseAction) SetDescription(text []string) {
    s.description = text
}

func (s *BaseAction) GetDescription() []string {
    return s.description
}

func (s *BaseAction) GetCombatUtilityForTargetedUseOnLocation(engine Engine, caster *Actor, target geometry.Point, allyPositions, enemyPositions map[geometry.Point]*Actor) int {
    return s.targetedCombatUtility(engine, caster, target, allyPositions, enemyPositions)
}

func (s *BaseAction) GetCombatUtilityForUseAtLocation(engine Engine, caster *Actor, userPosition geometry.Point, allyPositions, enemyPositions map[geometry.Point]*Actor) int {
    return s.combatUtility(engine, caster, userPosition, allyPositions, enemyPositions)
}
func (s *BaseAction) SetCombatUtilityForTargetedUseOnLocation(utility func(engine Engine, caster *Actor, target geometry.Point, allyPositions, enemyPositions map[geometry.Point]*Actor) int) {
    s.targetedCombatUtility = utility
}

func (s *BaseAction) SetCombatUtilityForUseAtLocation(utility func(engine Engine, caster *Actor, userPosition geometry.Point, allyPositions, enemyPositions map[geometry.Point]*Actor) int) {
    s.combatUtility = utility
}

func (s *BaseAction) SetNoCombatUtility() {
    s.combatUtility = func(engine Engine, caster *Actor, userPosition geometry.Point, allyPositions, enemyPositions map[geometry.Point]*Actor) int {
        return -50
    }
    s.targetedCombatUtility = func(engine Engine, caster *Actor, target geometry.Point, allyPositions, enemyPositions map[geometry.Point]*Actor) int {
        return -50
    }
}

func (s *BaseAction) CanPayCost(engine Engine, member *Actor) bool {
    return s.canPayCost(engine, member)
}

type Action interface {
    Name() string
    Execute(engine Engine, caster *Actor)
    ExecuteOnTarget(engine Engine, caster *Actor, pos geometry.Point)
    GetValidTargets(engine Engine, caster *Actor, usePosition geometry.Point) map[geometry.Point]bool
    CanPayCost(engine Engine, member *Actor) bool
    IsTargeted() bool
    GetValue() int
    GetColor() color.Color
    GetDescription() []string
    GetCombatUtilityForTargetedUseOnLocation(engine Engine, caster *Actor, target geometry.Point, allyPositions, enemyPositions map[geometry.Point]*Actor) int
    GetCombatUtilityForUseAtLocation(engine Engine, caster *Actor, userPosition geometry.Point, allyPositions, enemyPositions map[geometry.Point]*Actor) int
    LabelWithCost() string
}
