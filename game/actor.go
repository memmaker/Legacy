package game

import (
    "Legacy/geometry"
    "Legacy/recfile"
    "Legacy/util"
    "fmt"
    "image/color"
    "io"
    "sort"
    "strconv"
)

type ItemSlot string

const (
    ItemSlotRightHand ItemSlot = "right hand"
    ItemSlotLeftHand  ItemSlot = "left hand"
    ItemSlotRanged    ItemSlot = "ranged"
    ItemSlotScroll    ItemSlot = "scroll"
)

type ArmorSlot string

func (s ArmorSlot) IsEqualTo(other ArmorSlot) bool {
    // three cases
    // 1 exact match
    // both are rings
    // both are amulets

    if s == other {
        return true
    }
    if s.IsRing() && other.IsRing() {
        return true
    }
    if s.IsAmulet() && other.IsAmulet() {
        return true
    }
    return false
}

func (s ArmorSlot) IsRing() bool {
    return s == AccessorySlotRingLeft || s == AccessorySlotRingRight
}

func (s ArmorSlot) IsAmulet() bool {
    return s == AccessorySlotAmuletOne || s == AccessorySlotAmuletTwo
}

func (s ArmorSlot) BaseValue() int {
    switch s {
    case ArmorSlotHelmet:
        return 2000
    case ArmorSlotBreastPlate:
        return 12000
    case ArmorSlotShoes:
        return 1500
    case AccessorySlotRobe:
        return 900
    case AccessorySlotRingLeft:
        fallthrough
    case AccessorySlotRingRight:
        return 900
    case AccessorySlotAmuletOne:
        fallthrough
    case AccessorySlotAmuletTwo:
        return 700
    }
    return 200
}

const (
    ArmorSlotHelmet      ArmorSlot = "helmet"
    ArmorSlotBreastPlate ArmorSlot = "breast plate"
    ArmorSlotShoes       ArmorSlot = "shoes"

    AccessorySlotRobe      ArmorSlot = "robe"
    AccessorySlotRingLeft  ArmorSlot = "ring left"
    AccessorySlotRingRight ArmorSlot = "ring right"
    AccessorySlotAmuletOne ArmorSlot = "amulet one"
    AccessorySlotAmuletTwo ArmorSlot = "amulet two"
)

func GetAllArmorSlots() []ArmorSlot {
    return []ArmorSlot{
        ArmorSlotHelmet,
        ArmorSlotBreastPlate,
        ArmorSlotShoes,
        AccessorySlotRobe,
        AccessorySlotRingLeft,
        AccessorySlotRingRight,
        AccessorySlotAmuletOne,
        AccessorySlotAmuletTwo,
    }
}

type Equippable interface {
    Item
    GetWearer() ItemWearer
    SetWearer(wearer ItemWearer)
    Unequip()
    IsEquipped() bool
}
type Wearable interface {
    Equippable
    GetSlot() ArmorSlot
    IsBetterThan(other Wearable) bool
}

type Handheld interface {
    Equippable
    IsBetterThan(other Handheld) bool
}
type Actor struct {
    GameObject
    icon           int32
    iconFrameCount int
    name           string
    party          *Party
    dialogue       *Dialogue
    description    string

    // weapon slots
    equippedLeftHand  Handheld
    equippedRightHand Handheld
    equippedRanged    *Weapon

    // direct armor slots
    equippedArmor map[ArmorSlot]*Armor

    // accessory slots
    equippedAccessories map[ArmorSlot]*Armor

    // magic slots
    equippedScrolls []*Scroll

    internalName string
    isHuman      bool

    inventory []Item

    // stats
    mana      int
    maxHealth int
    health    int

    experiencePoints int

    attributes   AttributeHolder
    skillset     SkillSet
    innateSkills []*BaseAction

    color                 color.Color
    isTinted              bool
    combatFaction         string
    originalCombatFaction string
    isAggressive          bool
    engagementRange       int
    statusEffects         map[StatusEffectName]StatusEffect
    deathIcon             int32
    zoneOfEngagement      map[geometry.Point]int
}

func NewActor(name string, icon int32) *Actor {
    return &Actor{
        name:           name,
        icon:           icon,
        iconFrameCount: 1,
        health:         23,
        maxHealth:      23,

        isHuman:             true,
        inventory:           []Item{},
        skillset:            NewSkillSet(),
        color:               color.White,
        equippedArmor:       make(map[ArmorSlot]*Armor),
        equippedAccessories: make(map[ArmorSlot]*Armor),
        statusEffects:       make(map[StatusEffectName]StatusEffect),
        deathIcon:           24,

        attributes: NewAttributeHolder(),
    }
}

