package game

import (
    "Legacy/recfile"
    "Legacy/util"
    "image/color"
)

type Tombstone struct {
    BaseObject
    isHoly bool
}

func (s *Tombstone) TintColor() color.Color {
    return color.White
}

func (s *Tombstone) Name() string {
    if s.isHoly {
        return "a tombstone with a cross"
    }
    return "a tombstone"
}

func (s *Tombstone) ToRecordAndType() (recfile.Record, string) {
    return recfile.Record{
        {Name: "name", Value: s.name},
        {Name: "icon", Value: recfile.Int32Str(s.icon)},
        {Name: "pos", Value: s.Pos().Encode()},
        {Name: "isHidden", Value: recfile.BoolStr(s.isHidden)},
        {Name: "isHoly", Value: recfile.BoolStr(s.isHoly)},
        {Name: "description", Value: recfile.StringsStr(s.description)},
    }, "tombstone"
}

func NewTombstone(isHoly bool) *Tombstone {
    icon := int32(215)
    if isHoly {
        icon = 216
    }
    return &Tombstone{
        BaseObject: BaseObject{
            icon: icon,
        },
        isHoly: isHoly,
    }
}

func (s *Tombstone) Description() []string {
    if len(s.description) > 0 {
        return s.description
    }
    return []string{
        "You see a nameless tombstone.",
    }
}

func (s *Tombstone) Icon(tick uint64) int32 {
    return s.icon
}
func (s *Tombstone) IsWalkable(person *Actor) bool {
    return true
}

func (s *Tombstone) IsTransparent() bool {
    return true
}

func (s *Tombstone) GetContextActions(engine Engine) []util.MenuItem {
    return []util.MenuItem{
        {
            Text: "Examine",
            Action: func() {
                engine.ShowScrollableText(s.Description(), color.White, false)
            },
        },
    }
}
