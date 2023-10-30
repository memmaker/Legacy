package main

import (
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/renderer"
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
)

func (g *GridEngine) openContextMenu() {
    if len(g.contextActions) == 0 {
        return
    }
    if len(g.contextActions) == 1 {
        g.contextActions[0].Action()
        return
    }
    g.openMenu(g.contextActions)
}

func (g *GridEngine) openMenu(items []renderer.MenuItem) {
    g.openMenuWithTitle("", items)
}
func (g *GridEngine) openMenuWithTitle(title string, items []renderer.MenuItem) {
    gridMenu := renderer.NewGridMenu(g.gridRenderer, items)
    gridMenu.SetAutoClose()
    if title != "" {
        gridMenu.SetTitle(title)
    }
    g.inputElement = gridMenu
    g.inputElement.OnMouseMoved(g.lastMousePosX, g.lastMousePosY)
}

func (g *GridEngine) openSpellMenu() {
    var menuItems []renderer.MenuItem
    spells := g.playerParty.GetSpells()
    if len(spells) == 0 {
        g.ShowText([]string{"You don't have any spells."})
        return
    }
    for _, s := range spells {
        spell := s
        label := fmt.Sprintf("%s (%d)", spell.Name(), spell.ManaCost())
        menuItems = append(menuItems, renderer.MenuItem{
            Text: label,
            Action: func() {
                if spell.IsTargeted() {
                    g.chooseTarget(func(target geometry.Point) {
                        spell.CastOnTarget(g, g.GetAvatar(), target)
                    })
                } else {
                    spell.Cast(g, g.GetAvatar())
                }
            },
        })
    }
    g.openMenu(menuItems)
}

func (g *GridEngine) ShowDrinkPotionMenu(potion *game.Potion) {
    var menuItems []renderer.MenuItem
    for _, m := range g.playerParty.GetMembers() {
        member := m
        menuItems = append(menuItems, renderer.MenuItem{
            Text: fmt.Sprintf("%s", member.Name()),
            Action: func() {
                if !potion.IsEmpty() {
                    g.DrinkPotion(potion, member)
                }

            },
        })
    }
    g.openMenu(menuItems)
}

func (g *GridEngine) OpenPickpocketMenu(victim *game.Actor) {
    var itemList []renderer.MenuItem
    items := victim.GetItemsToSteal()
    for _, i := range items {
        item := i
        itemList = append(itemList, renderer.MenuItem{
            Text: item.Name(),
            Action: func() {
                if item.GetHolder() == victim {
                    g.TryPickpocketItem(item, victim)
                }
            },
        })
    }
    g.openMenu(itemList)
}

func (g *GridEngine) ShowEquipMenu(a *game.Armor) {
    if g.IsPlayerControlled(a.GetHolder()) {
        if len(g.GetPartyMembers()) == 1 {
            g.EquipArmor(g.GetAvatar(), a)
            return
        }
        var equipMenuItems []renderer.MenuItem
        for _, partyMember := range g.GetPartyMembers() {
            if partyMember.CanEquip(a) {
                equipMenuItems = append(equipMenuItems, renderer.MenuItem{
                    Text: partyMember.Name(),
                    Action: func() {
                        g.EquipArmor(partyMember, a)
                    },
                })
            }
        }
        g.openMenu(equipMenuItems)
    }
}

func (g *GridEngine) openDebugMenu() {
    g.openMenu([]renderer.MenuItem{
        {
            Text: "Toggle NoClip",
            Action: func() {
                noClipActive := g.currentMap.ToggleNoClip()
                g.Print(fmt.Sprintf("DEBUG(No Clip): %t", noClipActive))
            },
        },
        {
            Text: "impulse 9",
            Action: func() {
                g.playerParty.AddFood(100)
                g.playerParty.AddGold(100)
                g.playerParty.AddLockpicks(100)
                g.Print("DEBUG(impulse 9): Added 100 food, gold and lockpicks")
            },
        },
        {
            Text: "Set some flags",
            Action: func() {
                g.flags.SetFlag("can_talk_to_ghosts", 1)
                g.Print("DEBUG(Flags): Set flag \"can_talk_to_ghosts\"")
            },
        },
        {
            Text: "Set Avatar Name to 'Nova'",
            Action: func() {
                g.avatar.SetName("Nova")
                g.Print("DEBUG(Avatar): Set name to 'Nova'")
            },
        },
        {
            Text: "Show all Flags",
            Action: func() {
                g.ShowColoredText(g.flags.GetDebugInfo(), color.White, false)
            },
        },
    })
}

type Modal interface {
    Draw(screen *ebiten.Image)
    ActionUp()
    ActionDown()
    ActionConfirm()
    ShouldClose() bool
}

type UIWidget interface {
    Draw(screen *ebiten.Image)
    ActionUp()
    ActionDown()
    ActionConfirm()
    OnMouseClicked(x int, y int)
    OnMouseMoved(x int, y int)
    ShouldClose() bool
}
