package game

import (
    "Legacy/geometry"
    "Legacy/recfile"
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
func (s *Shrine) ToRecordAndType() (recfile.Record, string) {
    return recfile.Record{
        {Name: "name", Value: s.name},
        {Name: "icon", Value: recfile.Int32Str(s.icon)},
        {Name: "pos", Value: s.Pos().Encode()},
        {Name: "isHidden", Value: recfile.BoolStr(s.isHidden)},
        {Name: "principle", Value: string(s.principle)},
    }, "shrine"
}

func NewShrineFromRecord(record recfile.Record) *Shrine {
    shrine := NewShrine("", Void)
    for _, field := range record {
        switch field.Name {
        case "name":
            shrine.name = field.Value
        case "icon":
            shrine.icon = field.AsInt32()
        case "pos":
            shrine.SetPos(geometry.MustDecodePoint(field.Value))
        case "isHidden":
            shrine.isHidden = field.AsBool()
        case "principle":
            shrine.principle = Principle(field.Value)
        }
    }
    return shrine
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

func (s *Shrine) Icon(uint64) int32 {
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
        Text: "Meditate",
        Action: func() {
            engine.Flags().IncrementFlag(fmt.Sprintf("meditated_for_%s", strings.ToLower(string(s.principle))))
            engine.ShowColoredText([]string{
                "You meditate at the shrine.",
                "Aligning yourself with",
                fmt.Sprintf("the principle of %s.", s.principle),
            }, color.White, true)
        },
    })
    return actions
}