func NewActorFromFile(file io.ReadCloser, icon int32, toPages func(height int, inputText []string) [][]string) *Actor {
    defer file.Close()
    actorData := recfile.ReadMulti(file)

    coreRecord := actorData["Details"][0].ToMap()
    conversation := NewDialogueFromRecords(actorData["Conversation"], toPages)

    health, _ := coreRecord.GetInt("Health")
    description := coreRecord["Description"]

    newActor := &Actor{
        name:                coreRecord["Name"],
        icon:                icon,
        health:              health,
        maxHealth:           health,
        description:         description,
        dialogue:            conversation,
        isHuman:             true,
        skillset:            NewSkillSet(),
        attributes:          NewAttributeHolder(),
        equippedArmor:       make(map[ArmorSlot]*Armor),
        equippedAccessories: make(map[ArmorSlot]*Armor),
        statusEffects:       make(map[StatusEffectName]StatusEffect),
        deathIcon:           24,
    }

    if armorString, hasArmor := coreRecord["Torso"]; hasArmor {
        newActor.equippedArmor[ArmorSlotBreastPlate] = NewArmorFromPredicate(recfile.StrPredicate(armorString))
    }

    if armorString, hasArmor := coreRecord["Head"]; hasArmor {
        newActor.equippedArmor[ArmorSlotHelmet] = NewArmorFromPredicate(recfile.StrPredicate(armorString))
    }

    if weaponString, hasWeapon := coreRecord["RightHand"]; hasWeapon {
        newActor.equippedRightHand = NewWeaponFromPredicate(recfile.StrPredicate(weaponString))
    }

    if recordsForInventory, hasInventory := actorData["Inventory"]; hasInventory && len(recordsForInventory) > 0 {
        inventory := recordsForInventory[0].ToValueList()
        newActor.inventory = toInventory(newActor, itemsFromStrings(inventory))
    }

    if recordForSkills, hasSkills := actorData["Skills"]; hasSkills && len(recordForSkills) > 0 {
        skillRecord := recordForSkills[0]
        for _, field := range skillRecord {
            if field.Name == "ActiveSkill" {
                newActor.AddInnateSkill(NewActiveSkillFromName(SkillName(field.Value)))
            }
        }
    }

    for _, attributeName := range GetAllAttributeNames() {
        if value, hasAttr := coreRecord[string(attributeName)]; hasAttr {
            newActor.attributes.SetAttribute(attributeName, recfile.StrInt(value))
        }
    }

    for _, skillName := range GetAllSkillNames() {
        if value, hasSkill := coreRecord[string(skillName)]; hasSkill {
            newActor.skillset.SetSkill(skillName, recfile.StrInt(value))
        }
    }

    return newActor
}

func NewActorFromRecord(record recfile.Record) *Actor {
    a := &Actor{
        skillset:            NewSkillSet(), // TODO
        equippedArmor:       make(map[ArmorSlot]*Armor),
        equippedAccessories: make(map[ArmorSlot]*Armor),
        statusEffects:       make(map[StatusEffectName]StatusEffect),
        attributes:          NewAttributeHolder(),
        deathIcon:           24,
        //dialogue:        NewDialogueFromRecords(actorData["Conversation"]),
    }
    for _, field := range record {
        switch field.Name {
        case "name":
            a.name = field.Value
        case "internalName":
            a.internalName = field.Value
        case "position":
            a.SetPos(geometry.MustDecodePoint(field.Value))
        case "icon":
            a.icon = field.AsInt32()
        case "iconFrames":
            a.iconFrameCount = field.AsInt()
        case "isHuman":
            a.isHuman = field.AsBool()
        case "health":
            a.health = field.AsInt()
        case "maxhealth":
            a.maxHealth = field.AsInt()
        case "xp":
            a.experiencePoints = field.AsInt()
        case "mana":
            a.mana = field.AsInt()
        case "description":
            a.description = field.Value
        }
    }
    return a
}

func (a *Actor) ToRecord() recfile.Record {
    actorRecord := recfile.Record{
        recfile.Field{Name: "name", Value: a.Name()},
        recfile.Field{Name: "internalName", Value: a.GetInternalName()},
        recfile.Field{Name: "position", Value: a.Pos().Encode()},
        recfile.Field{Name: "icon", Value: recfile.Int32Str(a.Icon(0))},
        recfile.Field{Name: "iconFrames", Value: strconv.Itoa(a.GetIconFrameCount())},

        recfile.Field{Name: "isHuman", Value: recfile.BoolStr(a.IsHuman())},

        recfile.Field{Name: "health", Value: strconv.Itoa(a.GetHealth())},
        recfile.Field{Name: "maxhealth", Value: strconv.Itoa(a.GetMaxHealth())},
        recfile.Field{Name: "xp", Value: strconv.Itoa(a.GetXP())},
        recfile.Field{Name: "level", Value: strconv.Itoa(a.GetLevel())},
        recfile.Field{Name: "mana", Value: strconv.Itoa(a.GetMana())},

        recfile.Field{Name: "description", Value: a.description},
        // TODO: dialogue state, attributes & skills, equipment

    }
    if a.dialogue != nil {
        for keyword := range a.dialogue.keyWordsGiven {
            actorRecord = append(actorRecord, recfile.Field{Name: "d_key", Value: keyword})
        }
        for keyword := range a.dialogue.previouslyAsked {
            actorRecord = append(actorRecord, recfile.Field{Name: "d_prev", Value: keyword})
        }
    }
    return actorRecord
}

func (a *Actor) Unequip(item Item) {
    switch item.(type) {
    case Handheld:
        handheldItem := item.(Handheld)
        handheldItem.SetWearer(nil)

        if a.equippedLeftHand == handheldItem {
            a.equippedLeftHand = nil
        }
        if a.equippedRightHand == handheldItem {
            a.equippedRightHand = nil
        }
        if a.equippedRanged == handheldItem {
            a.equippedRanged = nil
        }
    case *Armor:
        armor := item.(*Armor)
        armor.SetWearer(nil)
        if equippedInSlot, isArmor := a.equippedArmor[armor.slot]; isArmor {
            if equippedInSlot == armor {
                delete(a.equippedArmor, armor.slot)
            }
        } else if equippedInAccessorySlot, isAccessory := a.equippedAccessories[armor.slot]; isAccessory {
            if equippedInAccessorySlot == armor {
                delete(a.equippedAccessories, armor.slot)
            }
        }
    case *Scroll:
        scroll := item.(*Scroll)
        scroll.SetWearer(nil)
        for i, equippedScroll := range a.equippedScrolls {
            if equippedScroll == scroll {
                a.equippedScrolls = append(a.equippedScrolls[:i], a.equippedScrolls[i+1:]...)
                break
            }
        }
    }

    a.party.onItemEquipStatusChanged([]Item{item})
}

