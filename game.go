package main

import (
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/renderer"
    "Legacy/ui"
    "Legacy/util"
    "fmt"
    "image/color"
    "math/rand"
    "sort"
    "strconv"
)

func (g *GridEngine) openPartyMenu() {
    partyOptions := []util.MenuItem{
        {
            Text:   "Search",
            Action: g.searchForHiddenObjects,
        },
        {
            Text:   "Inventory",
            Action: g.openExtendedInventory,
        },
        {
            Text:   "Keys",
            Action: g.openKeyInventory,
        },
        {
            Text:   "Ranged",
            Action: g.combatManager.PlayerStartsRangedAttack,
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
            Text:   "Journal",
            Action: g.openJournal,
        },
        {
            Text:   "Message log",
            Action: g.openPrintLog,
        },
    }
    if g.playerParty.HasFollowers() {
        partyRanged := util.MenuItem{
            Text:   "Party Ranged",
            Action: g.combatManager.OrchestratedRangedAttack,
        }
        //insert at index 3
        partyOptions = append(partyOptions[:3], append([]util.MenuItem{partyRanged}, partyOptions[3:]...)...)
        partyOptions = append(partyOptions, util.MenuItem{
            Text: "Split",
            Action: func() {
                g.OpenMenu(g.playerParty.GetSplitActions(g))
            },
        })
        if g.playerParty.IsSplit() {
            partyOptions = append(partyOptions, util.MenuItem{
                Text:   "Join",
                Action: g.TryJoinParty,
            })
        }
        partyOptions = append(partyOptions, util.MenuItem{
            Text:   "Dismiss",
            Action: g.openDismissMenu,
        })
    }

    g.OpenMenu(partyOptions)
}

func (g *GridEngine) openCharDetails(actor *game.Actor) {
    if actor == nil {
        return
    }
    g.CloseAllModals()
    window := g.ShowFixedFormatText(actor.GetDetails(g))
    window.SetTitle(actor.Name())
}
func (g *GridEngine) openCharSkills(partyIndex int) {
    actor := g.playerParty.GetMember(partyIndex)
    if actor != nil {
        skills := actor.GetSkills()
        window := g.ShowFixedFormatText(skills.AsStringTable())
        window.SetTitle(actor.Name())
    }
}

func (g *GridEngine) openCharBuffs(partyIndex int) {
    actor := g.playerParty.GetMember(partyIndex)
    if actor != nil {
        window := g.ShowFixedFormatText(actor.BuffsAsStringTable())
        window.SetTitle(actor.Name())
    }
}
func (g *GridEngine) openExtendedInventory() {
    g.OpenPartyInventoryOnPage(0)
}
func (g *GridEngine) OpenPartyInventoryOnPage(page int) {
    g.CloseAllModals()
    if !g.playerParty.HasItems() {
        g.ShowText([]string{"Your party has no items."})
        return
    }
    inventoryWindow := ui.NewInventoryWindow(g, g.gridRenderer)
    inventoryWindow.SwitchPageTo(ui.InventoryPage(page))
    g.PushModal(inventoryWindow)
}

func (g *GridEngine) openKeyInventory() {
    playerKeys := g.playerParty.GetKeys()
    if len(playerKeys) == 0 {
        g.Print("Your party has no keys.")
        return
    }
    g.CloseAllModals()
    keyListByLocation := make(map[string][]*game.Key)
    var listOfLocations []string
    for _, i := range playerKeys {
        key := i
        keyLocation := key.KeySource()
        if _, ok := keyListByLocation[keyLocation]; !ok {
            keyListByLocation[keyLocation] = []*game.Key{}
            listOfLocations = append(listOfLocations, keyLocation)
        }
        keyListByLocation[keyLocation] = append(keyListByLocation[keyLocation], key)
    }
    sort.Strings(listOfLocations)
    var asText []string
    for i, location := range listOfLocations {
        if i > 0 {
            asText = append(asText, "")
        }
        asText = append(asText, fmt.Sprintf("%s:", location))
        for _, key := range keyListByLocation[location] {
            prefix := "  "
            if g.playerParty.HasUsedKey(key.GetKeyString()) {
                prefix = " ө"
            }
            asText = append(asText, fmt.Sprintf("%s%s", prefix, key.Name()))
        }
    }
    g.ShowScrollableText(asText, color.White, false)
}

