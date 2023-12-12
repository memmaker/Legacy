package main

import (
    "Legacy/ega"
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/gocoro"
    "Legacy/ui"
    "Legacy/util"
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
    "os"
    "time"
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
func (g *GridEngine) openActiveSkillsMenu(member *game.Actor, actions []game.Action) {
    var menuItems []util.MenuItem
    if len(actions) == 0 {
        g.ShowText([]string{"Nothing available."})
        return
    }
    for _, s := range actions {
        activeSkill := s
        itemColor := ega.BrightWhite
        if !activeSkill.CanPayCost(g, member) {
            itemColor = ega.White
        }
        label := activeSkill.LabelWithCost()
        menuItems = append(menuItems, util.MenuItem{
            Text:        label,
            TextColor:   itemColor,
            TooltipText: activeSkill.GetDescription(),
            Action: func() {
                if !activeSkill.CanPayCost(g, member) {
                    g.Print("You can't pay the cost for this ability.")
                    return
                }
                if activeSkill.IsTargeted() {
                    g.CloseAllModals()
                    g.combatManager.PlayerUsesActiveSkill(member, activeSkill)
                } else {
                    activeSkill.Execute(g, member)
                }
            },
        })
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
            Text: "Enter the Dungeon",
            Action: func() {
                delete(g.mapsInMemory, "!gen_dungeon_level_1")
                g.TransitionToNamedLocation("!gen_dungeon_level_1", "ladder_up")
            },
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
                g.giveAllSpells()
                g.Flags().SetFlag("needs_form_32", 1)
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
                g.avatar.Damage(g, 10)
            },
        },
        {
            Text: "Get All Skills",
            Action: func() {
                g.avatar.AddAllSkills()
            },
        },
        {
            Text: "Get Buffs",
            Action: func() {
                g.AddStatusEffect(g.GetAvatar(), game.StatusBlessed(), 10)
                g.AddStatusEffect(g.GetAvatar(), game.StatusHolyBonus(), 10)
            },
        },
        {
            Text: "Get Weak & Undead",
            Action: func() {
                g.AddStatusEffect(g.GetAvatar(), game.StatusWeak(), 1)
                g.AddStatusEffect(g.GetAvatar(), game.StatusUndead(), 1)
            },
        },
        {
            Text: "Edit Skills",
            Action: func() {
                g.openSkillEditor()
            },
        },
        {
            Text: "Edit Attributes",
            Action: func() {
                g.openAttributeEditor()
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
                g.ShowScrollableText(g.rules.GetXPTable(2, 21), color.White, false)
            },
        },
        {
            Text: "Show Skill Check Tables",
            Action: func() {
                g.ShowScrollableText(g.rules.GetSkillCheckTable(), color.White, false)
            },
        },
        {
            Text: "Test NPC movement",
            Action: func() {
                g.startActorPath("tauci_front_guard", "guard_patrol")
            },
        },
    })
}
func (g *GridEngine) startActorPath(internalName, pathName string) {
    actor := g.GetActorByInternalName(internalName)
    waypoints := g.currentMap.GetNamedPath(pathName)
    if waypoints == nil || len(waypoints) == 0 {
        g.Print(fmt.Sprintf("No path %s", pathName))
        return
    }
    if actor == nil {
        g.Print(fmt.Sprintf("No actor %s", internalName))
        return
    }
    err := g.RunAnimationScript(g.getNPCMovementRoutine(actor, waypoints))
    if err != nil {
        println(err.Error())
    }
}
func (g *GridEngine) getNPCMovementRoutine(actor *game.Actor, waypoints []geometry.Point) func(exe *gocoro.Execution) {
    return func(exe *gocoro.Execution) {
        for _, currentWaypoint := range waypoints {
            for actor.IsSleeping() {
                _ = exe.YieldTime(2 * time.Second)
            }
            var currentPath []geometry.Point
            moveBlockedCount := 0
            for actor.Pos() != currentWaypoint {
                //
                if currentPath == nil || len(currentPath) == 0 {
                    currentPath = g.currentMap.GetJPSPath(actor.Pos(), currentWaypoint, g.currentMap.IsCurrentlyPassable)
                    if len(currentPath) > 0 {
                        currentPath = currentPath[1:] // remove the first element, which is the current position
                    }
                }
                for len(currentPath) > 0 {
                    moveBlockedCount = 0
                    nextPos := currentPath[0]
                    currentPath = currentPath[1:]
                    for actor.Pos() != nextPos {
                        g.TryMoveNPCOnPath(actor, nextPos)
                        _ = exe.YieldTime(900 * time.Millisecond)
                        moveBlockedCount++
                        if moveBlockedCount > 5 {
                            currentPath = nil
                            moveBlockedCount = 0
                            break
                        }
                    }
                }
            }
        }
    }
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
    g.setMap(g.mapsInMemory[currentMapName])

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
