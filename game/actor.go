package game

import (
    "Legacy/geometry"
    "Legacy/recfile"
    "Legacy/renderer"
    "Legacy/util"
    "fmt"
    "image/color"
    "io"
    "sort"
    "strconv"
)

type EquipmentSlot string

const (
    ArmorSlotHead     EquipmentSlot = "head"
    ArmorSlotTorso    EquipmentSlot = "torso"
    ItemSlotLeftHand  EquipmentSlot = "left hand"
    ItemSlotRightHand EquipmentSlot = "right hand"
    ItemSlotScroll    EquipmentSlot = "scroll"
)

type Actor struct {
    GameObject
    icon           int32
    iconFrameCount int
    name           string
    party          *Party
    dialogue       *Dialogue
    description    string

    equippedLeftHand  Wearable
    equippedRightHand Wearable
    equippedArmor     *Armor
    equippedHelmet    *Armor
    equippedScrolls   []*Scroll

    internalName string
    isHuman      bool

    inventory []Item

    // stats
    mana             int
    maxHealth        int
    health           int
    baseArmor        int
    baseMeleeDamage  int
    baseRangedDamage int
    experiencePoints int
    level            int

    skillset SkillSet
    buffs    map[BuffType][]Buff
    color    color.Color
    isTinted bool
}

func NewActor(name string, icon int32) *Actor {
    return &Actor{
        name:            name,
        icon:            icon,
        iconFrameCount:  1,
        health:          23,
        maxHealth:       23,
        isHuman:         true,
        inventory:       []Item{},
        skillset:        NewSkillSet(),
        level:           1,
        baseMeleeDamage: 5,
        buffs:           make(map[BuffType][]Buff),
        color:           color.White,
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
        name:            coreRecord["Name"],
        icon:            icon,
        health:          health,
        maxHealth:       health,
        description:     description,
        dialogue:        conversation,
        isHuman:         true,
        skillset:        NewSkillSet(),
        level:           1,
        baseMeleeDamage: 5,
        buffs:           make(map[BuffType][]Buff),
    }

    if recordsForInventory, hasInventory := actorData["Inventory"]; hasInventory && len(recordsForInventory) > 0 {
        inventory := recordsForInventory[0].ToValueList()
        newActor.inventory = toInventory(newActor, itemsFromStrings(inventory))
    }
    return newActor
}

