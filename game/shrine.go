package game

import (
    "Legacy/renderer"
    "fmt"
    "image/color"
    "strings"
)

type Principle string

const (
    Void    Principle = "Void"
    Wealth  Principle = "Wealth"
    Loyalty Principle = "Loyalty"
)

type Shrine struct {
    BaseObject
    name      string
    principle Principle
}

func (s *Shrine) TintColor() color.Color {
    return color.White
}

func (s *Shrine) Name() string {
    return s.name
}

func NewShrine(name string, principle Principle) *Shrine {
    return &Shrine{
        BaseObject: BaseObject{
            icon: 186,
        },
        principle: principle,
        name:      name,
    }
}

func (s *Shrine) Description() []string {
    return []string{
        "Somebody has built a shrine here.",
        "It is called:",
        s.name,
    }
}

func (s *Shrine) Icon(uint64) int {
    return s.icon
}
func (s *Shrine) IsWalkable(person *Actor) bool {
    return true
}

func (s *Shrine) IsTransparent() bool {
    return true
}

func (s *Shrine) GetContextActions(engine Engine) []renderer.MenuItem {
    actions := s.BaseObject.GetContextActions(engine, s)
    actions = append(actions, renderer.MenuItem{
        Text: fmt.Sprintf("Meditate at \"%s\"", s.Name()),
        Action: func() {
            engine.Flags().IncrementFlag(fmt.Sprintf("meditated_for_%s", strings.ToLower(string(s.principle))))
            engine.ShowScrollableText([]string{
                "You meditate at the shrine.",
                "Aligning yourself with",
                fmt.Sprintf("the principle of %s.", s.principle),
            }, color.White)
        },
    })
    return actions
}
