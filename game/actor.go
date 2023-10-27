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
    equippedArmor     Item // TODO: change type to Armor
    equippedScrolls   []Scroll
    internalName      string
    mana              int
    Health            int
    isHuman           bool
    maxHealth         int
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
    readUntil(scanner, conversationRegex) // TODO: npcInventory :=

    for scanner.Scan() {
        conversationReader.read(scanner.Text())
    }
    conversation := conversationReader.end()

    coreRecord := recordReader.ReadLines(npcCoreData)[0].ToMap()
    health, _ := coreRecord.GetInt("Health")

    return &Actor{
        name:        coreRecord["Name"],
        icon:        icon,
        Health:      health,
        description: npcDescription,
        dialogue:    NewDialogue(conversation),
        isHuman:     true,
    }
}

func NewActor(name string, icon int) *Actor {
    return &Actor{
        name:           name,
        icon:           icon,
        iconFrameCount: 1,
        Health:         23,
        maxHealth:      23,
        isHuman:        true,
    }
}
func (a *Actor) SetHuman(isHuman bool) {
    a.isHuman = isHuman
}
func (a *Actor) Icon(tick uint64) int {
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
        armor = a.equippedArmor.Name()
    }

    details := []string{
        a.name,
        fmt.Sprintf("Health    : %d", a.Health),
        fmt.Sprintf("Mana      : %d", a.mana),
        fmt.Sprintf("Left Hand : %s", leftHand),
        fmt.Sprintf("Right Hand: %s", rightHand),
        fmt.Sprintf("Armor     : %s", armor),
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
            Text: fmt.Sprintf("Talk to \"%s\"", a.name),
            Action: func() {
                engine.StartConversation(a)
            },
        }
        lookAt := renderer.MenuItem{
            Text: fmt.Sprintf("Look at \"%s\"", a.name),
            Action: func() {
                engine.ShowScrollableText(a.LookDescription(), color.White)
            },
        }

        items = append(items, talkTo, lookAt)
    }
    return items
}

func (a *Actor) HasKey(key string) bool {
    if a.party != nil {
        return a.party.HasKey(key)
    }
    return false
}

func (a *Actor) IsNextTo(other *Actor) bool {
    ownPos := a.Pos()
    otherPos := other.Pos()
    return geometry.DistanceManhattan(ownPos, otherPos) <= 2
}

type SalesOffer struct {
    Item  Item
    Price int
}

func (a *Actor) GetItemsToSell() []SalesOffer {
    return []SalesOffer{
        SalesOffer{
            Item:  NewKey("Fake Key", "fake_key", color.White),
            Price: 10,
        },
    }
}

func (a *Actor) RemoveItem(item Item) {
    // TODO: implement
}

func (a *Actor) AddGold(price int) {
    // TODO: implement
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
