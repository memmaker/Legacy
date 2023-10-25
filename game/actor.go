package game

import (
    "Legacy/geometry"
    "Legacy/renderer"
    "fmt"
    "image/color"
    "strconv"
)

type Actor struct {
    GameObject
    icon   int
    name   string
    Health int
    party  *Party
}

func (a *Actor) SetParty(party *Party) {
    a.party = party
}
func NewActor(name string, icon int) *Actor {
    return &Actor{
        name:   name,
        icon:   icon,
        Health: 10,
    }
}

func (a *Actor) Icon() int {
    return a.icon
}

func (a *Actor) Name() string {
    return a.name
}

func (a *Actor) GetDetails() []string {
    return []string{
        a.name,
        "Health: " + strconv.Itoa(a.Health),
    }
}

func (a *Actor) LookDescription() []string {
    healthString := "healthy"
    if a.Health < 5 {
        healthString = "wounded"
    }
    return []string{
        "A person is standing here.",
        fmt.Sprintf("It looks %s.", healthString),
    }
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
