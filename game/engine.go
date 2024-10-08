package game

import (
    "Legacy/geometry"
    "Legacy/gridmap"
    "Legacy/renderer"
    "Legacy/util"
    "image/color"
)

type ItemContainer interface {
    GetItems() []Item
    RemoveItem(item Item)
}

type Engine interface {
    StartConversation(a *Actor, conversation *Dialogue)
    ShowScrollableText(text []string, textcolor color.Color, autolayout bool)
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
    OpenPlantMenu(victim *Actor)
    Print(text string)
    AddFood(amount int)
    AddGold(amount int)
    AddLockpicks(amount int)
    GetPartySize() int
    RemoveLockpick()
    ShowDrinkPotionMenu(potion *Potion)
    ManaSpent(caster *Actor, cost int)
    DamageAvatar(amount int)
    TriggerEvent(event string)
    GetMapName() string
    CurrentTick() uint64
    TicksToSeconds(ticks uint64) float64
    ShowMultipleChoiceDialogue(canBeClosed bool, icon int32, text [][]string, choices []util.MenuItem)
    RemoveItem(item Item)
    GetPartyMembers() []*Actor
    ShowEquipMenu(a Equippable)
    PlayerStartsCombat(opponent *Actor)
    PlayerTriesBackstab(opponent *Actor)
    PlayerStartsOffensiveSpell(caster *Actor, spell *Spell)
    GetAoECircle(pos geometry.Point, radius int) []geometry.Point
    CombatHitAnimation(pos geometry.Point, atlasName renderer.AtlasName, icon int32, tintColor color.Color, whenDone func())
    FixedDamageAt(caster *Actor, pos geometry.Point, amount int)
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
    PlayerMovement(point geometry.Point)
    GetRegion(regionName string) geometry.Rect
    DrawCharInWorld(charToDraw rune, pos geometry.Point)
    RaiseAsUndeadAt(caster *Actor, pos geometry.Point)
    GetActorByInternalName(internalName string) *Actor
    GetDialogueFromFile(conversationId string) *Dialogue
    GetVisibleMap() geometry.Rect
    GetParty() *Party
    OpenMenu(actions []util.MenuItem)
    OpenEquipmentDetails(partyIndex int)
    EquipItem(actor *Actor, item Equippable)
    OpenPartyInventoryOnPage(page int)
    CloseAllModals()

    GetWorldTime() WorldTime
    AdvanceWorldTime(days, hours, minutes int)
    Kill(actor *Actor)
    TakeItem(item Item)
    SkillCheck(actor *Actor, skill SkillName, difficulty DifficultyLevel) bool
    SkillCheckAvatar(skill SkillName, difficulty DifficultyLevel) bool
    HasSkill(skill SkillName) bool
    GetRelativeDifficulty(skill SkillName, difficulty DifficultyLevel) DifficultyLevel
    TransitionToNamedLocation(targetMap string, location string)
    ResetAllLockedDoorsOnMap(mapName string)
    GetChestByInternalName(internalName string) *Chest
    CloseConversation()
    UnlockDoorsByKeyName(keyName string)
    TryMoveNPCOnPath(actor *Actor, dest geometry.Point)
    IsSneaking() bool
    AddStatusEffect(victim *Actor, statusEffect StatusEffect, stacks int)
}
