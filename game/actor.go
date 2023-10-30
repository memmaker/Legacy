package game

import (
    "Legacy/geometry"
    recfile "Legacy/recfile"
    "Legacy/renderer"
    "bufio"
    "fmt"
    "image/color"
    "io"
    "regexp"
)

type EquipmentSlot string

const (
    EquipmentSlotWeapon EquipmentSlot = "weapon"
    EquipmentSlotArmor  EquipmentSlot = "armor"
    EquipmentSlotHelmet EquipmentSlot = "helmet"
)

type Actor struct {
    GameObject
    icon              int
    iconFrameCount    int
    name              string
    party             *Party
    dialogue          *Dialogue
    description       []string
    equippedLeftHand  Item
    equippedRightHand Item
    equippedArmor     *Armor
    equippedHelmet    *Armor
    equippedScrolls   []*Scroll
    internalName      string
    mana              int
    Health            int
    isHuman           bool
    maxHealth         int
    baseArmor         int

    inventory []Item
}

func (a *Actor) Unequip(item Item) {
    switch item.(type) {
    case *Armor:
        armor := item.(*Armor)
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

func readUntil(reader *bufio.Scanner, stop *regexp.Regexp) []string {
    var lines []string
    for reader.Scan() {
        line := reader.Text()
        if len(line) == 0 {
            continue
        }
        if stop.MatchString(line) {
            break
        }
        lines = append(lines, line)
    }
    return lines
}
func NewActorFromFile(file io.Reader, icon int) *Actor {
    // line by line
    scanner := bufio.NewScanner(file)
    recordReader := recfile.NewReader()
    conversationReader := NewConversationReader()

    // first marker "# Description"
    descriptionRegex := regexp.MustCompile(`^# Description$`)
    inventoryRegex := regexp.MustCompile(`^# Inventory$`)
    conversationRegex := regexp.MustCompile(`^# Conversation$`)

    npcCoreData := readUntil(scanner, descriptionRegex)
    npcDescription := readUntil(scanner, inventoryRegex)
    npcInventory := readUntil(scanner, conversationRegex) // TODO: npcInventory :=

    for scanner.Scan() {
        conversationReader.read(scanner.Text())
    }
    conversation := conversationReader.end()

    coreRecord := recordReader.ReadLines(npcCoreData)[0].ToMap()
    health, _ := coreRecord.GetInt("Health")

    newActor := &Actor{
        name:        coreRecord["Name"],
        icon:        icon,
        Health:      health,
        description: npcDescription,
        dialogue:    NewDialogue(conversation),
        isHuman:     true,
    }
    newActor.inventory = toInventory(newActor, itemsFromStrings(cleanInventory(npcInventory)))
    return newActor
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

func NewActor(name string, icon int) *Actor {
    return &Actor{
        name:           name,
        icon:           icon,
        iconFrameCount: 1,
        Health:         23,
        maxHealth:      23,
        isHuman:        true,
        inventory: []Item{
            NewKeyFromImportance("Inventory Key", "fake_key", 1),
        },
    }
}
func (a *Actor) SetHuman(isHuman bool) {
    a.isHuman = isHuman
}
func (a *Actor) Icon(tick uint64) int {
    if !a.IsAlive() {
        return 24
    }
    if a.iconFrameCount == 1 {
        return a.icon
    }
    delays := tick / 20
    return a.icon + int(delays%uint64(a.iconFrameCount))
}

func (a *Actor) Name() string {
    return a.name
}

func (a *Actor) GetDetails() []string {

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

    details := []string{
        a.name,
        fmt.Sprintf("Health    : %d", a.Health),
        fmt.Sprintf("Mana      : %d", a.mana),
        fmt.Sprintf("Armor     : %d", a.GetTotalArmor()),
        fmt.Sprintf("Left Hand : %s", leftHand),
        fmt.Sprintf("Right Hand: %s", rightHand),
        fmt.Sprintf("Armor     : %s", armor),
        fmt.Sprintf("Helmet    : %s", helmet),
    }

    return details
}

func (a *Actor) LookDescription() []string {
    healthString := "healthy"
    if a.Health < 5 {
        healthString = "wounded"
    }
    return append(a.description, fmt.Sprintf("The person looks %s.", healthString))
}

func (a *Actor) GetContextActions(engine Engine) []renderer.MenuItem {
    var items []renderer.MenuItem
    if a != engine.GetAvatar() {
        talkTo := renderer.MenuItem{
            Text: "Talk",
            Action: func() {
                engine.StartConversation(a)
            },
        }
        lookAt := renderer.MenuItem{
            Text: "Look",
            Action: func() {
                engine.ShowColoredText(a.LookDescription(), color.White, true)
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
        items = append(items, talkTo, lookAt, attack)
        if engine.GetAvatar().IsRightNextTo(a) {
            items = append(items, steal)
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
    return []SalesOffer{
        SalesOffer{
            Item:  NewKeyFromImportance("Sell Key", "fake_key", 1),
            Price: 10,
        },
    }
}

func (a *Actor) RemoveItem(item Item) {
    for i, inventoryItem := range a.inventory {
        if inventoryItem == item {
            a.inventory = append(a.inventory[:i], a.inventory[i+1:]...)
            break
        }
    }
}

func (a *Actor) CanEquip(item Item) bool {
    switch item.(type) {
    case *Armor:
        return true
    case *Scroll:
        return true
    }
    return false
}

func (a *Actor) Equip(item Item) {
    if !a.CanEquip(item) {
        return
    }
    switch item.(type) {
    case *Armor:
        armor := item.(*Armor)
        armor.Unequip(item)
        switch armor.slot {
        case ArmorSlotHead:
            a.equippedHelmet = armor
            armor.SetWearer(a)
        case ArmorSlotTorso:
            a.equippedArmor = armor
            armor.SetWearer(a)
        }
    case *Scroll:
        scroll := item.(*Scroll)
        scroll.GetWearer().Unequip(item)
        a.equippedScrolls = append(a.equippedScrolls, scroll)
        scroll.SetWearer(a)
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

func (a *Actor) FullRest() {
    if a.Health < a.maxHealth {
        a.Health = a.maxHealth
    }
}

func (a *Actor) Damage(amount int) {
    a.Health -= amount
}

func (a *Actor) GetItemsToSteal() []Item {
    return a.inventory
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
    return a.baseArmor + armorSum
}

func (a *Actor) IsAlive() bool {
    return a.Health > 0
}

func (a *Actor) DropInventory() []Item {
    dropped := a.inventory
    a.inventory = []Item{}
    return dropped
}