func (a *Actor) OnAddedToParty(party *Party) {
    a.party = party
    a.SetCombatFaction("_player_party")
}

func (a *Actor) OnRemovedFromParty() {
    a.party = nil
    a.RestoreOriginalCombatFaction()
}

func toInventory(actor *Actor, items []Item) []Item {
    for _, item := range items {
        item.SetHolder(actor)
    }
    return items
}

func itemsFromStrings(inventory []string) []Item {
    var items []Item
    for _, itemString := range inventory {
        item := NewItemFromString(itemString)
        items = append(items, item)
    }
    return items
}

func (a *Actor) SetHuman(isHuman bool) {
    a.isHuman = isHuman
}
func (a *Actor) Icon(tick uint64) int32 {
    if !a.IsAlive() && a.isHuman {
        return a.deathIcon
    }
    if a.iconFrameCount == 1 {
        return a.icon
    }
    delays := tick / 20
    return a.icon + int32(delays%uint64(a.iconFrameCount))
}

func (a *Actor) Name() string {
    return a.name
}

func (a *Actor) GetMainStats(engine Engine) []string {
    _, xpNeeded := engine.CanLevelUp(a)
    tableData := []util.TableRow{
        {Label: "Level", Columns: []string{strconv.Itoa(a.GetLevel())}},
        {Label: "XP", Columns: []string{strconv.Itoa(a.experiencePoints)}},
        {Label: "Next Lvl.", Columns: []string{strconv.Itoa(xpNeeded)}},
        {Label: "Health", Columns: []string{fmt.Sprintf("%d/%d", a.health, a.maxHealth)}},
        {Label: "Mana", Columns: []string{fmt.Sprintf("%d", a.mana)}},
        {Label: "----", Columns: []string{"----"}},
        {Label: "Strength", Columns: []string{fmt.Sprintf("%d", a.attributes.GetAttribute(Strength))}},
        {Label: "Perception", Columns: []string{fmt.Sprintf("%d", a.attributes.GetAttribute(Perception))}},
        {Label: "Endurance", Columns: []string{fmt.Sprintf("%d", a.attributes.GetAttribute(Endurance))}},
        {Label: "Charisma", Columns: []string{fmt.Sprintf("%d", a.attributes.GetAttribute(Charisma))}},
        {Label: "Intelligence", Columns: []string{fmt.Sprintf("%d", a.attributes.GetAttribute(Intelligence))}},
        {Label: "Agility", Columns: []string{fmt.Sprintf("%d", a.attributes.GetAttribute(Agility))}},
        {Label: "Luck", Columns: []string{fmt.Sprintf("%d", a.attributes.GetAttribute(Luck))}},
    }
    return util.TableLayout(tableData)
}
func (a *Actor) GetDerivedStats(engine Engine) []string {
    tableData := []util.TableRow{
        {Label: "Initiative", Columns: []string{strconv.Itoa(a.attributes.GetInitiative())}},
        {Label: "Movement", Columns: []string{strconv.Itoa(a.attributes.GetMovementAllowance(a.GetTotalEncumbrance()))}},
        {Label: "Base Melee Damage", Columns: []string{strconv.Itoa(a.attributes.GetBaseMeleeDamage())}},
        {Label: "Base Ranged Damage", Columns: []string{strconv.Itoa(a.attributes.GetBaseRangedDamage())}},
        {Label: "Base Armor", Columns: []string{strconv.Itoa(a.attributes.GetBaseArmor())}},
    }
    return util.TableLayout(tableData)
}

func (a *Actor) LookDescription() []string {
    healthString := "healthy"
    var description []string
    if a.description != "" {
        description = append(description, util.AutoLayout(a.description, 32)...)
    }
    // wearables
    if len(a.equippedArmor) > 0 {
        description = append(description, "", "The person is wearing:")

        if helmet, exists := a.GetHelmet(); exists {
            description = append(description, fmt.Sprintf("  %s", helmet.Name()))
        }
        if breastPlate, exists := a.GetArmorBreastPlate(); exists {
            description = append(description, fmt.Sprintf("  %s", breastPlate.Name()))
        }
    }

    // hands
    if a.equippedLeftHand != nil || a.equippedRightHand != nil {
        description = append(description, "", "The person is holding:")
        if a.equippedLeftHand != nil {
            description = append(description, fmt.Sprintf("  %s", a.equippedLeftHand.Name()))
        }
        if a.equippedRightHand != nil {
            description = append(description, fmt.Sprintf("  %s", a.equippedRightHand.Name()))
        }
    }

    healthRatio := float64(a.health) / float64(a.maxHealth)

    if healthRatio < 0.5 {
        healthString = "wounded"
    } else if healthRatio < 0.25 {
        healthString = "severely wounded"
    } else if healthRatio < 0.1 {
        healthString = "near death"
    }

    if len(description) > 0 {
        description = append(description, "")
    }
    description = append(description, fmt.Sprintf("The person looks %s.", healthString))
    if len(a.statusEffects) > 0 {
        description = append(description, "")
        description = append(description, "The person is affected by:")
        for effect, _ := range a.statusEffects {
            description = append(description, fmt.Sprintf("  %s", effect))
        }
    }
    return description
}

