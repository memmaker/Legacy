package main

import (
    "Legacy/ega"
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/renderer"
    "fmt"
)

func (g *GridEngine) openIconWindow(icon int, text []string, onLastPage func()) {
    if len(text) == 0 {
        return
    }
    g.modalElement = renderer.NewMultiPageWindow(g.gridRenderer, 3, icon, text, onLastPage)
}

func (g *GridEngine) openConversationMenu(topLeft geometry.Point, items []renderer.MenuItem) {
    g.inputElement = renderer.NewGridDialogueMenu(g.gridRenderer, topLeft, items)
    g.inputElement.OnMouseMoved(g.lastMousePosX, g.lastMousePosY)
}
func (g *GridEngine) StartConversation(npc *game.Actor) {
    startedConversationFlag := fmt.Sprintf("talked_to_%s", npc.GetInternalName())
    g.flags.IncrementFlag(startedConversationFlag)
    // NOTE: Conversations can have a line length of 27 chars
    if !npc.HasDialogue() {
        g.Print(fmt.Sprintf("\"%s\" has nothing to say.", npc.Name()))
        return
    }
    loadedDialogue := npc.GetDialogue()
    options := loadedDialogue.GetOptions(g.playerKnowledge, g.flags)

    var openingLines []string
    var items []renderer.MenuItem

    if !g.playerKnowledge.HasTalkedTo(npc.Name()) && loadedDialogue.HasFirstTimeText() {
        g.playerKnowledge.AddTalkedTo(npc.Name())
        opening := loadedDialogue.GetOpening()
        openingLines = opening.Text
        if len(opening.ForcedChoice) > 0 {
            items = g.toForcedMenuItems(npc, loadedDialogue, opening.ForcedChoice)
        }
    } else {
        openingLines = npc.LookDescription()
    }

    if len(items) == 0 {
        items = g.toMenuItems(npc, loadedDialogue, options)
    }

    g.openIconWindow(npc.Icon(0), openingLines, func() {
        if len(items) > 0 {
            g.openConversationMenu(geometry.Point{X: 3, Y: 13}, items)
        }
    })
}
func (g *GridEngine) ShowMultipleChoiceDialogue(icon int, text []string, choices []renderer.MenuItem) {
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
    g.inputElement = gridMenu
    g.inputElement.OnMouseMoved(g.lastMousePosX, g.lastMousePosY)
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
        choice := c
        items = append(items, renderer.MenuItem{
            Text: choice.Label,
            Action: func() {
                g.handleDialogueChoice(dialogue, choice.FollowUpTrigger, npc)
            },
        })
    }
    return items
}

func (g *GridEngine) handleDialogueChoice(dialogue *game.Dialogue, followUp string, npc *game.Actor) {
    response := dialogue.GetResponseAndAddKnowledge(g.playerKnowledge, followUp)
    g.inputElement = nil
    quitsDialogue := g.handleDialogueEffect(npc, response.Effect)
    if quitsDialogue {
        g.openIconWindow(npc.Icon(0), response.Text, func() { g.modalElement = nil })
    } else {
        g.openIconWindow(npc.Icon(0), response.Text, func() {
            var optionsMenuItems []renderer.MenuItem

            if len(response.ForcedChoice) > 0 {
                optionsMenuItems = g.toForcedMenuItems(npc, dialogue, response.ForcedChoice)
            } else {
                newOptions := dialogue.GetOptions(g.playerKnowledge, g.flags)
                optionsMenuItems = g.toMenuItems(npc, dialogue, newOptions)
            }
            if len(optionsMenuItems) > 0 {
                g.openConversationMenu(geometry.Point{X: 3, Y: 13}, optionsMenuItems)
            }
        })
    }
}

func (g *GridEngine) handleDialogueEffect(npc *game.Actor, effect string) bool {
    switch effect {
    case "quits":
        return true
    case "joins":
        g.AddToParty(npc)
        return false
    case "sells":
        g.openVendorMenu(npc)
        return true
    }
    return false
}
