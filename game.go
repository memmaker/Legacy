package main

import (
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/renderer"
    "fmt"
    "image/color"
)

func (g *GridEngine) openPartyMenu() {
    partyOptions := []renderer.MenuItem{
        {
            Text:   "Search",
            Action: g.searchForHiddenObjects,
        },
        {
            Text:   "Inventory",
            Action: g.openPartyInventory,
        },
        {
            Text:   "Magic",
            Action: g.openSpellMenu,
        },
        {
            Text:   "Rest",
            Action: g.TryRestParty,
        },
        {
            Text: "Attack",
            Action: func() {
                g.Print("Not implemented yet")
            },
        },
    }
    if g.playerParty.HasFollowers() {
        partyOptions = append(partyOptions, renderer.MenuItem{
            Text: "Split",
            Action: func() {
                g.openMenu(g.playerParty.GetSplitActions(g))
            },
        })
        if g.splitControlled != nil {
            partyOptions = append(partyOptions, renderer.MenuItem{
                Text:   "Join",
                Action: g.TryJoinParty,
            })
        }
    }

    g.openMenu(partyOptions)
}

func (g *GridEngine) openCharDetails(partyIndex int) {
    actor := g.playerParty.GetMember(partyIndex)
    if actor != nil {
        g.showModal(actor.GetDetails())
    }
}

func (g *GridEngine) openPartyInventory() {
    //header := []string{"Inventory", "---------"}
    partyInventory := g.playerParty.GetStackedInventory()
    if len(partyInventory) == 0 {
        g.showModal([]string{"Your party has no items."})
        return
    }
    var menuItems []renderer.MenuItem
    for _, i := range partyInventory {
        itemStack := i
        menuItems = append(menuItems, renderer.MenuItem{
            Text: fmt.Sprintf("%s (%d)", itemStack[0].Name(), len(itemStack)),
            Action: func() {
                g.openMenu(itemStack[0].GetContextActions(g))
            },
        })
    }
    g.openMenu(menuItems)
}

func (g *GridEngine) showModal(text []string) {
    if len(text) == 0 {
        return
    }
    g.modalElement = renderer.NewScrollableTextWindowWithAutomaticSize(g.gridRenderer, text)
}

func (g *GridEngine) ShowScrollableText(text []string, color color.Color) {
    if len(text) == 0 {
        return
    }
    modal := renderer.NewScrollableTextWindowWithAutomaticSize(g.gridRenderer, text)
    modal.SetTextColor(color)
    g.modalElement = modal
}

func (g *GridEngine) PickUpItem(item game.Item) {
    g.playerParty.AddItem(item)
    g.currentMap.RemoveItem(item)
    g.Print(fmt.Sprintf("Taken \"%s\"", item.Name()))
    g.updateContextActions()
}
func (g *GridEngine) moveItemToParty(item game.Item, container game.ItemContainer) {
    container.RemoveItem(item)
    g.playerParty.AddItem(item)
    g.Print(fmt.Sprintf("Taken \"%s\"", item.Name()))
    g.updateContextActions()
}
func (g *GridEngine) DropItem(item game.Item) {
    g.playerParty.RemoveItem(item)
    destPos := g.avatar.Pos()
    if g.TryPlaceItem(item, destPos) {
        g.Print(fmt.Sprintf("Dropped \"%s\"", item.Name()))
        g.updateContextActions()
    }
}

func (g *GridEngine) TryPlaceItem(item game.Item, destPos geometry.Point) bool {
    if g.currentMap.IsItemAt(destPos) {
        freeCells := g.currentMap.GetFreeCellsForDistribution(g.avatar.Pos(), 1, func(p geometry.Point) bool {
            return g.currentMap.Contains(p) && g.currentMap.IsWalkableFor(p, g.avatar) && !g.currentMap.IsItemAt(p)
        })
        if len(freeCells) > 0 {
            destPos = freeCells[0]
        } else {
            g.Print(fmt.Sprintf("No space to drop \"%s\"", item.Name()))
            return false
        }
    }

    g.currentMap.AddItem(item, destPos)
    return true
}

func (g *GridEngine) SwitchAvatarTo(actor *game.Actor) {
    if !g.playerParty.IsMember(actor) {
        g.Print(fmt.Sprintf("\"%s\" is not in your party", actor.Name()))
        return
    }
    g.splitControlled = actor
    g.onAvatarMovedAlone()
}
func (g *GridEngine) AddToParty(npc *game.Actor) {
    if g.playerParty.IsFull() {
        g.Print(fmt.Sprintf("No room for \"%s\"", npc.Name()))
        return
    } else if g.playerParty.IsMember(npc) {
        g.Print(fmt.Sprintf("\"%s\" is already in your party", npc.Name()))
        return
    }
    g.playerParty.AddMember(npc)
}

