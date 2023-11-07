package game

import (
    "Legacy/geometry"
    "Legacy/gridmap"
    "Legacy/renderer"
    "image/color"
)

type ItemContainer interface {
    GetItems() []Item
    RemoveItem(item Item)
}

type Wearable interface {
    Item
    GetWearer() ItemWearer
    SetWearer(wearer ItemWearer)
    Unequip()
    IsEquipped() bool
}

type Engine interface {
    StartConversation(a *Actor, conversation *Dialogue)
    ShowScrollableText(text []string, textcolor color.Color, autolayout bool) *renderer.ScrollableTextWindow
    GetScrollFile(filename string) []string
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
    ShowMultipleChoiceDialogue(icon int32, text [][]string, choices []renderer.MenuItem)
    RemoveItem(item Item)
    GetPartyMembers() []*Actor
    ShowEquipMenu(a Wearable)
    StartCombat(opponents *Actor)
    GetAoECircle(pos geometry.Point, radius int) []geometry.Point
    HitAnimation(pos geometry.Point, atlasName renderer.AtlasName, icon int32, tintColor color.Color, whenDone func())
    SpellDamageAt(caster *Actor, pos geometry.Point, amount int)
    GetPartyEquipment() []Item
    GetRules() *Rules
    CanLevelUp(member *Actor) (bool, int)
    FreezeActorAt(pos geometry.Point, turns int)
    ProdActor(prodder *Actor, victim *Actor)
    GetBreakingToolName() string
    AskUserForString(prompt string, maxLength int, onConfirm func(text string))
    TeleportTo(text string)
    GetRandomPositionsInRegion(regionName string, count int) []geometry.Point
    GetGridMap() *gridmap.GridMap[*Actor, Item, Object]
    ChangeAppearance()
    RemoveDoorAt(pos geometry.Point)
    SetWallAt(pos geometry.Point)
    MoveAvatarInDirection(point geometry.Point)
    GetRegion(regionName string) geometry.Rect
    DrawCharInWorld(charToDraw rune, pos geometry.Point)
    RaiseAsUndeadForParty(pos geometry.Point)
    GetActorByInternalName(internalName string) *Actor
    GetDialogueFromFile(conversationId string) *Dialogue
    GetVisibleMap() geometry.Rect
}