func (g *GridEngine) sortInventory(partyInventory [][]game.Item) {
    sort.SliceStable(partyInventory, func(i, j int) bool {
        item := partyInventory[i][0]
        other := partyInventory[j][0]
        if other.Name() == item.Name() {
            if len(partyInventory[i]) == len(partyInventory[j]) {
                if wornItem, ok := item.(game.Wearable); ok && wornItem.IsEquipped() {
                    if otherWornItem, ok := other.(game.Wearable); ok && otherWornItem.IsEquipped() {
                        wearer := wornItem.GetWearer()
                        otherWearer := otherWornItem.GetWearer()
                        wearerIndex := g.playerParty.GetMemberIndex(wearer)
                        otherWearerIndex := g.playerParty.GetMemberIndex(otherWearer)
                        return wearerIndex < otherWearerIndex
                    } else {
                        return true
                    }
                }
            }
            return len(partyInventory[i]) < len(partyInventory[j])
        }
        return item.Name() < other.Name()
    })
}

func (g *GridEngine) ShowText(text []string) {
    g.ShowScrollableText(text, color.White, true)
}

func (g *GridEngine) ShowFixedFormatText(text []string) *ui.ScrollableTextWindow {
    return g.showScrollText(text, color.White, false)
}

func (g *GridEngine) ShowScrollableText(text []string, textcolor color.Color, autolayoutText bool) {
    g.showScrollText(text, textcolor, autolayoutText)
}

func (g *GridEngine) showScrollText(text []string, textcolor color.Color, autolayoutText bool) *ui.ScrollableTextWindow {
    if len(text) == 0 {
        return nil
    }
    g.CloseAllModals()

    var textWindow *ui.ScrollableTextWindow
    if autolayoutText {
        textWindow = ui.NewAutoTextWindow(g.gridRenderer, text)
    } else {
        textWindow = ui.NewFixedTextWindow(g.gridRenderer, text)
    }
    textWindow.SetTextColor(textcolor)
    g.PushModal(textWindow)
    g.lastShownText = text
    return textWindow
}

func (g *GridEngine) PickPocketItem(item game.Item, owner *game.Actor) {
    owner.RemoveItem(item)
    g.takeItem(item)
}

func (g *GridEngine) PickUpItem(item game.Item) {
    g.currentMap.RemoveItem(item)
    g.takeItem(item)
}
func (g *GridEngine) moveItemToParty(item game.Item, container game.ItemContainer) {
    container.RemoveItem(item)
    g.takeItem(item)
}

func (g *GridEngine) takeItem(item game.Item) {
    if pseudoItem, ok := item.(*game.PseudoItem); ok {
        pseudoItem.Take(g)
    } else {
        g.AddItem(item)
    }
    pickupEvent := item.GetPickupEvent()
    if pickupEvent != "" {
        g.TriggerEvent(pickupEvent)
    }
}
func (g *GridEngine) DropItem(item game.Item) {
    if equippable, ok := item.(game.Equippable); ok && equippable.IsEquipped() {
        equippable.Unequip()
    }
    g.playerParty.RemoveItem(item)
    destPos := g.avatar.Pos()
    if g.TryPlaceItem(item, destPos) {
        g.Print(fmt.Sprintf("Dropped \"%s\"", item.Name()))
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
    if g.IsInConversation() {
        g.Print("You can't do that right now.")
        return
    }
    if !g.playerParty.IsMember(actor) {
        g.Print(fmt.Sprintf("\"%s\" is not in your party", actor.Name()))
        return
    }
    g.playerParty.SwitchControlTo(actor)
    g.onViewedActorMoved(actor.Pos())

    if g.IsModalOpen() {
        g.topModal().OnAvatarSwitched()
    }
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
        if !member.IsNearTo(g.avatar) {
            g.Print(fmt.Sprintf("\"%s\" is not next to you.", member.Name()))
            return
        }
    }
    g.playerParty.ReturnControlToLeader()
}

