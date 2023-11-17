package main

import (
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/ui"
    "Legacy/util"
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
    "os"
)

func (g *GridEngine) openContextMenu(menu []util.MenuItem) {
    if len(menu) == 0 {
        return
    }
    if len(menu) == 1 {
        menu[0].Action()
        return
    }
    g.OpenMenu(menu)
}

func (g *GridEngine) OpenMenu(items []util.MenuItem) {
    g.openMenuWithTitle("", items)
}

func (g *GridEngine) openMenuWithTitle(title string, items []util.MenuItem) *ui.GridMenu {
    if len(items) == 0 {
        return nil
    }
    gridMenu := ui.NewGridMenu(g.gridRenderer, items)
    gridMenu.OnMouseMoved(g.lastMousePosX, g.lastMousePosY)
    gridMenu.SetAutoClose()
    if title != "" {
        gridMenu.SetTitle(title)
    }
    if g.lastInteractionWasMouse {
        gridMenu.PositionNearMouse(g.lastMousePosX, g.lastMousePosY)
    }
    g.PushModal(gridMenu)
    return gridMenu
}
func (g *GridEngine) NewTextInputAtY(yPos int, prompt string, onClose func(endedWith ui.EndAction, text string)) *ui.TextInput {
    cursorIcon := int32(28)
    cursorFrameCount := 4
    input := ui.NewTextInput(g.gridRenderer, geometry.Point{}, 15, cursorIcon, cursorFrameCount, onClose)
    input.SetDrawBorder(true)
    input.SetPrompt(prompt)
    input.CenterHorizontallyAtY(yPos)
    return input
}
func (g *GridEngine) openSpellMenu() {
    var menuItems []util.MenuItem
    spells := g.playerParty.GetSpells()
    if len(spells) == 0 {
        g.ShowText([]string{"You don't have any spells."})
        return
    }
    for _, s := range spells {
        spell := s
        if !g.GetAvatar().HasMana(spell.ManaCost()) {
            continue
        }
        label := fmt.Sprintf("%s (%d)", spell.Name(), spell.ManaCost())
        menuItems = append(menuItems, util.MenuItem{
            Text: label,
            Action: func() {
                if spell.IsTargeted() {
                    g.combatManager.PlayerStartsOffensiveSpell(g.GetAvatar(), spell)
                } else {
                    spell.Cast(g, g.GetAvatar())
                }
            },
        })
    }
    if len(menuItems) == 0 {
        g.ShowText([]string{"You don't have enough mana to cast any spells."})
        return
    }
    g.OpenMenu(menuItems)
}

func (g *GridEngine) openCombatSpellMenu(member *game.Actor) {
    var menuItems []util.MenuItem
    spells := member.GetEquippedSpells()
    if len(spells) == 0 {
        g.ShowText([]string{"You don't have any spells."})
        return
    }
    for _, s := range spells {
        spell := s
        if !member.HasMana(spell.ManaCost()) {
            continue
        }
        label := fmt.Sprintf("%s (%d)", spell.Name(), spell.ManaCost())
        menuItems = append(menuItems, util.MenuItem{
            Text: label,
            Action: func() {
                if spell.IsTargeted() {
                    g.CloseAllModals()
                    g.combatManager.SelectSpellTarget(member, spell)
                } else {
                    spell.Cast(g, member)
                }
            },
        })
    }
    if len(menuItems) == 0 {
        g.ShowText([]string{"You don't have enough mana to cast any spells."})
        return
    }
    g.OpenMenu(menuItems)
}

func (g *GridEngine) ShowDrinkPotionMenu(potion *game.Potion) {
    var menuItems []util.MenuItem
    for _, m := range g.playerParty.GetMembers() {
        member := m
        menuItems = append(menuItems, util.MenuItem{
            Text: fmt.Sprintf("%s", member.Name()),
            Action: func() {
                if !potion.IsEmpty() {
                    g.DrinkPotion(potion, member)
                }
            },
        })
    }
    g.OpenMenu(menuItems)
}

func (g *GridEngine) OpenPickpocketMenu(victim *game.Actor) {
    var itemList []util.MenuItem
    items := victim.GetItemsToSteal()
    if len(items) == 0 {
        g.ShowText([]string{"Nothing to steal."})
        return
    }
    for _, i := range items {
        item := i
        itemList = append(itemList, util.MenuItem{
            Text: item.Name(),
            Action: func() {
                if item.GetHolder() == victim {
                    g.TryPickpocketItem(item, victim)
                }
            },
        })
    }
    g.OpenMenu(itemList)
}

func (g *GridEngine) OpenPlantMenu(victim *game.Actor) {
    var itemList []util.MenuItem
    items := g.playerParty.GetFilteredStackedInventory(func(item game.Item) bool {
        return true
        // until we know better
        /*
           _, isArmor := item.(*game.Armor)
           _, isWeapon := item.(*game.Weapon)
           return !isArmor && !isWeapon
        */
    })
    if len(items) == 0 {
        g.ShowText([]string{"You don't have any items to plant."})
        return
    }
    for _, i := range items {
        itemStack := i
        firstItem := itemStack[0]
        itemList = append(itemList, util.MenuItem{
            Text: firstItem.Name(),
            Action: func() {
                g.TryPlantItem(firstItem, victim)
            },
        })

    }
    g.OpenMenu(itemList)
}