func NewActorFromRecord(record recfile.Record) *Actor {
    a := &Actor{
        buffs:    make(map[BuffType][]Buff),
        skillset: NewSkillSet(), // TODO
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
        case "unlitIcon":
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
        case "level":
            a.level = field.AsInt()
        case "mana":
            a.mana = field.AsInt()
        case "baseArmor":
            a.baseArmor = field.AsInt()
        case "baseMelee":
            a.baseMeleeDamage = field.AsInt()
        case "baseRanged":
            a.baseRangedDamage = field.AsInt()
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
        recfile.Field{Name: "unlitIcon", Value: recfile.Int32Str(a.Icon(0))},
        recfile.Field{Name: "iconFrames", Value: strconv.Itoa(a.GetIconFrameCount())},

        recfile.Field{Name: "isHuman", Value: recfile.BoolStr(a.IsHuman())},

        recfile.Field{Name: "health", Value: strconv.Itoa(a.GetHealth())},
        recfile.Field{Name: "maxhealth", Value: strconv.Itoa(a.GetMaxHealth())},
        recfile.Field{Name: "xp", Value: strconv.Itoa(a.GetXP())},
        recfile.Field{Name: "level", Value: strconv.Itoa(a.GetLevel())},
        recfile.Field{Name: "mana", Value: strconv.Itoa(a.GetMana())},

        recfile.Field{Name: "baseArmor", Value: strconv.Itoa(a.GetBaseArmor())},
        recfile.Field{Name: "baseMelee", Value: strconv.Itoa(a.GetBaseMelee())},
        recfile.Field{Name: "baseRanged", Value: strconv.Itoa(a.GetBaseRanged())},

        recfile.Field{Name: "description", Value: a.description},
        // TODO: dialogue state

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
    case *Weapon:
        weapon := item.(*Weapon)
        weapon.SetWearer(nil)
        if a.equippedLeftHand == weapon {
            a.equippedLeftHand = nil
        }
        if a.equippedRightHand == weapon {
            a.equippedRightHand = nil
        }
    case *Armor:
        armor := item.(*Armor)
        armor.SetWearer(nil)
        switch armor.slot {
        case ArmorSlotHead:
            if a.equippedHelmet == armor {
                a.equippedHelmet = nil
            }
        case ArmorSlotTorso:
            if a.equippedArmor == armor {
                a.equippedArmor = nil
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
}

func (a *Actor) SetParty(party *Party) {
    a.party = party
}

func toInventory(actor *Actor, items []Item) []Item {
    for _, item := range items {
        item.SetHolder(actor)
    }
    return items
}

func cleanInventory(inventory []string) []string {
    prefixToRemove := " - "

    var clean []string
    for _, itemString := range inventory {
        if len(itemString) == 0 {
            continue
        }
        if itemString[0:len(prefixToRemove)] == prefixToRemove {
            clean = append(clean, itemString[len(prefixToRemove):])
        } else {
            clean = append(clean, itemString)
        }
    }
    return clean
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
        return 24
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

func (a *Actor) GetDetails(engine Engine) []string {

    leftHand := "nothing"
    if a.equippedLeftHand != nil {
        leftHand = a.equippedLeftHand.Name()
    }
    rightHand := "nothing"
    if a.equippedRightHand != nil {
        rightHand = a.equippedRightHand.Name()
    }
    armor := "nothing"
    if a.equippedArmor != nil {
        armor = fmt.Sprintf("%s (%d)", a.equippedArmor.Name(), a.equippedArmor.protection)
    }
    helmet := "nothing"
    if a.equippedHelmet != nil {
        helmet = fmt.Sprintf("%s (%d)", a.equippedHelmet.Name(), a.equippedHelmet.protection)
    }
    _, xpNeeded := engine.CanLevelUp(a)
    tableData := []util.TableRow{
        {Label: "Level", Columns: []string{strconv.Itoa(a.level)}},
        {Label: "XP", Columns: []string{strconv.Itoa(a.experiencePoints)}},
        {Label: "Next Lvl.", Columns: []string{strconv.Itoa(xpNeeded)}},
        {Label: "Health", Columns: []string{fmt.Sprintf("%d/%d", a.health, a.maxHealth)}},
        {Label: "Mana", Columns: []string{fmt.Sprintf("%d", a.mana)}},
        {Label: "Armor", Columns: []string{fmt.Sprintf("%d", a.GetTotalArmor())}},
        {Label: "Melee Dmg.", Columns: []string{fmt.Sprintf("%d", a.GetMeleeDamage())}},
        {Label: "Left Hand", Columns: []string{leftHand}},
        {Label: "Right Hand", Columns: []string{rightHand}},
        {Label: "Armor", Columns: []string{armor}},
        {Label: "Helmet", Columns: []string{helmet}},
    }
    return util.TableLayout(tableData)
}

func (a *Actor) LookDescription() []string {
    healthString := "healthy"
    description := renderer.AutoLayout(a.description, 32)
    // wearables
    if a.equippedArmor != nil || a.equippedHelmet != nil {
        description = append(description, "", "The person is wearing:")
        if a.equippedArmor != nil {
            description = append(description, fmt.Sprintf("  %s", a.equippedArmor.Name()))
        }
        if a.equippedHelmet != nil {
            description = append(description, fmt.Sprintf("  %s", a.equippedHelmet.Name()))
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

    description = append(description, "")
    description = append(description, fmt.Sprintf("The person looks %s.", healthString))

    return description
}

func (a *Actor) GetContextActions(engine Engine) []renderer.MenuItem {
    var items []renderer.MenuItem
    if a != engine.GetAvatar() {
        talkTo := renderer.MenuItem{
            Text: "Talk",
            Action: func() {
                engine.StartConversation(a, a.GetDialogue())
            },
        }
        lookAt := renderer.MenuItem{
            Text: "Look",
            Action: func() {
                engine.ShowScrollableText(a.LookDescription(), color.White, false)
            },
        }
        steal := renderer.MenuItem{
            Text:   "Steal",
            Action: func() { engine.OpenPickpocketMenu(a) },
        }
        attack := renderer.MenuItem{
            Text: "Attack",
            Action: func() {
                engine.StartCombat(a)
            },
        }
        push := renderer.MenuItem{
            Text: "Prod",
            Action: func() {
                if engine.GetAvatar().IsRightNextTo(a) {
                    engine.ProdActor(engine.GetAvatar(), a)
                }
            },
        }
        items = append(items, talkTo, lookAt, attack)
        if engine.GetAvatar().IsRightNextTo(a) {
            items = append(items, steal, push)
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
    return geometry.DistanceManhattan(ownPos, otherPos) == 1
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

func (a *Actor) RemoveItem(item Item) {
    for i, inventoryItem := range a.inventory {
        if inventoryItem == item {
            a.inventory = append(a.inventory[:i], a.inventory[i+1:]...)
            break
        }
    }
    if item.GetHolder() == a {
        item.SetHolder(nil)
    }
}

func (a *Actor) CanEquip(item Item) bool {
    switch item.(type) {
    case *Armor:
        return true
    case *Scroll:
        return true
    case *Weapon:
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
    switch item.(type) {
    case *Armor:
        armor := item.(*Armor)
        a.EquipArmor(armor)
    case *Scroll:
        scroll := item.(*Scroll)
        a.EquipScroll(scroll)
    case *Weapon:
        weapon := item.(*Weapon)
        a.EquipWeapon(weapon)
    }
}

func (a *Actor) EquipArmor(armor *Armor) {
    if !a.CanEquip(armor) {
        return
    }
    armor.Unequip()
    switch armor.slot {
    case ArmorSlotHead:
        if a.equippedHelmet != nil {
            a.equippedHelmet.Unequip()
        }
        a.equippedHelmet = armor
        armor.SetWearer(a)
    case ArmorSlotTorso:
        if a.equippedArmor != nil {
            a.equippedArmor.Unequip()
        }
        a.equippedArmor = armor
        armor.SetWearer(a)
    }
}
func (a *Actor) EquipScroll(scroll *Scroll) {
    if !a.CanEquip(scroll) {
        return
    }
    scroll.Unequip()

    a.equippedScrolls = append(a.equippedScrolls, scroll)
    scroll.SetWearer(a)
}
func (a *Actor) EquipWeapon(weapon *Weapon) {
    if !a.CanEquip(weapon) {
        return
    }
    weapon.Unequip()

    if weapon.IsTwoHanded() {
        if a.equippedLeftHand != nil {
            a.equippedLeftHand.Unequip()
        }
        a.equippedLeftHand = weapon
    }

    if a.equippedRightHand != nil {
        a.equippedRightHand.Unequip()
    }
    a.equippedRightHand = weapon
    weapon.SetWearer(a)
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

func (a *Actor) FullRest() {
    if a.health < a.maxHealth {
        a.health = a.maxHealth
    }
    a.ClearBuffs()
}

func (a *Actor) Damage(amount int) {
    a.health -= amount
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
    armorSum := 0
    if a.equippedArmor != nil {
        armorSum += a.equippedArmor.protection
    }
    if a.equippedHelmet != nil {
        armorSum += a.equippedHelmet.protection
    }
    buffs := a.GetDefenseBuffBonus()
    return a.baseArmor + armorSum + buffs
}

func (a *Actor) GetMeleeDamage() int {
    buffs := a.GetOffenseBuffBonus()
    baseDamage := a.baseMeleeDamage + buffs
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
    return 5 // TODO
}

func (a *Actor) SetHealth(health int) {
    a.health = health
}

func (a *Actor) SetMana(mana int) {
    a.mana = mana
}

func (a *Actor) StripGear() {
    if a.equippedArmor != nil {
        a.equippedArmor.Unequip()
    }
    if a.equippedHelmet != nil {
        a.equippedHelmet.Unequip()
    }
}

func (a *Actor) AutoEquip(engine Engine) {
    var bestArmor *Armor
    var bestHelmet *Armor
    var bestWeapon *Weapon

    allEquipment := engine.GetPartyEquipment()
    for _, item := range allEquipment {
        if a.CanEquip(item) {
            if armorPiece, ok := item.(*Armor); ok {
                if armorPiece.IsEquipped() {
                    continue
                }
                if armorPiece.GetSlot() == ArmorSlotHead {
                    if armorPiece.IsBetterThan(bestHelmet) {
                        bestHelmet = armorPiece
                    }
                } else if armorPiece.GetSlot() == ArmorSlotTorso {
                    if armorPiece.IsBetterThan(bestArmor) {
                        bestArmor = armorPiece
                    }
                }
            } else if weapon, ok := item.(*Weapon); ok {
                if weapon.IsEquipped() {
                    continue
                }
                if weapon.IsBetterThan(bestWeapon) {
                    bestWeapon = weapon
                }
            }
        }
    }
    if bestArmor != nil {
        a.Equip(bestArmor)
    }
    if bestHelmet != nil {
        a.Equip(bestHelmet)
    }
    if bestWeapon != nil {
        a.Equip(bestWeapon)
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

func (a *Actor) GetSlotForItem(item Wearable) EquipmentSlot {
    switch {
    case a.equippedLeftHand == item:
        return ItemSlotLeftHand
    case a.equippedRightHand == item:
        return ItemSlotRightHand
    case a.equippedArmor == item:
        return ArmorSlotTorso
    case a.equippedHelmet == item:
        return ArmorSlotHead
    }
    return ""
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
    return a.level
}

func (a *Actor) GetMana() int {
    return a.mana
}
func (a *Actor) GetIconFrameCount() int {
    return a.iconFrameCount
}

func (a *Actor) GetBaseArmor() int {
    return a.baseArmor
}

func (a *Actor) GetBaseMelee() int {
    return a.baseMeleeDamage
}

func (a *Actor) GetBaseRanged() int {
    return a.baseRangedDamage
}

func (a *Actor) GetXPForKilling() int {
    return a.level * 10
}

func (a *Actor) AddAllSkills() {
    a.skillset.AddAll()
}

func (a *Actor) AddBuff(name string, buffType BuffType, strength int) {
    a.buffs[buffType] = append(a.buffs[buffType], Buff{
        Name:     name,
        Strength: strength,
    })
}

func (a *Actor) ClearBuffs() {
    a.buffs = make(map[BuffType][]Buff)
}

func (a *Actor) GetDefenseBuffBonus() int {
    bonusFromBuffs := 0
    for _, buff := range a.buffs[BuffTypeDefense] {
        bonusFromBuffs += buff.Strength
    }
    return bonusFromBuffs
}

func (a *Actor) GetOffenseBuffBonus() int {
    bonusFromBuffs := 0
    for _, buff := range a.buffs[BuffTypeOffense] {
        bonusFromBuffs += buff.Strength
    }
    return bonusFromBuffs
}

func (a *Actor) HasOffenseBuffs() bool {
    if len(a.buffs[BuffTypeOffense]) == 0 {
        return false
    }

    for _, buff := range a.buffs[BuffTypeOffense] {
        if buff.Strength > 0 {
            return true
        }
    }
    return false
}

func (a *Actor) HasDefenseBuffs() bool {
    if len(a.buffs[BuffTypeDefense]) == 0 {
        return false
    }

    for _, buff := range a.buffs[BuffTypeDefense] {
        if buff.Strength > 0 {
            return true
        }
    }
    return false
}

func (a *Actor) GetDefenseBuffsString() []string {
    var rows []util.TableRow
    for _, buff := range a.buffs[BuffTypeDefense] {
        rows = append(rows, util.TableRow{
            Label: fmt.Sprintf("+%d", buff.Strength), Columns: []string{buff.Name},
        })
    }

    return util.TableLayout(rows)
}

func (a *Actor) GetOffenseBuffsString() []string {
    var rows []util.TableRow
    for _, buff := range a.buffs[BuffTypeOffense] {
        rows = append(rows, util.TableRow{
            Label: fmt.Sprintf("+%d", buff.Strength), Columns: []string{buff.Name},
        })
    }

    return util.TableLayout(rows)
}
func (a *Actor) GetOffenseBuffs() []Buff {
    return a.buffs[BuffTypeOffense]
}

func (a *Actor) GetDefenseBuffs() []Buff {
    return a.buffs[BuffTypeDefense]
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

func (a *Actor) HasNegativeBuffs() bool {
    for _, buffs := range a.buffs {
        for _, buff := range buffs {
            if buff.Strength < 0 {
                return true
            }
        }
    }
    return false
}

func (a *Actor) BuffsAsStringTable() []string {
    off := a.GetOffenseBuffs()
    def := a.GetDefenseBuffs()
    var rows []util.TableRow
    if len(off) != 0 {
        rows = append(rows, util.TableRow{Label: "Offense", Columns: []string{""}})
        rows = append(rows, util.TableRow{Label: "-------", Columns: []string{""}})
        for _, b := range off {
            rows = append(rows, util.TableRow{Label: b.Name, Columns: []string{strconv.Itoa(b.Strength)}})
        }

    }
    if len(def) != 0 {
        if len(off) != 0 {
            rows = append(rows, util.TableRow{Label: "", Columns: []string{""}})
        }
        rows = append(rows, util.TableRow{Label: "Defense", Columns: []string{""}})
        rows = append(rows, util.TableRow{Label: "-------", Columns: []string{""}})
        for _, b := range def {
            rows = append(rows, util.TableRow{Label: b.Name, Columns: []string{strconv.Itoa(b.Strength)}})
        }
    }
    if len(rows) == 0 {
        return []string{"No buffs"}
    }
    return util.TableLayout(rows)
}

func (a *Actor) SetIcon(icon int32) {
    a.icon = icon
}

func (a *Actor) SetVendorInventory(items []Item) {
    a.inventory = append(a.inventory, items...)
}

func (a *Actor) GetEquippedSpells() []*Spell {
    var spells []*Spell
    for _, scroll := range a.equippedScrolls {
        spells = append(spells, scroll.spell)
    }
    return spells

}
