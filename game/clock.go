package game

import (
    "Legacy/geometry"
    "Legacy/gridmap"
    "Legacy/recfile"
    "Legacy/util"
    "image/color"
)

type Clock struct {
    BaseObject
    transition gridmap.Transition
}

func (w *Clock) TintColor() color.Color {
    return color.White
}

func (w *Clock) Name() string {
    return "a clock"
}

func (w *Clock) ToRecordAndType() (recfile.Record, string) {
    return recfile.Record{
        {Name: "name", Value: w.name},
        {Name: "icon", Value: recfile.Int32Str(w.icon)},
        {Name: "pos", Value: w.Pos().Encode()},
        {Name: "isHidden", Value: recfile.BoolStr(w.isHidden)},
    }, "clock"
}

func NewClockFromRecord(record recfile.Record) *Clock {
    clock := NewClock()
    for _, field := range record {
        switch field.Name {
        case "name":
            clock.name = field.Value
        case "icon":
            clock.icon = field.AsInt32()
        case "pos":
            clock.SetPos(geometry.MustDecodePoint(field.Value))
        case "isHidden":
            clock.isHidden = field.AsBool()
        }
    }
    return clock
}

func NewClock() *Clock {
    return &Clock{
        BaseObject: BaseObject{
            icon: 187,
        },
    }
}

func (w *Clock) Description() []string {
    return []string{"a clock"}
}

func (w *Clock) Icon(tick uint64) int32 {
    frame := util.GetLoopingFrameFromTick(tick, 1, 2)
    if frame == 0 {
        return 234
    } else {
        return 235
    }
}

func (w *Clock) IsWalkable(person *Actor) bool {
    return false
}

func (w *Clock) IsTransparent() bool {
    return true
}

func (w *Clock) GetContextActions(engine Engine) []util.MenuItem {
    return []util.MenuItem{
        {
            Text: "Examine",
            Action: func() {
                engine.ShowScrollableText([]string{engine.GetWorldTime().GetTime()}, color.White, true)
            },
        },
    }
}

func (w *Clock) SetTransitionTarget(transition gridmap.Transition) {
    w.transition = transition
}