func (a *Actor) GetContextActions(engine Engine) []util.MenuItem {
    var items []util.MenuItem
    if a != engine.GetAvatar() {
        talkTo := util.MenuItem{
            Text: "Talk",
            Action: func() {
                engine.StartConversation(a, a.GetDialogue())
            },
        }
        lookAt := util.MenuItem{
            Text: "Look",
            Action: func() {
                engine.ShowScrollableText(a.LookDescription(), color.White, false)
            },
        }

        stealDiff := engine.GetRelativeDifficulty(ThievingSkillPickpocket, a.GetAbsoluteDifficultyByAttribute(Perception)).ToString()

        steal := util.MenuItem{
            Text:   fmt.Sprintf("Steal - %s", stealDiff),
            Action: func() { engine.OpenPickpocketMenu(a) },
        }
        plant := util.MenuItem{
            Text:   fmt.Sprintf("Plant - %s", stealDiff),
            Action: func() { engine.OpenPlantMenu(a) },
        }
        attack := util.MenuItem{
            Text: "Attack",
            Action: func() {
                engine.PlayerStartsCombat(a)
            },
        }
        backstabDiff := engine.GetRelativeDifficulty(PhysicalSkillBackstab, a.GetAbsoluteDifficultyByAttribute(Perception)).ToString()
        backstab := util.MenuItem{
            Text: fmt.Sprintf("Backstab - %s", backstabDiff),
            Action: func() {
                engine.PlayerTriesBackstab(a)
            },
        }
        push := util.MenuItem{
            Text: "Prod",
            Action: func() {
                if engine.GetAvatar().IsRightNextTo(a) {
                    engine.ProdActor(engine.GetAvatar(), a)
                }
            },
        }
        items = append(items, talkTo, lookAt)
        if engine.GetAvatar().IsRightNextTo(a) {
            if engine.IsSneaking() {
                if engine.GetAvatar().CanBackstab(a) {
                    items = append(items, backstab)
                }
                items = append(items, steal, plant)
            } else {
                items = append(items, attack, push)
            }
        }
    }
    return items
}

func (a *Actor) HasKey(key string) bool {
    if a.party != nil {
        return a.party.HasKey(key)
    }
    return false
}

func (a *Actor) IsNearTo(other *Actor) bool {
    ownPos := a.Pos()
    otherPos := other.Pos()
    return geometry.DistanceManhattan(ownPos, otherPos) <= 2
}

func (a *Actor) IsRightNextTo(other *Actor) bool {
    ownPos := a.Pos()
    otherPos := other.Pos()
    return geometry.DistanceManhattan(ownPos, otherPos) <= 1
}

type SalesOffer struct {
    Item  Item
    Price int
}

func (a *Actor) GetItemsToSell() []SalesOffer {
    var items []SalesOffer
    for _, item := range a.inventory {
        if a.isStackableInList(items, item) {
            continue
        }
        if pseudo, ok := item.(*PseudoItem); ok {
            if pseudo.itemType == PseudoItemTypeGold {
                continue
            }
        }
        items = append(items, SalesOffer{
            Item:  item,
            Price: a.vendorPrice(item),
        })
    }

    sort.SliceStable(items, func(i, j int) bool {
        return items[i].Item.GetValue() > items[j].Item.GetValue()
    })
    return items
}

func (a *Actor) isStackableInList(items []SalesOffer, item Item) bool {
    for _, existingOffer := range items {
        if existingOffer.Item.CanStackWith(item) {
            return true
        }
    }
    return false
}

func (a *Actor) appendOffer(offer map[Loot][]SalesOffer, lootType Loot, item Item) map[Loot][]SalesOffer {
    if _, ok := offer[lootType]; !ok {
        offer[lootType] = make([]SalesOffer, 0)
    }

    return offer
}

func (a *Actor) RemoveItem(item Item) bool {
    if item.GetHolder() != a {
        return false
    }
    ownedItemUntilNow := false
    for i, inventoryItem := range a.inventory {
        if inventoryItem == item {
            a.inventory = append(a.inventory[:i], a.inventory[i+1:]...)
            ownedItemUntilNow = true
            break
        }
    }
    if ownedItemUntilNow {
        item.SetHolder(nil)
    }
    return ownedItemUntilNow
}

func (a *Actor) CanEquip(item Item) bool {
    switch item.(type) {
    case *Armor:
        return true
    case *Scroll:
        return true
    case Handheld:
        return true
    }
    return false
}

// Equip will
// 1. Check if the item can be equipped
// 2. Unequip the item if it is already equipped
// 3. Unequip any other item in the same slot
// 4. Equip the item
func (a *Actor) Equip(item Item) {
    if !a.CanEquip(item) {
        return
    }
    switch typedItem := item.(type) {
    case *Armor:
        a.EquipArmor(typedItem, typedItem.GetSlot())
    case *Scroll:
        a.EquipScroll(typedItem)
    case Handheld:
        a.EquipRightHand(typedItem)
    }
}