func (g *GridEngine) ForceJoinParty() {
    g.playerParty.ReturnControlToLeader()
    g.onViewedActorMoved(g.GetAvatar().Pos())
}
func (g *GridEngine) openVendorMenu(npc *game.Actor) { // TODO: replace with usage of the dialogue system
    itemsToSell := npc.GetItemsToSell()
    if len(itemsToSell) == 0 {
        g.conversationModal.SetText(oneLine("I have nothing to sell."))
        g.conversationModal.SetVendorOptions(nil)
        return
    }

    labelWidth, colWidth := getLineLengthInfoItems(itemsToSell)
    var menuItems []util.MenuItem
    for _, i := range itemsToSell {
        offer := i
        itemLine := util.TableLine(labelWidth, colWidth, offer.Item.Name(), strconv.Itoa(offer.Price))
        //itemLine := fmt.Sprintf("%s (%d)", offer.Item.Name(), offer.Price)
        menuItems = append(menuItems, util.MenuItem{
            Text:      itemLine,
            CharIcon:  offer.Item.InventoryIcon(),
            TextColor: offer.Item.TintColor(),
            Action: func() {
                g.TryBuyItem(npc, offer)
                g.openVendorMenu(npc)
            },
        })
    }
    g.conversationModal.SetVendorOptions(menuItems)
    g.conversationModal.OnMouseMoved(g.lastMousePosX, g.lastMousePosY)
    //g.openMenuForVendor(menuItems)
}

func (g *GridEngine) openTrainerMenu(npc *game.Actor, maxLevel int) {
    // we want to offer all the party members who currently can level up
    // to the given max level

    var eligibleMembers []*game.Actor

    for _, m := range g.playerParty.GetMembers() {
        if m.GetLevel() >= maxLevel {
            continue
        }
        canLevel, _ := g.rules.CanLevelUp(m.GetLevel(), m.GetXP())
        if !canLevel {
            continue
        }
        eligibleMembers = append(eligibleMembers, m)
    }

    if len(eligibleMembers) == 0 {
        g.conversationModal.SetText(oneLine("No member of your party can be trained here."))
        g.conversationModal.SetVendorOptions(nil)
        return
    }

    labelWidth, colWidth := getLineLengthInfoActors(eligibleMembers)
    var menuItems []util.MenuItem
    for _, i := range eligibleMembers {
        member := i
        levelUpCost := g.rules.GetTrainerCost(member.GetLevel())
        itemLine := util.TableLine(labelWidth, colWidth, member.Name(), fmt.Sprintf("(%d)", member.GetLevel()), fmt.Sprintf(" %dg", levelUpCost))
        //itemLine := fmt.Sprintf("%s (%d)", member.Item.Name(), member.Price)
        menuItems = append(menuItems, util.MenuItem{
            Text: itemLine,
            Action: func() {
                if g.playerParty.HasGold(levelUpCost) {
                    g.playerParty.RemoveGold(levelUpCost)
                    npc.AddGold(levelUpCost)
                    g.rules.LevelUp(member)
                    g.openTrainerMenu(npc, maxLevel)
                } else {
                    g.conversationModal.SetText(oneLine("You don't have enough gold."))
                    g.conversationModal.SetVendorOptions(nil)
                }
            },
        })
    }
    g.conversationModal.SetVendorOptions(menuItems)
    g.conversationModal.OnMouseMoved(g.lastMousePosX, g.lastMousePosY)
}

