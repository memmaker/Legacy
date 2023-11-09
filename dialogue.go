package main

import (
    "Legacy/ega"
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/recfile"
    "Legacy/renderer"
    "fmt"
)

func (g *GridEngine) openSpeechWindow(npc *game.Actor, textPages [][]string, onLastPage func()) {
    if len(textPages) == 0 {
        return
    }
    multipageWindow := renderer.NewMultiPageWindow(g.gridRenderer, onLastPage)
    multipageWindow.InitWithFixedText(textPages)
    multipageWindow.SetTitle(npc.Name())
    multipageWindow.AddTextActionButton(134, func(text []string) {
        g.addToJournal(npc, text)
    })
    multipageWindow.SetIcon(npc.Icon(0))
    multipageWindow.PositionAtY(3)
    multipageWindow.SetAutoCloseOnConfirm()

    g.modalElement = multipageWindow
}

func (g *GridEngine) openIconWindow(icon int32, textPages [][]string, onLastPage func()) *renderer.MultiPageWindow {
    if len(textPages) == 0 {
        return nil
    }
    multipageWindow := renderer.NewMultiPageWindow(g.gridRenderer, onLastPage)
    multipageWindow.InitWithFixedText(textPages)
    multipageWindow.SetIcon(icon)
    multipageWindow.PositionAtY(3)
    g.modalElement = multipageWindow
    return multipageWindow
}

func (g *GridEngine) openConversationMenu(topLeft geometry.Point, items []renderer.MenuItem) {
    g.switchInputElement(renderer.NewGridDialogueMenu(g.gridRenderer, topLeft, items))
    g.inputElement.OnMouseMoved(g.lastMousePosX, g.lastMousePosY)
}
func (g *GridEngine) StartConversation(npc *game.Actor, loadedDialogue *game.Dialogue) {
    // NOTE: Conversations can have a line length of 27 chars
    if !npc.IsAlive() {
        g.Print(fmt.Sprintf("%s is dead.", npc.Name()))
        return
    }
    if !npc.HasDialogue() {
        g.Print(fmt.Sprintf("\"%s\" has nothing to say.", npc.Name()))
        return
    }

    var openingLines [][]string
    var items []renderer.MenuItem
    var firstNode game.ConversationNode
    if !g.playerKnowledge.HasTalkedTo(npc.GetInternalName()) && loadedDialogue.HasFirstTimeText() {
        g.playerKnowledge.AddTalkedTo(npc.GetInternalName())
        firstNode = loadedDialogue.GetFirstTimeText()
    } else if loadedDialogue.HasOpening() {
        firstNode = loadedDialogue.GetOpening()
    } else {
        firstNode = game.ConversationNode{
            Text: oneLine("Hi there!"),
        }
    }

    g.playerKnowledge.AddKnowledge(firstNode.AddsKeywords)
    g.flags.SetFlags(firstNode.FlagsSet)
    openingLines = firstNode.Text
    if len(firstNode.ForcedChoice) > 0 {
        items = g.toForcedMenuItems(npc, loadedDialogue, firstNode.ForcedChoice)
    }
    options := loadedDialogue.GetOptions(g.GetAvatar(), g.playerKnowledge, g.flags)

    if len(items) == 0 {
        items = g.toMenuItems(npc, loadedDialogue, options)
    }

    startedConversationFlag := fmt.Sprintf("talked_to_%s", npc.GetInternalName())
    g.flags.IncrementFlag(startedConversationFlag)

    g.openSpeechWindow(npc, openingLines, func() {
        if len(items) > 0 {
            g.openConversationMenu(geometry.Point{X: 3, Y: 13}, items)
        }
    })
}

func (g *GridEngine) ShowMultipleChoiceDialogue(icon int32, text [][]string, choices []renderer.MenuItem) {
    g.openIconWindow(icon, text, func() {
        if len(choices) > 0 {
            g.openMenuForDialogue(choices)
        }
    })
}

func (g *GridEngine) openMenuForDialogue(items []renderer.MenuItem) {
    // geometry.Point{X: 3, Y: 13},
    gridMenu := renderer.NewGridMenuAtY(g.gridRenderer, items, 13)
    //gridMenu.SetAutoClose()
    g.switchInputElement(gridMenu)
    g.inputElement.OnMouseMoved(g.lastMousePosX, g.lastMousePosY)
}
func (g *GridEngine) openMenuForVendor(items []renderer.MenuItem) *renderer.GridMenu {
    // geometry.Point{X: 3, Y: 13},
    gridMenu := renderer.NewGridMenuAtY(g.gridRenderer, items, 10)
    gridMenu.SetAutoClose()
    g.switchInputElement(gridMenu)
    g.inputElement.OnMouseMoved(g.lastMousePosX, g.lastMousePosY)
    return gridMenu
}

