package main

import (
    "Legacy/game"
    "github.com/hajimehoshi/ebiten/v2"
)

// Interface Implementation for the Game

func (g *GridEngine) ManaSpent(caster *game.Actor, cost int) {
    caster.RemoveMana(cost)
    // TODO: remove HP from mother nature here..
    g.ManaSpentInWorld(caster.Pos(), cost)
}
func (g *GridEngine) DamageAvatar(amount int) {
    g.GetAvatar().Damage(amount)
    // TODO: visual indicator?
}

func (g *GridEngine) TriggerEvent(event string) {
    encounter := game.GetEncounter(g, event)
    encounter.Start()
    g.currentEncounter = encounter
}

func (g *GridEngine) GetPartyMembers() []*game.Actor {
    return g.playerParty.GetMembers()
}

func (g *GridEngine) GetMapName() string {
    return g.currentMap.GetName()
}

func (g *GridEngine) AddLockpicks(amount int) {
    g.playerParty.AddLockpicks(amount)
}

func (g *GridEngine) PartyHasLockpick() bool {
    return g.playerParty.GetLockpicks() > 0
}

func (g *GridEngine) RemoveLockpick() {
    g.playerParty.RemoveLockpicks(1)
}

func (g *GridEngine) RemoveItem(item game.Item) {
    if item.IsHeld() {
        item.GetHolder().RemoveItem(item)
    } else {
        g.currentMap.RemoveItem(item)
    }
}
func (g *GridEngine) Flags() *game.Flags {
    return g.flags
}

func (g *GridEngine) IsPlayerControlled(holder game.ItemHolder) bool {
    for _, member := range g.playerParty.GetMembers() {
        if member == holder {
            return true
        }
    }
    return g.playerParty == holder
}

func (g *GridEngine) GetTextFile(filename string) []string {
    return readLines(filename)
}

func (g *GridEngine) GetAvatar() *game.Actor {
    if g.splitControlled != nil {
        return g.splitControlled
    }
    return g.avatar
}

func (g *GridEngine) Print(text string) {
    g.textToPrint = text
    g.ticksForPrint = secondsToTicks(2)
}

func (g *GridEngine) CurrentTick() uint64 {
    return g.WorldTicks
}

func (g *GridEngine) TicksToSeconds(ticks uint64) float64 {
    return float64(ticks) / ebiten.ActualTPS()
}