func getLineLengthInfoItems(sell []game.SalesOffer) (int, int) {
    var longestItemNameLength, longestPriceLength int
    for _, i := range sell {
        if len(i.Item.Name()) > longestItemNameLength {
            longestItemNameLength = len(i.Item.Name())
        }
        if len(strconv.Itoa(i.Price)) > longestPriceLength {
            longestPriceLength = len(strconv.Itoa(i.Price))
        }
    }
    return longestItemNameLength, longestPriceLength
}
func getLineLengthInfoActors(actors []*game.Actor) (int, int) {
    var longestItemNameLength, longestLevelLength int
    for _, i := range actors {
        if len(i.Name()) > longestItemNameLength {
            longestItemNameLength = len(i.Name())
        }
        levelColWidth := len(strconv.Itoa(i.GetLevel())) + 2
        if levelColWidth > longestLevelLength {
            longestLevelLength = levelColWidth
        }
    }
    return longestItemNameLength, longestLevelLength
}

func (g *GridEngine) TryBuyItem(npc *game.Actor, offer game.SalesOffer) {
    if g.playerParty.GetGold() < offer.Price {
        g.conversationModal.SetText(oneLine("You don't have enough gold."))
        return
    }
    if !npc.RemoveItem(offer.Item) {
        g.conversationModal.SetText(oneLine("I don't have that item anymore."))
        return
    }
    g.playerParty.RemoveGold(offer.Price)

    npc.AddGold(offer.Price)

    g.AddItem(offer.Item)

    g.conversationModal.SetText(oneLine("Thank you for your business."))
}

