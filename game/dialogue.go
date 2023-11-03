package game

import (
    "Legacy/recfile"
    "fmt"
    "regexp"
    "sort"
    "strings"
)

// an NPC can respond to a number of keywords
// foreach keyword there is a condition for it to appear (knowledge of the player, world state, etc)
// a response can be a simple text, the end of the dialogue, or combat
// also every response can trigger a change in the world state

/*
Condition: Flag(has_key)
Condition: Skill(lockpick, 3)
Condition: Item(potion, 1)
Condition: Item(gold, 1000)
*/

type DialogueChoice struct {
    Text         string
    TransitionTo string
    NeededFlags  []string
    NeededSkills map[string]int
}
type ConversationNode struct {
    Text         []string
    FlagsSet     []string
    AddsKeywords []string
    Effects      []string
    ForcedChoice []DialogueChoice
    NeededFlags  []string
    NeededSkills map[string]int
    TriggerEvent string
    AddsItems    []string
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

func NewDialogueFromRecords(records []recfile.Record) *Dialogue {
    triggers := make(map[string]ConversationNode)
    for _, record := range records {
        currentTrigger := ""
        currentText := make([]string, 0)
        currentNode := ConversationNode{}
        currentOption := DialogueChoice{}
        for _, field := range record {
            fieldValue := field.Value
            fieldName := field.Name
            switch fieldName {
            case "Key":
                currentTrigger = fieldValue
            case "Text":
                currentText = strings.Split(fieldValue, "\n")
            case "Effect":
                currentNode.Effects = append(currentNode.Effects, fieldValue)
            case "GiveItem":
                currentNode.AddsItems = append(currentNode.AddsItems, fieldValue)
            case "NeedsFlag":
                currentNode.NeededFlags = append(currentNode.NeededFlags, fieldValue)
            case "TriggerEvent":
                currentNode.TriggerEvent = fieldValue
            case "SetsFlag":
                currentNode.FlagsSet = append(currentNode.FlagsSet, fieldValue)
            case "NeedsSkill":
                pred := recfile.StrPredicate(fieldValue)
                if pred != nil {
                    skillName := pred.Name()
                    skillLevel := pred.GetInt(0)
                    if currentNode.NeededSkills == nil {
                        currentNode.NeededSkills = make(map[string]int)
                    }
                    currentNode.NeededSkills[skillName] = skillLevel
                } else {
                    println(fmt.Sprintf("ERR: invalid skill: %s", fieldValue))
                }
            case "OptionNeedsFlag":
                currentOption.NeededFlags = append(currentOption.NeededFlags, fieldValue)
            case "OptionNeedsSkill":
                pred := recfile.StrPredicate(fieldValue)
                if pred != nil {
                    skillName := pred.Name()
                    skillLevel := pred.GetInt(0)
                    if currentOption.NeededSkills == nil {
                        currentOption.NeededSkills = make(map[string]int)
                    }
                    currentOption.NeededSkills[skillName] = skillLevel
                } else {
                    println(fmt.Sprintf("ERR: invalid skill: %s", fieldValue))
                }
            case "Option":
                parts := strings.Split(fieldValue, ":")
                if len(parts) != 2 {
                    println(fmt.Sprintf("ERR: invalid option: %s", fieldValue))
                } else {
                    currentOption.TransitionTo = strings.TrimSpace(parts[0])
                    currentOption.Text = strings.TrimSpace(parts[1])
                    currentNode.ForcedChoice = append(currentNode.ForcedChoice, currentOption)
                    currentOption = DialogueChoice{}
                }
            }
        }
        addedKeyWords, strippedText := parseKeywords(currentText)
        currentNode.Text = strippedText
        currentNode.AddsKeywords = addedKeyWords
        triggers[currentTrigger] = currentNode
    }
    return NewDialogue(triggers)
}

func (d *Dialogue) GetOptions(speaker *Actor, pk *PlayerKnowledge, flags *Flags) []string {
    var options []string

    for k, _ := range pk.knowsAbout {
        if node, ok := d.triggers[k]; ok {
            if len(node.NeededFlags) > 0 {
                if !flags.AllSet(node.NeededFlags) {
                    continue
                }
            }
            if len(node.NeededSkills) > 0 {
                if !speaker.GetSkills().HasSkills(node.NeededSkills) {
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

func (d *Dialogue) HasOpening() bool {
    if _, exists := d.triggers["_opening"]; exists {
        return true
    }
    return false
}

func (d *Dialogue) GetFirstTimeText() ConversationNode {
    return d.triggers["_first_time"]
}

func (d *Dialogue) GetOpening() ConversationNode {
    return d.triggers["_opening"]
}

func (d *Dialogue) HasBeenUsed(keyword string) bool {
    return d.previouslyAsked[keyword]
}

type JournalEntry struct {
    Time   uint64
    Text   []string
    Source string
}

type PlayerKnowledge struct {
    knowsAbout map[string]bool
    talkedTo   map[string]bool
    journal    map[string][]JournalEntry
}

func NewPlayerKnowledge() *PlayerKnowledge {
    return &PlayerKnowledge{
        knowsAbout: make(map[string]bool),
        talkedTo:   make(map[string]bool),
        journal:    make(map[string][]JournalEntry),
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

func (p *PlayerKnowledge) AddJournalEntry(source string, text []string, time uint64) {
    if _, exists := p.journal[source]; !exists {
        p.journal[source] = make([]JournalEntry, 0)
    }
    p.journal[source] = append(p.journal[source], JournalEntry{
        Time:   time,
        Text:   text,
        Source: source,
    })
}

func (p *PlayerKnowledge) GetChronologicalJournal() []string {
    entries := p.getSortedEntries()
    var result []string
    for index, entry := range entries {
        header := fmt.Sprintf("%s (%d)", entry.Source, entry.Time)
        result = append(result, header, "")
        result = append(result, entry.Text...)
        if index < len(entries)-1 {
            result = append(result, "", "")
        }
    }
    return result
}

func (p *PlayerKnowledge) GetJournalBySource(source string) []string {
    entries := p.journal[source]
    var result []string
    header := fmt.Sprintf("%s", source)
    result = append(result, header)
    for index, entry := range entries {
        result = append(result, fmt.Sprintf("at %d", entry.Time), "")
        result = append(result, entry.Text...)
        if index < len(entries)-1 {
            result = append(result, "", "")
        }
    }
    return result
}

func (p *PlayerKnowledge) getSortedEntries() []JournalEntry {
    var entries []JournalEntry
    for _, entry := range p.journal {
        entries = append(entries, entry...)
    }
    sort.SliceStable(entries, func(i, j int) bool {
        return entries[i].Time < entries[j].Time
    })
    return entries
}

func (p *PlayerKnowledge) GetJournalSources() []string {
    var sources []string
    for k, _ := range p.journal {
        sources = append(sources, k)
    }
    sort.SliceStable(sources, func(i, j int) bool {
        return sources[i] < sources[j]
    })
    return sources
}

func (p *PlayerKnowledge) IsJournalEmpty() bool {
    return len(p.journal) == 0
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
