package dialogue

import (
    "bufio"
    "io"
    "regexp"
    "sort"
    "strings"
)

// an NPC can respond to a number of keywords
// foreach keyword there is a condition for it to appear (knowledge of the player, world state, etc)
// a response can be a simple text, the end of the dialogue, or combat
// also every response can trigger a change in the world state

type Response struct {
    Text         []string
    AddsKeywords []string
    Effect       string
}
type Dialogue struct {
    triggers        map[string]Response
    previouslyAsked map[string]bool
    keyWordsGiven   map[string]bool
}

func NewDialogueFromFile(dialogueFile io.Reader) *Dialogue {
    return &Dialogue{
        triggers:        triggersFromFile(dialogueFile),
        previouslyAsked: make(map[string]bool),
        keyWordsGiven:   make(map[string]bool),
    }
}

func (d *Dialogue) GetOptions(pk *PlayerKnowledge) []string {
    var options []string

    for k, _ := range pk.knowsAbout {
        if _, ok := d.triggers[k]; ok {
            options = append(options, k)
        }
    }

    sort.SliceStable(options, func(i, j int) bool {
        return options[i] < options[j]
    })
    // prepend "name", "job"
    options = append([]string{"name", "job"}, options...)
    return options
}

func (d *Dialogue) GetResponseAndAddKnowledge(pk *PlayerKnowledge, keyword string) ([]string, string) {
    d.previouslyAsked[keyword] = true
    response := d.triggers[keyword]
    pk.AddKnowledge(response.AddsKeywords)
    d.RememberKeywords(response.AddsKeywords)
    return response.Text, response.Effect
}

func (d *Dialogue) RememberKeywords(keywords []string) {
    for _, k := range keywords {
        d.keyWordsGiven[k] = true
    }
}

type PlayerKnowledge struct {
    knowsAbout map[string]bool
}

func NewPlayerKnowledge() *PlayerKnowledge {
    return &PlayerKnowledge{
        knowsAbout: make(map[string]bool),
    }
}

func (p *PlayerKnowledge) AddKnowledge(knowledge []string) {
    for _, k := range knowledge {
        p.knowsAbout[k] = true
    }
}

func triggersFromFile(file io.Reader) map[string]Response {
    // line by line
    scanner := bufio.NewScanner(file)
    triggers := make(map[string]Response)
    currentTrigger := ""
    currentOption := ""
    currentText := make([]string, 0)
    for scanner.Scan() {
        line := scanner.Text()
        if len(line) == 0 {
            currentText = append(currentText, line)
        } else if line[0] == ';' {
            continue
        } else if line[0] == '#' && line[1] == ' ' {
            if currentTrigger != "" {
                addedKeyWords, strippedText := parseKeywords(currentText)
                currentText = make([]string, 0)
                triggers[currentTrigger] = Response{
                    Text:         strippedText,
                    Effect:       currentOption,
                    AddsKeywords: addedKeyWords,
                }
                currentOption = ""
            }
            currentTrigger = line[2:]
        } else if line[0] == '#' && line[1] == '#' {
            currentOption = line[3:]
        } else {
            currentText = append(currentText, line)
        }
    }
    if currentTrigger != "" {
        addedKeyWords, strippedText := parseKeywords(currentText)
        triggers[currentTrigger] = Response{
            Text:         strippedText,
            Effect:       currentOption,
            AddsKeywords: addedKeyWords,
        }
    }
    return triggers
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
    return addedKeyWords, strippedText
}