func (a *Actor) EquipArmor(armor *Armor, slot ArmorSlot) {
    if !a.CanEquip(armor) || armor == nil {
        return
    }
    armor.Unequip()
    changedItems := []Item{armor}

    if armor.IsAccessory() {
        chosenSlot := slot
        equippedInSlot, ok := a.equippedAccessories[chosenSlot]
        if ok && equippedInSlot != nil {
            equippedInSlot.Unequip()
            changedItems = append(changedItems, equippedInSlot)
        }
        a.equippedAccessories[chosenSlot] = armor
        armor.SetWearer(a)
        armor.SetSlotUsed(chosenSlot)
        a.party.onItemEquipStatusChanged(changedItems)
        return
    }
    equippedInSlot, ok := a.equippedArmor[armor.slot]
    if ok && equippedInSlot != nil {
        equippedInSlot.Unequip()
        changedItems = append(changedItems, equippedInSlot)
    }
    a.equippedArmor[armor.slot] = armor
    armor.SetWearer(a)

    a.party.onItemEquipStatusChanged(changedItems)
}
func (a *Actor) EquipScroll(scroll *Scroll) {
    if !a.CanEquip(scroll) || scroll == nil {
        return
    }
    scroll.Unequip()

    a.equippedScrolls = append(a.equippedScrolls, scroll)
    scroll.SetWearer(a)
    scroll.SetHolder(a)
    a.party.onItemEquipStatusChanged([]Item{scroll})
}

func (a *Actor) EquipRangedWeapon(weapon *Weapon) {
    if !a.CanEquip(weapon) || weapon == nil || !weapon.IsRanged() {
        return
    }
    weapon.Unequip()
    var oldWeapon *Weapon
    if a.equippedRanged != nil {
        a.equippedRanged.Unequip()
        oldWeapon = a.equippedRanged
    }
    a.equippedRanged = weapon
    weapon.SetWearer(a)
    weapon.SetHolder(a)
    if oldWeapon != nil {
        a.party.onItemEquipStatusChanged([]Item{oldWeapon, weapon})
    } else {
        a.party.onItemEquipStatusChanged([]Item{weapon})
    }
}
func (a *Actor) EquipRightHand(item Handheld) {
    if !a.CanEquip(item) || item == nil {
        return
    }
    item.Unequip()
    var oldItem Handheld
    if a.equippedRightHand != nil {
        a.equippedRightHand.Unequip()
        oldItem = a.equippedRightHand
    }
    a.equippedRightHand = item
    item.SetWearer(a)
    item.SetHolder(a)
    if oldItem != nil {
        a.party.onItemEquipStatusChanged([]Item{oldItem, item})
    } else {
        a.party.onItemEquipStatusChanged([]Item{item})
    }
}
func (a *Actor) EquipLeftHand(item Handheld) {
    if !a.CanEquip(item) || item == nil {
        return
    }
    item.Unequip()
    var oldItem Handheld
    if a.equippedLeftHand != nil {
        a.equippedLeftHand.Unequip()
        oldItem = a.equippedLeftHand
    }
    a.equippedLeftHand = item
    item.SetWearer(a)
    item.SetHolder(a)
    if oldItem != nil {
        a.party.onItemEquipStatusChanged([]Item{oldItem, item})
    } else {
        a.party.onItemEquipStatusChanged([]Item{item})
    }
}

func (a *Actor) HasDialogue() bool {
    return a.dialogue != nil
}

func (a *Actor) GetDialogue() *Dialogue {
    return a.dialogue
}

func (a *Actor) SetName(name string) {
    a.name = name
}

func (a *Actor) SetInternalName(name string) {
    a.internalName = name
}

func (a *Actor) GetInternalName() string {
    return a.internalName
}

func (a *Actor) HasMana(cost int) bool {
    return a.mana >= cost
}

func (a *Actor) AddMana(mana int) {
    a.mana += mana
}

func (a *Actor) RemoveMana(mana int) {
    a.mana -= mana
}

func (a *Actor) SetIconFrames(frames int) {
    a.iconFrameCount = frames
}

func (a *Actor) IsHuman() bool {
    return a.isHuman
}

func (a *Actor) FullRest(engine Engine) {
    if a.health < a.maxHealth {
        a.health = a.maxHealth
    }
    a.OnRest(engine)
}

func (a *Actor) Damage(engine Engine, amount int) {
    a.health -= amount
    a.OnDamageReceived(engine, amount)
}

func (a *Actor) GetItemsToSteal() []Item {
    return a.inventory
}

func (a *Actor) AddXP(amount int) {
    a.experiencePoints += amount
}

func (a *Actor) AddGold(amount int) {
    for _, item := range a.inventory {
        if pseudo, ok := item.(*PseudoItem); ok {
            if pseudo.itemType == PseudoItemTypeGold {
                pseudo.amount += amount
                return
            }
        }
    }
    a.inventory = append(a.inventory, NewPseudoItemFromTypeAndAmount(PseudoItemTypeGold, amount))
}

func (a *Actor) GetTotalArmor() int {
    baseArmor := a.attributes.GetBaseArmor()
    armorSum := 0
    for _, armor := range a.equippedArmor {
        armorSum += armor.GetProtection()
    }
    return baseArmor + armorSum
}

func (a *Actor) GetMeleeDamage() int {
    baseDamage := a.attributes.GetBaseMeleeDamage()
    if a.equippedRightHand != nil {
        if weapon, ok := a.equippedRightHand.(*Weapon); ok {
            return weapon.GetDamage(baseDamage)
        }
    }
    if a.equippedLeftHand != nil {
        if weapon, ok := a.equippedLeftHand.(*Weapon); ok {
            return weapon.GetDamage(baseDamage)
        }
    }
    return baseDamage
}

