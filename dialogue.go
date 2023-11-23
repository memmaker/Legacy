package main

import (
    "Legacy/ega"
    "Legacy/game"
    "Legacy/recfile"
    "Legacy/ui"
    "Legacy/util"
    "fmt"
)

func (g *GridEngine) openSpeechWindow(npc *game.Actor) {
    toJournal := func(text []string) {
        g.addToJournal(npc.Name(), text)
    }
    g.conversationModal = ui.NewConversationModal(g.gridRenderer, toJournal)
    g.conversationModal.SetIcon(npc.Icon(0))
    g.conversationModal.SetTitle(npc.Name())
}

func (g *GridEngine) StartConversation(npc *game.Actor, loadedDialogue *game.Dialogue) {
    // NOTE: Conversations can have a line length of 27 chars
    if !npc.IsAlive() {
        g.Print(fmt.Sprintf("%s is dead.", npc.Name()))
        return
    }
    if npc.IsSleeping() {
        g.Print(fmt.Sprintf("%s is sleeping.", npc.Name()))
        return
    }
    if loadedDialogue == nil {
        g.Print(fmt.Sprintf("\"%s\" has nothing to say.", npc.Name()))
        return
    }

    firstNode := "_opening"
    if !g.playerKnowledge.HasTalkedTo(npc.GetInternalName()) && loadedDialogue.HasFirstTimeText() {
        g.playerKnowledge.AddTalkedTo(npc.GetInternalName())
        if loadedDialogue.HasFirstTimeText() {
            firstNode = "_first_time"
        }
    }
    startedConversationFlag := fmt.Sprintf("talked_to_%s", npc.GetInternalName())
    g.flags.IncrementFlag(startedConversationFlag)

    response := loadedDialogue.GetResponseAndAddKnowledge(g.GetAvatar(), g.playerKnowledge, g, firstNode)
    g.handleDialogueChoice(loadedDialogue, response, npc)
}

func (g *GridEngine) ShowMultipleChoiceDialogue(icon int32, text [][]string, choices []util.MenuItem) {
    toJournal := func(page []string) {
        g.addToJournal(g.currentMap.GetDisplayName(), page)
    }
    g.conversationModal = ui.NewConversationModal(g.gridRenderer, toJournal)
    g.conversationModal.SetIcon(icon)
    g.conversationModal.SetText(text)
    g.conversationModal.SetOptions(choices)
    g.conversationModal.OnMouseMoved(g.lastMousePosX, g.lastMousePosY)
}

func (g *GridEngine) toMenuItems(npc *game.Actor, dialogue *game.Dialogue, options []string) []util.MenuItem {
    var items []util.MenuItem
    for _, o := range options {
        option := o
        textColor := ega.BrightYellow
        if dialogue.HasBeenUsed(option) {
            textColor = ega.BrightWhite
        }
        items = append(items, util.MenuItem{
            Text: option,
            Action: func() {
                response := dialogue.GetResponseAndAddKnowledge(g.GetAvatar(), g.playerKnowledge, g, option)
                g.handleDialogueChoice(dialogue, response, npc)
            },
            TextColor: textColor,
        })
    }
    return items
}

func (g *GridEngine) toForcedMenuItems(npc *game.Actor, dialogue *game.Dialogue, choices []game.DialogueChoice) []util.MenuItem {
    var items []util.MenuItem
    for _, c := range choices {
        if !game.EvalConditionals(g.GetAvatar(), g, c.Conditionals) {
            continue
        }

        choice := c

        label := choice.Text
        if choice.SkillCheck != nil {
            baseDifficulty := choice.SkillCheck.Difficulty
            if choice.SkillCheck.IsVersusAntagonist {
                baseDifficulty = npc.GetAntagonistDifficultyByAttribute(choice.SkillCheck.VersusAttribute)
            }
            label = fmt.Sprintf("%s (%s - %s)", label, choice.SkillCheck.SkillName, g.GetRelativeDifficulty(choice.SkillCheck.SkillName, baseDifficulty).ToString())
        }
        items = append(items, util.MenuItem{
            Text: label,
            Action: func() {
                transitionTarget := choice.TransitionOnSuccess
                // do the checks..
                checksFailed := false
                if len(choice.Checks) > 0 {
                    checksFailed = !game.EvalConditionals(g.GetAvatar(), g, choice.Checks)
                }
                if choice.SkillCheck != nil {
                    if choice.SkillCheck.IsVersusAntagonist {
                        if !g.SkillCheckAvatarVs(choice.SkillCheck.SkillName, npc, choice.SkillCheck.VersusAttribute) {
                            checksFailed = true
                        }
                    } else {
                        if !g.SkillCheckAvatar(choice.SkillCheck.SkillName, choice.SkillCheck.Difficulty) {
                            checksFailed = true
                        }
                    }
                }
                if checksFailed {
                    transitionTarget = choice.TransitionOnFail
                }
                if transitionTarget == "" {
                    println("ERR: No transition target for choice", choice.Text)
                }
                response := dialogue.GetResponseAndAddKnowledge(g.GetAvatar(), g.playerKnowledge, g, transitionTarget)
                g.handleDialogueChoice(dialogue, response, npc)
            },
        })
    }
    return items
}

