package game

import (
    "Legacy/geometry"
    "Legacy/gridmap"
    "Legacy/recfile"
    "Legacy/util"
    "fmt"
    "image/color"
)

type Well struct {
    BaseObject
    hasRope    bool
    hasPlanks  bool
    transition gridmap.Transition
}

func (w *Well) TintColor() color.Color {
    return color.White
}

func (w *Well) Name() string {
    return "a well"
}

func (w *Well) ToRecordAndType() (recfile.Record, string) {
    return recfile.Record{
        {Name: "name", Value: w.name},
        {Name: "icon", Value: recfile.Int32Str(w.icon)},
        {Name: "pos", Value: w.Pos().Encode()},
        {Name: "isHidden", Value: recfile.BoolStr(w.isHidden)},
        {Name: "hasRope", Value: recfile.BoolStr(w.hasRope)},
        {Name: "hasPlanks", Value: recfile.BoolStr(w.hasPlanks)},
        {Name: "transition", Value: w.transition.Encode()},
    }, "well"
}

func NewWellFromRecord(record recfile.Record) *Well {
    well := NewWell()
    for _, field := range record {
        switch field.Name {
        case "name":
            well.name = field.Value
        case "icon":
            well.icon = field.AsInt32()
        case "pos":
            well.SetPos(geometry.MustDecodePoint(field.Value))
        case "isHidden":
            well.isHidden = field.AsBool()
        case "hasRope":
            well.hasRope = field.AsBool()
        case "hasPlanks":
            well.hasPlanks = field.AsBool()
        case "transition":
            well.transition = gridmap.MustDecodeTransition(field.Value)
        }
    }
    return well
}

func NewWell() *Well {
    return &Well{
        BaseObject: BaseObject{
            icon: 187,
        },
    }
}

func (w *Well) Description() []string {
    if w.hasRope {
        return []string{"A well with a rope"}
    }
    return []string{"A well"}
}

func (w *Well) Icon(uint64) int32 {
    icon := int32(231) // with rope & planks
    if !w.hasPlanks {
        icon = 233
    }
    if !w.hasRope {
        return icon - 1
    }
    return icon
}
func (w *Well) IsWalkable(person *Actor) bool {
    return false
}

func (w *Well) IsTransparent() bool {
    return true
}

func (w *Well) GetContextActions(engine Engine) []util.MenuItem {
    actions := w.BaseObject.GetContextActions(engine, w)
    if w.hasPlanks {
        actions = append(actions,
            util.MenuItem{
                Text: fmt.Sprintf("Take planks"),
                Action: func() {
                    if w.hasPlanks {
                        w.hasPlanks = false
                        engine.TakeItem(NewTool(ToolTypeWoodenPlanks, "wooden planks"))
                    }
                },
            })
    } else if engine.GetParty().HasTool(ToolTypeWoodenPlanks) {
        actions = append(actions,
            util.MenuItem{
                Text: fmt.Sprintf("Attach planks"),
                Action: func() {
                    if !w.hasPlanks {
                        w.hasPlanks = true
                        engine.GetParty().RemoveTool(ToolTypeWoodenPlanks)
                    }
                },
            })
    }
    if w.hasRope {
        actions = append(actions,
            util.MenuItem{
                Text: fmt.Sprintf("Take rope"),
                Action: func() {
                    if w.hasRope {
                        w.hasRope = false
                        engine.TakeItem(NewTool(ToolTypeRope, "a rope"))
                    }
                },
            })
    } else if engine.GetParty().HasTool(ToolTypeRope) {
        actions = append(actions,
            util.MenuItem{
                Text: fmt.Sprintf("Attach rope"),
                Action: func() {
                    if !w.hasRope {
                        w.hasRope = true
                        engine.GetParty().RemoveTool(ToolTypeRope)
                    }
                },
            })
    }
    if w.hasRope && !w.transition.IsEmpty() {
        actions = append(actions,
            util.MenuItem{
                Text: fmt.Sprintf("Climb down"),
                Action: func() {
                    engine.TransitionToNamedLocation(w.transition.TargetMap, w.transition.TargetLocation)
                },
            })
    }
    return actions
}

func (w *Well) SetRope(rope bool) {
    w.hasRope = rope
}

func (w *Well) SetPlanks(planks bool) {
    w.hasPlanks = planks
}

func (w *Well) SetTransitionTarget(transition gridmap.Transition) {
    w.transition = transition
}
