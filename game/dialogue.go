package game

import (
    "regexp"
    "sort"
    "strings"
)

// an NPC can respond to a number of keywords
// foreach keyword there is a condition for it to appear (knowledge of the player, world state, etc)
// a response can be a simple text, the end of the dialogue, or combat
// also every response can trigger a change in the world state

type DialogueChoice struct {
    Label           string
    FollowUpTrigger string
}
type ConversationNode struct {
    Text         []string
    AddsKeywords []string
    Effect       string
    ForcedChoice []DialogueChoice
    NeededFlags  []string
}
type Dialogue struct {
    triggers        map[string]ConversationNode
    previouslyAsked map[string]bool
    keyWordsGiven   map[string]bool
}

func NewDialogue(triggers map[string]ConversationNode) *Dialogue {
    return &Dialogue{
        triggers:        triggers,
        previouslyAsked: make(map[string]bool),
        keyWordsGiven:   make(map[string]bool),
    }
}

func (d *Dialogue) GetOptions(pk *PlayerKnowledge, flags *Flags) []string {
    var options []string

    for k, _ := range pk.knowsAbout {
        if node, ok := d.triggers[k]; ok {
            if len(node.NeededFlags) > 0 {
                if !flags.AllSet(node.NeededFlags) {
                    continue
                }
            }
            options = append(options, k)
        }
    }

    sort.SliceStable(options, func(i, j int) bool {
        return options[i] < options[j]
    })
    // prepend "name", "job"
    defaultOptions := []string{"job", "name"}
    for _, k := range defaultOptions {
        if node, ok := d.triggers[k]; ok {
            if len(node.NeededFlags) > 0 {
                if !flags.AllSet(node.NeededFlags) {
                    continue
                }
            }
            // prepend to the slice
            options = append([]string{k}, options...)
        }
    }
    return options
}

func (d *Dialogue) GetResponseAndAddKnowledge(pk *PlayerKnowledge, keyword string) ConversationNode {
    d.previouslyAsked[keyword] = true
    response := d.triggers[keyword]
    pk.AddKnowledge(response.AddsKeywords)
    d.RememberKeywords(response.AddsKeywords)
    return response
}

func (d *Dialogue) RememberKeywords(keywords []string) {
    for _, k := range keywords {
        d.keyWordsGiven[k] = true
    }
}

func (d *Dialogue) HasFirstTimeText() bool {
    if _, exists := d.triggers["_first_time"]; exists {
        return true
    }
    return false
}

func (d *Dialogue) GetFirstTimeText() []string {
    return d.triggers["_first_time"].Text
}

func (d *Dialogue) HasBeenUsed(keyword string) bool {
    return d.previouslyAsked[keyword]
}

type PlayerKnowledge struct {
    knowsAbout map[string]bool
    talkedTo   map[string]bool
}

func NewPlayerKnowledge() *PlayerKnowledge {
    return &PlayerKnowledge{
        knowsAbout: make(map[string]bool),
        talkedTo:   make(map[string]bool),
    }
}

func (p *PlayerKnowledge) AddKnowledge(knowledge []string) {
    for _, k := range knowledge {
        p.knowsAbout[k] = true
    }
}

func (p *PlayerKnowledge) AddTalkedTo(name string) {
    p.talkedTo[name] = true
}

func (p *PlayerKnowledge) HasTalkedTo(name string) bool {
    return p.talkedTo[name]
}

func parseKeywords(text []string) ([]string, []string) {
    var addedKeyWords []string
    var strippedText []string
    found := false
    // find parts of the string that is surrounded with =example= (equal sign)
    regex := regexp.MustCompile(`=[a-zA-Z0-9 ]+=`)
    for _, line := range text {
        // find all matches
        matches := regex.FindAllString(line, -1)
        // replace all matches with the text inside the match
        for _, match := range matches {
            keyword := match[1 : len(match)-1]
            line = strings.ReplaceAll(line, match, keyword)
            addedKeyWords = append(addedKeyWords, keyword)
            found = true
        }
        strippedText = append(strippedText, line)
    }
    if !found {
        return nil, text
    }
    if lastLine := strippedText[len(strippedText)-1]; lastLine == "" {
        strippedText = strippedText[:len(strippedText)-1]
    }
    return addedKeyWords, strippedText
}