func (g *GridEngine) TryRestParty() {
    if !g.playerParty.NeedsRest() {
        g.Print("Your party is not tired.")
        return
    }
    if g.playerParty.TryRest() {
        g.AdvanceWorldTimeWithMessage(0, 8, 0)
        time := g.GetWorldTime()
        pages := g.gridRenderer.AutolayoutArrayToIconPages(5, []string{
            "You have eaten some food",
            "and rested the night.",
            "Your party has been healed.",
            "It is now:",
            time.GetTimeAndDate(),
        })
        window := g.openIconWindow(g.GetAvatar().Icon(0), pages, func() {})
        window.SetAutoCloseOnConfirm()
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
    neighbors = append(neighbors, source)
    for _, n := range neighbors {
        neighbor := n
        g.HitAnimation(neighbor, renderer.AtlasEntities, 195, color.White, func() {
            g.revealHiddenObjectsAt(neighbor)
        })
    }
}

func (g *GridEngine) revealHiddenObjectsAt(location geometry.Point) {
    foundSomething := false
    if g.currentMap.IsSecretDoorAt(location) {
        wallTile := g.currentMap.GetCell(location).TileType
        g.currentMap.SetTile(location, wallTile.WithIsWalkable(true))
        g.currentMap.AddObject(game.NewDoor(), location)
        foundSomething = true
        g.flags.IncrementFlag("found_hidden_things")
    } else {
        if g.currentMap.IsObjectAt(location) {
            objectAt := g.currentMap.ObjectAt(location)
            if objectAt != nil && objectAt.IsHidden() {
                message := objectAt.Discover()
                foundSomething = true
                g.flags.IncrementFlag("found_hidden_things")
                if len(message) > 0 {
                    g.ShowText(message)
                }
            }
        } else if g.currentMap.IsItemAt(location) {
            itemAt := g.currentMap.ItemAt(location)
            if itemAt != nil && itemAt.IsHidden() {
                message := itemAt.Discover()
                foundSomething = true
                g.flags.IncrementFlag("found_hidden_things")
                if len(message) > 0 {
                    g.ShowText(message)
                }
            }
        } else if g.currentMap.IsActorAt(location) {
            actorAt := g.currentMap.GetActor(location)
            if actorAt != nil && actorAt.IsHidden() {
                message := actorAt.Discover()
                foundSomething = true
                g.flags.IncrementFlag("found_hidden_things")
                if len(message) > 0 {
                    g.ShowText(message)
                }
            }
        }
    }
    if foundSomething {
        g.Print("Fascinating..")
    }
}

func (g *GridEngine) PartyHasKey(key string) bool {
    return g.playerParty.HasKey(key)
}

func (g *GridEngine) AddFood(amount int) {
    g.Print(fmt.Sprintf("%d food added", amount))
    g.playerParty.AddFood(amount)
}

func (g *GridEngine) AddLockpicks(amount int) {
    g.Print(fmt.Sprintf("%d lockpicks added", amount))
    g.playerParty.AddLockpicks(amount)
}
func (g *GridEngine) AddItem(item game.Item) {
    if pseudoItem, ok := item.(*game.PseudoItem); ok {
        pseudoItem.Take(g)
        return
    }
    g.Print(fmt.Sprintf("Received \"%s\"", item.Name()))
    g.playerParty.AddItem(item)
}
func (g *GridEngine) AddGold(amount int) {
    g.Print(fmt.Sprintf("%d gold added", amount))
    g.playerParty.AddGold(amount)
}

func (g *GridEngine) GetPartySize() int {
    return len(g.playerParty.GetMembers())
}

func (g *GridEngine) CreateLootForContainer(level int, lootType []game.Loot) []game.Item {
    var lootFound []game.Item
    for _, loot := range lootType {
        var lootItems []game.Item
        randFloat := rand.Float64()
        switch loot {
        case game.LootLockpicks:
            lockpickAmount := max(level, int(float64(level)*3*randFloat))
            lootItems = []game.Item{game.NewPseudoItemFromTypeAndAmount(game.PseudoItemTypeLockpick, lockpickAmount)}
        case game.LootGold:
            goldAmount := max(10*level, int(float64(level)*100.0*randFloat))
            lootItems = []game.Item{game.NewPseudoItemFromTypeAndAmount(game.PseudoItemTypeGold, goldAmount)}
        case game.LootFood:
            foodAmount := max(2*level, int(float64(level)*6*randFloat))
            lootItems = []game.Item{game.NewPseudoItemFromTypeAndAmount(game.PseudoItemTypeFood, foodAmount)}
        case game.LootPotions:
            potionAmount := max(1, int(float64(level)*2*randFloat))
            lootItems = g.createPotions(potionAmount)
        case game.LootArmor:
            armorAmount := max(1, int(float64(level)*randFloat))
            lootItems = g.createArmor(level, armorAmount)
        case game.LootWeapon:
            weaponAmount := max(1, int(float64(level)*randFloat))
            lootItems = g.createWeapons(level, weaponAmount)

        }
        lootFound = append(lootFound, lootItems...)
    }
    return lootFound
}

func (g *GridEngine) ShowContainer(container game.ItemContainer) {
    var menuItems []util.MenuItem
    containerItems := container.GetItems()
    if len(containerItems) == 0 {
        g.ShowText([]string{"The container is empty."})
        return
    }
    if len(containerItems) > 1 {
        menuItems = append(menuItems, util.MenuItem{
            CharIcon: 162,
            Text:     "Take all",
            Action: func() {
                for i := len(containerItems) - 1; i >= 0; i-- {
                    g.moveItemToParty(containerItems[i], container)
                }
            },
        })
    }
    for _, i := range containerItems {
        item := i
        menuItems = append(menuItems, util.MenuItem{
            CharIcon:  item.InventoryIcon(),
            TextColor: item.TintColor(),
            Text:      item.Name(),
            Action: func() {
                g.moveItemToParty(item, container)
            },
        })
    }
    g.OpenMenu(menuItems)
}

func oneLine(text string) [][]string {
    return [][]string{{text}}
}
