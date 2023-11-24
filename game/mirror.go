package game

import (
    "Legacy/recfile"
    "Legacy/util"
    "image/color"
)

type Mirror struct {
    BaseObject
    isBroken   bool
    isMagical  bool
    brokenIcon int32
}

func (s *Mirror) TintColor() color.Color {
    return color.White
}

func (s *Mirror) Name() string {
    if s.isBroken {
        return "a broken mirror"
    }
    return "a mirror"
}

func (s *Mirror) ToRecordAndType() (recfile.Record, string) {
    return recfile.Record{
        {Name: "name", Value: s.name},
        {Name: "icon", Value: recfile.Int32Str(s.icon)},
        {Name: "pos", Value: s.Pos().Encode()},
        {Name: "isHidden", Value: recfile.BoolStr(s.isHidden)},
    }, "mirror"
}

func NewMirror(isMagical bool, isBroken bool) *Mirror {
    return &Mirror{
        BaseObject: BaseObject{
            icon: 196,
        },
        brokenIcon: 206,
        isBroken:   isBroken,
        isMagical:  isMagical,
    }
}

func (s *Mirror) Description() []string {
    if s.isBroken {
        return []string{
            "You see a broken mirror. It's still reflecting your face, but it's cracked.",
        }
    }
    if s.isMagical {
        return []string{
            "You see a mirror. It is glowing with a strange light.",
        }
    }
    return []string{
        "You see a mirror. It's reflecting your face.",
    }
}

func (s *Mirror) Icon(tick uint64) int32 {
    if s.isBroken {
        return s.brokenIcon
    }
    if s.isMagical {
        return s.frameFromTick(tick)
    }
    return s.icon
}
func (s *Mirror) IsWalkable(person *Actor) bool {
    return false
}

func (s *Mirror) IsTransparent() bool {
    return true
}

func (s *Mirror) GetContextActions(engine Engine) []util.MenuItem {
    baseExamine := s.BaseObject.GetContextActions(engine, s)
    travelAction := util.MenuItem{
        Text: "Touch the mirror",
        Action: func() {
            engine.AskUserForString("Where to? ", 14, func(text string) {
                engine.TeleportTo(text)
            })
        },
    }
    changeAppearanceAction := util.MenuItem{
        Text:   "Look into the mirror",
        Action: engine.ChangeAppearance,
    }
    if s.isMagical && !s.isBroken {
        return []util.MenuItem{travelAction, changeAppearanceAction}
    }
    return append(baseExamine, changeAppearanceAction)
}

func (s *Mirror) frameFromTick(tick uint64) int32 {
    magicalBaseIcon := int32(207)
    frame := util.GetFrameFromTick(tick, 0.5, 2)
    return magicalBaseIcon + frame
}