func (g *GridEngine) toMenuItems(npc *game.Actor, dialogue *game.Dialogue, options []string) []renderer.MenuItem {
    var items []renderer.MenuItem
    for _, o := range options {
        option := o
        textColor := ega.BrightYellow
        if dialogue.HasBeenUsed(option) {
            textColor = ega.BrightWhite
        }
        items = append(items, renderer.MenuItem{
            Text: option,
            Action: func() {
                g.handleDialogueChoice(dialogue, option, npc)
            },
            TextColor: textColor,
        })
    }
    return items
}

func (g *GridEngine) toForcedMenuItems(npc *game.Actor, dialogue *game.Dialogue, choices []game.DialogueChoice) []renderer.MenuItem {
    var items []renderer.MenuItem
    for _, c := range choices {
        if len(c.NeededFlags) > 0 && !g.Flags().AllSet(c.NeededFlags) {
            continue
        }
        choice := c
        items = append(items, renderer.MenuItem{
            Text: choice.Text,
            Action: func() {
                transitionTarget := c.TransitionOnSuccess
                if len(c.NeededSkills) > 0 && !g.GetAvatar().GetSkills().HasSkills(c.NeededSkills) {
                    transitionTarget = c.TransitionOnFail
                }
                g.handleDialogueChoice(dialogue, transitionTarget, npc)
            },
        })
    }
    return items
}

func (g *GridEngine) handleDialogueChoice(dialogue *game.Dialogue, followUp string, npc *game.Actor) {
    response := dialogue.GetResponseAndAddKnowledge(g.playerKnowledge, followUp)
    if len(response.AddsItems) > 0 {
        for _, item := range response.AddsItems {
            g.AddItem(game.NewItemFromString(item))
        }
    }
    g.closeInputElement()
    g.flags.SetFlags(response.FlagsSet)

    flow := ConversationFlowContinue

    for _, effect := range response.Effects {
        effectFlow := g.handleDialogueEffect(npc, effect)
        flow = effectFlow
    }

    if flow == ConversationFlowQuit {
        g.openSpeechWindow(npc, response.Text, func() { g.modalElement = nil })
    } else if flow == ConversationFlowCloseInputOnly {
        g.openSpeechWindow(npc, response.Text, func() {})
    } else if flow == ConversationFlowContinue {
        g.openSpeechWindow(npc, response.Text, func() {
            var optionsMenuItems []renderer.MenuItem

            if len(response.ForcedChoice) > 0 {
                optionsMenuItems = g.toForcedMenuItems(npc, dialogue, response.ForcedChoice)
            } else {
                newOptions := dialogue.GetOptions(g.GetAvatar(), g.playerKnowledge, g.flags)
                optionsMenuItems = g.toMenuItems(npc, dialogue, newOptions)
            }
            if len(optionsMenuItems) > 0 {
                g.openConversationMenu(geometry.Point{X: 3, Y: 13}, optionsMenuItems)
            }
        })
    }
}

type ConversationFlow int

const (
    ConversationFlowContinue ConversationFlow = iota
    ConversationFlowQuit
    ConversationFlowCloseInputOnly
)

func (g *GridEngine) handleDialogueEffect(npc *game.Actor, effect string) ConversationFlow {
    switch effect {
    case "quits":
        return ConversationFlowQuit
    case "joins":
        g.AddToParty(npc)
        return ConversationFlowQuit
    case "sells":
        g.openVendorMenu(npc)
        return ConversationFlowCloseInputOnly
    }
    effectPredicate := recfile.StrPredicate(effect)
    if effectPredicate != nil {
        switch effectPredicate.Name() {
        case "trainsToLevel":
            maxLevel := effectPredicate.GetInt(0)
            g.openTrainerMenu(npc, maxLevel)
            return ConversationFlowQuit
        case "giveFood":
            amount := effectPredicate.GetInt(0)
            g.AddFood(amount)
        case "giveGold":
            amount := effectPredicate.GetInt(0)
            g.AddGold(amount)
        case "giveLockpicks":
            amount := effectPredicate.GetInt(0)
            g.AddLockpicks(amount)
        case "giveXP":
            amount := effectPredicate.GetInt(0)
            g.AddXP(amount)
        case "giveSkill":
            skill := effectPredicate.GetString(0)
            g.AddSkill(g.GetAvatar(), skill)
        case "giveBuff":
            buffType := game.BuffType(effectPredicate.GetString(0))
            name := effectPredicate.GetString(1)
            strength := effectPredicate.GetInt(2)
            g.AddBuff(g.GetAvatar(), name, buffType, strength)
        }
    }
    return ConversationFlowContinue
}