func (g *GridEngine) TryJoinParty() {
    for _, member := range g.playerParty.GetMembers() {
        if member == g.avatar {
            continue
        }
        if !member.IsNextTo(g.avatar) {
            g.Print(fmt.Sprintf("\"%s\" is not next to you.", member.Name()))
            return
        }
    }
    g.splitControlled = nil
}

func (g *GridEngine) openVendorMenu(npc *game.Actor) {
    itemsToSell := npc.GetItemsToSell()
    if len(itemsToSell) == 0 {
        g.openSpeechWindow(npc, []string{"Unfortunately, I have nothing left to sell."}, func() {})
        return
    }
    var menuItems []renderer.MenuItem
    for _, i := range itemsToSell {
        offer := i
        itemLine := fmt.Sprintf("%s (%d)", offer.Item.Name(), offer.Price)
        menuItems = append(menuItems, renderer.MenuItem{
            Text: itemLine,
            Action: func() {
                g.TryBuyItem(npc, offer)
            },
        })
    }
    g.openMenu(menuItems)
}

func (g *GridEngine) TryBuyItem(npc *game.Actor, offer game.SalesOffer) {
    if g.playerParty.GetGold() < offer.Price {
        g.openSpeechWindow(npc, []string{"You don't have enough gold."}, func() {})
        return
    }
    npc.RemoveItem(offer.Item)
    npc.AddGold(offer.Price)

    g.playerParty.RemoveGold(offer.Price)
    g.playerParty.AddItem(offer.Item)
    g.openSpeechWindow(npc, []string{"Thank you for your business."}, func() {})
}

func (g *GridEngine) TryRestParty() {
    if g.playerParty.TryRest() {
        g.openSpeechWindow(g.GetAvatar(), []string{"You have eaten some food", "and rested the night.", "Your party has been healed."}, func() {})
    } else {
        g.Print("Not enough food to rest.")
    }
}

func (g *GridEngine) searchForHiddenObjects() {
    source := g.GetAvatar().Pos()
    neighbors := g.currentMap.NeighborsAll(source, func(p geometry.Point) bool {
        if !g.currentMap.Contains(p) {
            return false
        }
        return true
    })
    foundSomething := false
    for _, neighbor := range neighbors {
        if g.currentMap.IsSecretDoorAt(neighbor) {
            wallTile := g.currentMap.GetCell(neighbor).TileType
            g.currentMap.SetTile(neighbor, wallTile.WithIsWalkable(true))
            g.currentMap.AddObject(game.NewDoor(), neighbor)
            foundSomething = true
            g.flags.IncrementFlag("found_hidden_things")
        } else {
            if g.currentMap.IsObjectAt(neighbor) {
                objectAt := g.currentMap.ObjectAt(neighbor)
                if objectAt != nil && objectAt.IsHidden() {
                    objectAt.SetHidden(false)
                    foundSomething = true
                    g.flags.IncrementFlag("found_hidden_things")
                }
            } else if g.currentMap.IsItemAt(neighbor) {
                itemAt := g.currentMap.ItemAt(neighbor)
                if itemAt != nil && itemAt.IsHidden() {
                    itemAt.SetHidden(false)
                    foundSomething = true
                    g.flags.IncrementFlag("found_hidden_things")
                }
            } else if g.currentMap.IsActorAt(neighbor) {
                actorAt := g.currentMap.GetActor(neighbor)
                if actorAt != nil && actorAt.IsHidden() {
                    actorAt.SetHidden(false)
                    foundSomething = true
                    g.flags.IncrementFlag("found_hidden_things")
                }
            }
        }
    }

    if foundSomething {
        g.Print("Fascinating..")
    } else {
        g.Print("You find nothing.")
    }
}

func (g *GridEngine) PartyHasKey(key string) bool {
    return g.playerParty.HasKey(key)
}

func (g *GridEngine) AddFood(amount int) {
    g.playerParty.AddFood(amount)
}

func (g *GridEngine) AddGold(amount int) {
    g.playerParty.AddGold(amount)
}

func (g *GridEngine) GetPartySize() int {
    return len(g.playerParty.GetMembers())
}

func (g *GridEngine) CreateLoot(level int, lootType game.Loot) []game.Item {
    //TODO
    return []game.Item{
        game.NewKey("Loot Key", "key_loot", color.White),
    }
}

func (g *GridEngine) ShowContainer(container game.ItemContainer) {
    var menuItems []renderer.MenuItem
    for _, i := range container.GetItems() {
        item := i
        menuItems = append(menuItems, renderer.MenuItem{
            Text: i.Name(),
            Action: func() {
                g.moveItemToParty(item, container)
            },
        })
    }
    g.openMenu(menuItems)
}