func (a *Actor) GetRangedDamage() int {
    baseDamage := a.attributes.GetBaseRangedDamage()
    if a.equippedRanged != nil {
        return a.equippedRanged.GetDamage(baseDamage)
    }
    return baseDamage
}

func (a *Actor) IsAlive() bool {
    return a.health > 0
}

func (a *Actor) DropInventory() []Item {
    dropped := a.inventory
    a.inventory = []Item{}
    for _, item := range dropped {
        item.SetHolder(nil)
    }
    return dropped
}

func (a *Actor) GetMovementAllowance() int {
    encumberance := a.GetTotalEncumbrance()
    return a.attributes.GetMovementAllowance(encumberance)
}

func (a *Actor) SetHealth(health int) {
    a.health = health
}

func (a *Actor) SetMana(mana int) {
    a.mana = mana
}

func (a *Actor) StripGear() {
    for _, item := range a.equippedArmor {
        item.Unequip()
    }
    for _, item := range a.equippedAccessories {
        item.Unequip()
    }

}

func (a *Actor) AutoEquip(engine Engine) {
    bestHandheld := map[ItemSlot]Handheld{
        ItemSlotRightHand: nil,
        ItemSlotRanged:    nil,
    }
    bestArmor := map[ArmorSlot]*Armor{
        ArmorSlotBreastPlate: nil,
        ArmorSlotHelmet:      nil,
        ArmorSlotShoes:       nil,

        AccessorySlotRobe:      nil,
        AccessorySlotRingLeft:  nil,
        AccessorySlotRingRight: nil,
        AccessorySlotAmuletOne: nil,
        AccessorySlotAmuletTwo: nil,
    }
    allEquipment := engine.GetPartyEquipment()

    for _, item := range allEquipment {
        wearable, isWearable := item.(*Armor)
        if isWearable && a.CanEquip(item) && !wearable.IsEquipped() {
            currentBest, hasSlot := bestArmor[wearable.GetSlot()]
            if !hasSlot {
                continue
            }
            if wearable.IsBetterThan(currentBest) {
                bestArmor[wearable.GetSlot()] = wearable
            }

            continue
        }

        handheld, isHandheld := item.(Handheld)
        if isHandheld && a.CanEquip(item) && !handheld.IsEquipped() {
            if weapon, isWeapon := item.(*Weapon); isWeapon {
                slot := ItemSlotRightHand
                if weapon.IsRanged() {
                    slot = ItemSlotRanged
                }
                currentBest := bestHandheld[slot]
                if currentBest == nil || weapon.IsBetterThan(currentBest.(*Weapon)) {
                    bestHandheld[slot] = weapon
                }

            }
        }

    }
    for _, item := range bestArmor {
        if item != nil {
            engine.EquipItem(a, item)
        }
    }
    for _, item := range bestHandheld {
        if item != nil {
            engine.EquipItem(a, item)
        }
    }
}

func (a *Actor) GetMagicDefense() int {
    return 0
}

func (a *Actor) vendorPrice(item Item) int {
    return int(float64(item.GetValue()) * 1.5)
}

func (a *Actor) GetSkills() *SkillSet {
    return &a.skillset
}
func (a *Actor) GetHealth() int {
    return a.health
}

func (a *Actor) GetMaxHealth() int {
    return a.maxHealth
}

func (a *Actor) GetXP() int {
    return a.experiencePoints
}

func (a *Actor) GetLevel() int {
    return a.attributes.GetAttribute(Level)
}

func (a *Actor) GetMana() int {
    return a.mana
}
func (a *Actor) GetIconFrameCount() int {
    return a.iconFrameCount
}

func (a *Actor) GetXPForKilling() int {
    return a.GetLevel() * 25
}

func (a *Actor) AddAllSkills() {
    a.skillset.AddMasteryInAllSkills()
}

func (a *Actor) OnRest(engine Engine) {
    for effectName, effect := range a.statusEffects {
        if onRestEffect, ok := effect.(OnRestEffect); ok {
            onRestEffect.OnRest(engine, a)
            if effect.IsExpired() {
                a.RemoveStatusEffect(engine, effectName)
            }
        }
    }
}

func (a *Actor) TintColor() color.Color {
    return a.color
}

func (a *Actor) SetTintColor(color color.Color) {
    a.color = color
}

func (a *Actor) IsTinted() bool {
    return a.isTinted
}

func (a *Actor) SetTinted(value bool) {
    a.isTinted = value
}

func (a *Actor) SetIcon(icon int32) {
    a.icon = icon
}

func (a *Actor) SetVendorInventory(items []Item) {
    for _, item := range items {
        item.SetHolder(a)
    }
    a.inventory = append(a.inventory, items...)
}

func (a *Actor) GetEquippedSpells() []Action {
    var spells []Action
    for _, scroll := range a.equippedScrolls {
        spells = append(spells, scroll.spell)
    }
    return spells

}

func (a *Actor) GetActiveSkills() []Action {
    var skills []Action
    // todo: add skills from items, etc.
    for _, skill := range a.innateSkills {
        skills = append(skills, skill)
    }
    leftItem := a.equippedLeftHand
    if leftItem != nil {
        skills = append(skills, leftItem.GetEmbeddedActions()...)
    }
    rightItem := a.equippedRightHand
    if rightItem != nil {
        skills = append(skills, rightItem.GetEmbeddedActions()...)
    }
    return skills
}

func (a *Actor) GetRightHandItem() (Handheld, bool) {
    return a.equippedRightHand, a.equippedRightHand != nil
}