func (g *GridEngine) ShowEquipMenu(a game.Equippable) {
    if g.IsPlayerControlled(a.GetHolder()) {
        if len(g.GetPartyMembers()) == 1 {
            g.EquipItem(g.GetAvatar(), a)
            return
        }
        var equipMenuItems []util.MenuItem
        for _, m := range g.GetPartyMembers() {
            partyMember := m
            if partyMember.CanEquip(a) {
                equipMenuItems = append(equipMenuItems, util.MenuItem{
                    Text: partyMember.Name(),
                    Action: func() {
                        g.EquipItem(partyMember, a)
                    },
                })
            }
        }
        g.OpenMenu(equipMenuItems)
    }
}

func (g *GridEngine) openDebugMenu() {
    g.OpenMenu([]util.MenuItem{
        {
            Text:   "Teleport",
            Action: g.openTransportMenu,
        },
        {
            Text: "impulse 9",
            Action: func() {
                g.playerParty.AddFood(100)
                g.playerParty.AddGold(10000000)
                g.playerParty.AddLockpicks(100)
                g.avatar.SetHealth(1000)
                g.avatar.SetMana(1000)
                g.playerParty.AddXPForEveryone(1000000)
                g.playerParty.AddItem(game.NewTool(game.ToolTypePickaxe, "a pickaxe"))
                g.playerParty.AddItem(game.NewTool(game.ToolTypeShovel, "a shovel"))
                g.giveAllArmorsAndWeapons()
                g.Print("DEBUG(impulse 9)")
            },
        },
        {
            Text: "Toggle NoClip",
            Action: func() {
                noClipActive := g.currentMap.ToggleNoClip()
                g.Print(fmt.Sprintf("DEBUG(No Clip): %t", noClipActive))
            },
        },

        {
            Text: "Damage to Avatar",
            Action: func() {
                g.avatar.Damage(10)
            },
        },
        {
            Text: "Get All Skills",
            Action: func() {
                g.avatar.AddAllSkills()
            },
        },
        {
            Text: "Save Game",
            Action: func() {
                g.saveGameToDirectory("./saves/test01")
            },
        },
        {
            Text: "Load Game",
            Action: func() {
                g.loadGameFromDirectory("./saves/test01")
            },
        },
        {
            Text: "Show all Flags",
            Action: func() {
                g.ShowScrollableText(g.flags.GetDebugInfo(), color.White, false)
            },
        },
        {
            Text: "Show XP Table",
            Action: func() {
                g.ShowScrollableText(g.rules.GetXPTable(2, 30), color.White, false)
            },
        },
    })
}

func (g *GridEngine) ChangeAppearance() {
    iconWindow := ui.NewIconWindow(g.gridRenderer)
    iconWindow.SetOnClose(func() {
        g.GetAvatar().SetIcon(iconWindow.Icon())
    })
    iconWindow.SetAllowedIcons(g.allowedPartyIcons)
    iconWindow.SetCurrentIcon(g.GetAvatar().Icon(0))
    iconWindow.SetFixedText([]string{"What do you see?"})
    iconWindow.SetYOffset(10)
    g.PushModal(iconWindow)
}

func (g *GridEngine) saveGameToDirectory(directory string) {
    var _ = os.MkdirAll(directory, 0755) // remove party from the map, so we don't save it twice
    g.RemovePartyFromMap(g.currentMap)

    savePartyState(g.playerParty, directory) // current map is saved in party state
    saveExtendedState(g.flags, g.playerKnowledge, directory)
    saveAllMaps(g.getAllLoadedMaps(), directory)

    g.PlacePartyBackOnCurrentMap()
}

func (g *GridEngine) loadGameFromDirectory(directory string) {
    party, currentMapName := loadPartyState(directory)
    party.SetRules(g.rules)

    g.avatar = party.GetMember(0)
    g.playerParty = party

    g.flags, g.playerKnowledge = loadExtendedState(directory) //TODO
    g.mapsInMemory = loadAllMaps(directory)

    // set the current map
    // place the party on the map
    g.currentMap = g.mapsInMemory[currentMapName]

    g.initMapWindow(g.currentMap.MapWidth, g.currentMap.MapHeight)

    g.PlacePartyBackOnCurrentMap()
}

type Modal interface {
    Draw(screen *ebiten.Image)
    ShouldClose() bool
    OnAvatarSwitched()

    OnCommand(command ui.CommandType) bool
    OnMouseMoved(x int, y int) (bool, ui.Tooltip)
    OnMouseClicked(x int, y int) bool
    OnMouseWheel(x int, y int, dy float64) bool
}
