package game

import (
    "Legacy/renderer"
    "image/color"
)

type ItemContainer interface {
    GetItems() []Item
    RemoveItem(item Item)
}

type Engine interface {
    StartConversation(a *Actor)
    ShowColoredText(text []string, textcolor color.Color, autolayout bool)
    GetTextFile(filename string) []string
    PickUpItem(item Item)
    DropItem(item Item)
    GetAvatar() *Actor
    IsPlayerControlled(holder ItemHolder) bool
    SwitchAvatarTo(member *Actor)
    Flags() *Flags
    CreateLootForContainer(level int, lootType []Loot) []Item
    ShowContainer(container ItemContainer)
    OpenPickpocketMenu(victim *Actor)
    Print(text string)
    AddFood(amount int)
    AddGold(amount int)
    AddLockpicks(amount int)
    GetPartySize() int
    PartyHasKey(key string) bool
    PartyHasLockpick() bool
    RemoveLockpick()
    ShowDrinkPotionMenu(potion *Potion)
    ManaSpent(caster *Actor, cost int)
    DamageAvatar(amount int)
    TriggerEvent(event string)
    GetMapName() string
    CurrentTick() uint64
    TicksToSeconds(ticks uint64) float64
    ShowMultipleChoiceDialogue(icon int, text []string, choices []renderer.MenuItem)
    RemoveItem(item Item)
    GetPartyMembers() []*Actor
    ShowEquipMenu(a *Armor)
    StartCombat(opponents *Actor)
}