func (a *Actor) GetLeftHandItem() (Handheld, bool) {
    return a.equippedLeftHand, a.equippedLeftHand != nil
}

func (a *Actor) GetRangedItem() (*Weapon, bool) {
    return a.equippedRanged, a.equippedRanged != nil
}

func (a *Actor) GetHelmet() (*Armor, bool) {
    armor, ok := a.equippedArmor[ArmorSlotHelmet]
    return armor, ok
}

func (a *Actor) GetArmorBreastPlate() (*Armor, bool) {
    breastPlate, ok := a.equippedArmor[ArmorSlotBreastPlate]
    return breastPlate, ok
}

func (a *Actor) GetShoes() (*Armor, bool) {
    shoes, ok := a.equippedArmor[ArmorSlotShoes]
    return shoes, ok
}

func (a *Actor) GetRobe() (*Armor, bool) {
    robe, ok := a.equippedAccessories[AccessorySlotRobe]
    return robe, ok
}

func (a *Actor) GetRingLeft() (*Armor, bool) {
    ring, ok := a.equippedAccessories[AccessorySlotRingLeft]
    return ring, ok
}

func (a *Actor) GetRingRight() (*Armor, bool) {
    ring, ok := a.equippedAccessories[AccessorySlotRingRight]
    return ring, ok
}

func (a *Actor) GetAmuletOne() (*Armor, bool) {
    amulet, ok := a.equippedAccessories[AccessorySlotAmuletOne]
    return amulet, ok
}

func (a *Actor) GetAmuletTwo() (*Armor, bool) {
    amulet, ok := a.equippedAccessories[AccessorySlotAmuletTwo]
    return amulet, ok
}

func (a *Actor) chooseAccessorySlot(armor *Armor) ArmorSlot {
    if armor.GetSlot().IsRing() {
        if equipped, ok := a.equippedAccessories[AccessorySlotRingLeft]; !ok || equipped == nil {
            return AccessorySlotRingLeft
        }
        return AccessorySlotRingRight
    }
    if armor.GetSlot().IsAmulet() {
        if equipped, ok := a.equippedAccessories[AccessorySlotAmuletOne]; !ok || equipped == nil {
            return AccessorySlotAmuletOne
        }
        return AccessorySlotAmuletTwo
    }
    return armor.GetSlot()
}

func (a *Actor) SetCombatFaction(faction string) {
    a.combatFaction = faction
    if a.originalCombatFaction == "" {
        a.originalCombatFaction = faction
    }
}
func (a *Actor) GetCombatFaction() string {
    return a.combatFaction
}

func (a *Actor) RestoreOriginalCombatFaction() {
    if a.originalCombatFaction != "" {
        a.combatFaction = a.originalCombatFaction
    }
}

func (a *Actor) HasRangedWeaponEquipped() bool {
    return a.equippedRanged != nil
}

func (a *Actor) GetItemByName(name string) (Item, bool) {
    for _, item := range a.inventory {
        if item.Name() == name {
            return item, true
        }
    }
    return nil, false
}

func (a *Actor) UsedKey(key string) {
    if a.party != nil {
        a.party.UsedKey(key)
    }
}

func (a *Actor) AddItem(item Item) {
    a.inventory = append(a.inventory, item)
    item.SetHolder(a)
}

func (a *Actor) SetAggressive(isAggressive bool) {
    a.isAggressive = isAggressive
}

func (a *Actor) IsAggressive() bool {
    return a.isAggressive
}

func (a *Actor) GetNPCEngagementRange() int {
    return a.engagementRange
}

func (a *Actor) SetNPCEngagementRange(rangeInTiles int) {
    a.engagementRange = rangeInTiles
}

func (a *Actor) OnMeleeHitPerformed(engine Engine, victim *Actor) {
    if a.HasMeleeWeaponEquipped() {
        weaponUsed := a.GetMeleeWeapon()
        weaponUsed.OnHitProc(engine, a, victim)
    }
}

func (a *Actor) OnRangedHitPerformed(engine Engine, victim *Actor) {
    if a.HasRangedWeaponEquipped() {
        weaponUsed := a.GetRangedWeapon()
        weaponUsed.OnHitProc(engine, a, victim)
    }
}

func (a *Actor) OnDamageReceived(engine Engine, amount int) {
    if amount <= 0 {
        return
    }
    for _, effect := range a.statusEffects {
        if onDamageEffect, isOnDamageEffect := effect.(OnDamageEffect); isOnDamageEffect {
            onDamageEffect.OnDamageReceived(engine, a, amount)
            if onDamageEffect.IsExpired() {
                a.RemoveStatusEffect(engine, effect.Name())
            }
        }
    }
}

func (a *Actor) HasMeleeWeaponEquipped() bool {
    if a.equippedRightHand != nil {
        if _, ok := a.equippedRightHand.(*Weapon); ok {
            return true
        }
    }
    if a.equippedLeftHand != nil {
        if _, ok := a.equippedLeftHand.(*Weapon); ok {
            return true
        }
    }
    return false
}

func (a *Actor) GetMeleeWeapon() *Weapon {
    if a.equippedRightHand != nil {
        if weapon, ok := a.equippedRightHand.(*Weapon); ok {
            return weapon
        }
    }
    if a.equippedLeftHand != nil {
        if weapon, ok := a.equippedLeftHand.(*Weapon); ok {
            return weapon
        }
    }
    return nil
}

func (a *Actor) GetRangedWeapon() *Weapon {
    if a.equippedRanged != nil {
        return a.equippedRanged
    }
    return nil
}