func (g *GridEngine) handleDialogueChoice(dialogue *game.Dialogue, response game.ConversationNode, npc *game.Actor) {
    if response.IsEmpty() {
        g.closeConversation()
        println("ERR: Empty response")
        return
    }
    if g.conversationModal == nil {
        g.openSpeechWindow(npc)
    }

    flow := ConversationFlowContinue
    var effectCalls []func()
    for _, effect := range response.Effects {
        effectCall, effectFlow := g.handleDialogueEffect(npc, effect)
        if effectCall != nil {
            effectCalls = append(effectCalls, effectCall)
        }
        if effectFlow > flow {
            flow = effectFlow
        }
    }
    if len(effectCalls) > 0 && flow != ConversationFlowEffectAfterLastPage {
        callAll(effectCalls)
    }
    if flow == ConversationFlowJustQuit {
        g.conversationModal.SetOnClose(func() {
            g.closeConversation() // this MUST be the only way to close a conversation
        })
    } else if flow == ConversationFlowQuitAfterText {
        g.conversationModal.SetText(response.Text)
        g.conversationModal.SetOptions(nil)
        g.conversationModal.SetOnClose(func() {
            g.closeConversation() // this MUST be the only way to close a conversation
        })
    } else if flow == ConversationFlowVendor { // what's the difference between this and ConversationFlowQuit?
        g.conversationModal.SetText(response.Text)
    } else if flow == ConversationFlowEffectAfterLastPage {
        g.conversationModal.SetText(response.Text)
        g.conversationModal.SetOptions(nil) // implies end of conversation
        g.conversationModal.SetOnClose(func() {
            callAll(effectCalls)
        })
    } else if flow == ConversationFlowContinue {
        var optionsMenuItems []util.MenuItem
        if len(response.ForcedChoice) > 0 {
            optionsMenuItems = g.toForcedMenuItems(npc, dialogue, response.ForcedChoice)
        } else {
            newOptions := dialogue.GetOptions(g.GetAvatar(), g.playerKnowledge, g)
            optionsMenuItems = g.toMenuItems(npc, dialogue, newOptions)
        }
        g.conversationModal.SetText(response.Text)
        g.conversationModal.SetOptions(optionsMenuItems)
    }
    g.conversationModal.OnMouseMoved(g.lastMousePosX, g.lastMousePosY)
}

func callAll(effectCalls []func()) {
    for _, effectCall := range effectCalls {
        effectCall()
    }
}

type ConversationFlow int

const (
    ConversationFlowContinue ConversationFlow = iota
    ConversationFlowVendor
    ConversationFlowQuitAfterText
    ConversationFlowJustQuit
    ConversationFlowEffectAfterLastPage
)

func (g *GridEngine) handleDialogueEffect(npc *game.Actor, effect string) (func(), ConversationFlow) {
    switch effect {
    case "quits":
        return nil, ConversationFlowQuitAfterText
    case "joins":
        return func() { g.AddToParty(npc) }, ConversationFlowEffectAfterLastPage
    case "sells":
        return func() { g.openVendorMenu(npc) }, ConversationFlowEffectAfterLastPage
    case "combat":
        return func() { g.EnemyStartsCombat(npc) }, ConversationFlowEffectAfterLastPage
    }
    effectPredicate := recfile.StrPredicate(effect)
    if effectPredicate != nil {
        switch effectPredicate.Name() {
        case "addKeyword":
            keyword := effectPredicate.GetString(0)
            g.playerKnowledge.AddKnowledge([]string{keyword})
        case "trainsToLevel":
            maxLevel := effectPredicate.GetInt(0)
            g.openTrainerMenu(npc, maxLevel)
            return nil, ConversationFlowJustQuit
        case "giveXP":
            amount := effectPredicate.GetInt(0)
            g.AddXP(amount)
        case "giveSkill":
            skill := effectPredicate.GetString(0)
            g.AddSkill(g.GetAvatar(), skill)
        case "removeTrigger":
            triggerName := effectPredicate.GetString(0)
            g.GetGridMap().RemoveNamedTrigger(triggerName)
        case "setFlag":
            flagName := effectPredicate.GetString(0)
            g.flags.IncrementFlag(flagName)
        case "setFlagTo":
            flagName := effectPredicate.GetString(0)
            g.flags.SetFlag(flagName, effectPredicate.GetInt(1))
        case "triggerEvent":
            eventName := effectPredicate.GetString(0)
            g.TriggerEvent(eventName)
        case "giveItem":
            item, hasItem := npc.GetItemByName(effectPredicate.GetString(0))
            if hasItem {
                npc.RemoveItem(item)
                g.AddItem(item)
            }
        case "giveBuff":
            buffType := game.BuffType(effectPredicate.GetString(0))
            name := effectPredicate.GetString(1)
            strength := effectPredicate.GetInt(2)
            g.AddBuff(g.GetAvatar(), name, buffType, strength)
        default:
            println("Unknown effect:", effect)
        }
    }

    return nil, ConversationFlowContinue
}
