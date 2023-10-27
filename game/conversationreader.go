package game

import "regexp"

type ConversationReader struct {
    triggers             map[string]ConversationNode
    currentTrigger       string
    currentOption        string
    currentText          []string
    forcedOptionRegex    *regexp.Regexp
    currentForcedOptions []DialogueChoice
    currentNeededFlags   []string
    flagRegex            *regexp.Regexp
}

func NewConversationReader() ConversationReader {
    // #! yes: _spare_time
    //#! no: _no_spare_time
    regexForOptions := regexp.MustCompile(`^#! (?P<effect>[a-zA-Z0-9_]+): (?P<trigger>[a-zA-Z0-9_]+)$`)
    // #f can_talk_to_ghosts
    regexForFlags := regexp.MustCompile(`^#f (?P<flag>[a-zA-Z0-9_]+)$`)
    return ConversationReader{
        triggers:             make(map[string]ConversationNode),
        currentTrigger:       "",
        currentOption:        "",
        currentText:          make([]string, 0),
        forcedOptionRegex:    regexForOptions,
        flagRegex:            regexForFlags,
        currentForcedOptions: make([]DialogueChoice, 0),
        currentNeededFlags:   make([]string, 0),
    }
}
func (c *ConversationReader) read(line string) {
    if len(line) == 0 {
        c.currentText = append(c.currentText, line)
    } else if line[0] == ';' {
        return
    } else if line[0] == '#' && line[1] == ' ' {
        if c.currentTrigger != "" {
            addedKeyWords, strippedText := parseKeywords(c.currentText)
            c.currentText = make([]string, 0)
            c.triggers[c.currentTrigger] = ConversationNode{
                Text:         strippedText,
                Effect:       c.currentOption,
                AddsKeywords: addedKeyWords,
                ForcedChoice: c.currentForcedOptions,
                NeededFlags:  c.currentNeededFlags,
            }
            c.currentForcedOptions = make([]DialogueChoice, 0)
            c.currentNeededFlags = make([]string, 0)
            c.currentOption = ""
        }
        c.currentTrigger = line[2:]
    } else if line[0] == '#' && line[1] == '#' {
        c.currentOption = line[3:]
    } else if c.flagRegex.MatchString(line) {
        // get key and value
        matches := c.flagRegex.FindStringSubmatch(line)
        flag := matches[1]
        c.currentNeededFlags = append(c.currentNeededFlags, flag)
    } else if c.forcedOptionRegex.MatchString(line) {
        // get key and value
        matches := c.forcedOptionRegex.FindStringSubmatch(line)
        label := matches[1]
        followUpTrigger := matches[2]
        c.currentForcedOptions = append(c.currentForcedOptions, DialogueChoice{
            Label:           label,
            FollowUpTrigger: followUpTrigger,
        })
    } else {
        c.currentText = append(c.currentText, line)
    }
}
func (c *ConversationReader) end() map[string]ConversationNode {
    if c.currentTrigger != "" {
        addedKeyWords, strippedText := parseKeywords(c.currentText)
        c.triggers[c.currentTrigger] = ConversationNode{
            Text:         strippedText,
            Effect:       c.currentOption,
            AddsKeywords: addedKeyWords,
            ForcedChoice: c.currentForcedOptions,
            NeededFlags:  c.currentNeededFlags,
        }
    }
    return c.triggers
}