func (a *Actor) AddStatusEffect(engine Engine, statusEffect StatusEffect) {
    if a.HasStatusEffectWithName(statusEffect.Name()) {
        a.statusEffects[statusEffect.Name()].OnReapply(engine, a)
        return
    }

    a.statusEffects[statusEffect.Name()] = statusEffect
    statusEffect.OnApply(engine, a)

    if attributeEffect, ok := statusEffect.(AttributeModifier); ok {
        for _, modifier := range attributeEffect.GetModifiers() {
            a.attributes.AddModifier(
                modifier.Attribute,
                string(statusEffect.Name()),
                modifier.GetValue,
                attributeEffect.IsExpired,
            )
        }
    }
    if derivedAttributeEffect, ok := statusEffect.(DerivedAttributeModifier); ok {
        for _, modifier := range derivedAttributeEffect.GetDerivedModifiers() {
            a.attributes.AddDerivedModifier(
                modifier.Attribute,
                string(statusEffect.Name()),
                modifier.GetValue,
                derivedAttributeEffect.IsExpired,
            )
        }
    }

}

func (a *Actor) IsSleeping() bool {
    return a.HasStatusEffectWithName(StatusEffectNameSleeping)
}

func (a *Actor) HasStatusEffectWithName(statusEffectName StatusEffectName) bool {
    _, hasEffect := a.statusEffects[statusEffectName]
    return hasEffect
}

func (a *Actor) CanBackstab(victim *Actor) bool {
    if !a.HasMeleeWeaponEquipped() {
        return false
    }
    weapon := a.GetMeleeWeapon()
    return weapon.IsDagger()
}

func (a *Actor) GetArmor(slot ArmorSlot) (*Armor, bool) {
    armor, ok := a.equippedArmor[slot]
    return armor, ok
}

func (a *Actor) SetDeathIcon(icon int32) {
    a.deathIcon = icon
}

func (a *Actor) RemoveStatusEffect(engine Engine, effectName StatusEffectName) {
    if existingEffect, ok := a.statusEffects[effectName]; ok {
        existingEffect.OnRemove(engine, a)
        delete(a.statusEffects, effectName)
    }
}

func (a *Actor) HasNamedItemsWithCount(name string, count int) bool {
    itemCount := 0
    for _, item := range a.inventory {
        if item.Name() == name {
            itemCount++
        }
    }
    return itemCount >= count
}

// GetTotalEncumbrance returns the total encumberance of the actor, including all equipped armor
// Min: 1, Max: 7
func (a *Actor) GetTotalEncumbrance() int {
    total := 0
    for _, item := range a.equippedArmor {
        total += item.GetEncumbrance()
    }
    return total
}

func (a *Actor) GetAbsoluteDifficultyByAttribute(attribute AttributeName) DifficultyLevel {
    attributeValue := a.attributes.GetAttribute(attribute)
    return DifficultyLevelFromInt(attributeValue - 5)
}

func (a *Actor) SetNPCEngagementZone(zone map[geometry.Point]int) {
    a.zoneOfEngagement = zone
}

func (a *Actor) IsInEngagementZone(point geometry.Point) bool {
    _, ok := a.zoneOfEngagement[point]
    return ok
}

func (a *Actor) GetEngagementZone() map[geometry.Point]int {
    return a.zoneOfEngagement
}

func (a *Actor) GetDebugInfos() []string {
    infos := []string{
        fmt.Sprintf("Name: %s", a.name),
        fmt.Sprintf("InternalName: %s", a.internalName),
        fmt.Sprintf("Health: %d/%d", a.health, a.maxHealth),
        fmt.Sprintf("Mana: %d", a.mana),
        fmt.Sprintf("XP: %d", a.experiencePoints),
        fmt.Sprintf("Level: %d", a.GetLevel()),
        fmt.Sprintf("Armor: %d", a.GetTotalArmor()),
        fmt.Sprintf("Melee: %d", a.GetMeleeDamage()),
        fmt.Sprintf("Ranged: %d", a.GetRangedDamage()),
        fmt.Sprintf("CombatFaction: %s", a.combatFaction),
        fmt.Sprintf("IsAggressive: %t", a.isAggressive),
    }
    if len(a.innateSkills) > 0 {
        infos = append(infos, "Skills:")
        for _, skill := range a.innateSkills {
            infos = append(infos, fmt.Sprintf("  %s", skill.Name()))
        }
    }
    return infos
}

func (a *Actor) GetAttributes() *AttributeHolder {
    return &a.attributes
}

func (a *Actor) HasStatusEffects() bool {
    return len(a.statusEffects) > 0
}

func (a *Actor) GetStatusEffectsTable() []string {
    var table []string
    var nameOrder []StatusEffectName
    for effectName, _ := range a.statusEffects {
        nameOrder = append(nameOrder, effectName)
    }
    sort.SliceStable(nameOrder, func(i, j int) bool {
        return nameOrder[i] < nameOrder[j]
    })
    for index, effectName := range nameOrder {
        effect := a.statusEffects[effectName]
        if index != 0 {
            table = append(table, "")
        }
        table = append(table, string(effect.Name()))
        table = append(table, util.LeftPadCountMulti(effect.Description(), 1)...)
    }
    return table
}

func (a *Actor) AddInnateSkill(skill *BaseAction) {
    a.innateSkills = append(a.innateSkills, skill)
}

func (a *Actor) CanAct() bool {
    return a.IsAlive() && !a.IsSleeping()
}
