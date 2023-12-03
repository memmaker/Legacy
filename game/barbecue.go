package game

import (
    "Legacy/geometry"
    "Legacy/recfile"
    "Legacy/util"
    "fmt"
    "image/color"
)

type Barbecue struct {
    BaseObject
    foodCount int
}

func (s *Barbecue) TintColor() color.Color {
    return color.White
}

func (s *Barbecue) Name() string {
    return "a barbecue"
}

func (s *Barbecue) ToRecordAndType() (recfile.Record, string) {
    return recfile.Record{
        {Name: "name", Value: s.name},
        {Name: "icon", Value: recfile.Int32Str(s.icon)},
        {Name: "pos", Value: s.Pos().Encode()},
        {Name: "isHidden", Value: recfile.BoolStr(s.isHidden)},
        {Name: "foodCount", Value: recfile.IntStr(s.foodCount)},
    }, "barbecue"
}

func NewBarbecueFromRecord(record recfile.Record) *Barbecue {
    barbecue := NewBarbecue(0)
    for _, field := range record {
        switch field.Name {
        case "name":
            barbecue.name = field.Value
        case "icon":
            barbecue.icon = field.AsInt32()
        case "pos":
            barbecue.SetPos(geometry.MustDecodePoint(field.Value))
        case "isHidden":
            barbecue.isHidden = field.AsBool()
        case "foodCount":
            barbecue.foodCount = field.AsInt()
        }
    }
    return barbecue
}

func NewBarbecue(foodCount int) *Barbecue {
    return &Barbecue{
        BaseObject: BaseObject{
            icon: 187,
        },
        foodCount: foodCount,
    }
}

func (s *Barbecue) Description() []string {
    if s.foodCount > 0 {
        return []string{
            "This fireplace is burning.",
            "There is some food on the fire.",
            fmt.Sprintf("Should be about %d rations.", s.foodCount),
        }
    } else {
        return []string{
            "This fireplace is burning.",
            "There is nothing on the fire.",
        }
    }
}

func (s *Barbecue) Icon(uint64) int32 {
    return s.icon
}
func (s *Barbecue) IsWalkable(person *Actor) bool {
    return false
}

func (s *Barbecue) IsTransparent() bool {
    return true
}

func (s *Barbecue) GetContextActions(engine Engine) []util.MenuItem {
    actions := s.BaseObject.GetContextActions(engine, s)
    if s.foodCount > 0 {
        partySize := engine.GetPartySize()
        actions = append(actions,
            util.MenuItem{
                Text: fmt.Sprintf("Take 1 ration"),
                Action: func() {
                    s.foodCount--
                    engine.AddFood(1)
                },
            })
        if s.foodCount >= partySize {
            actions = append(actions,
                util.MenuItem{
                    Text: fmt.Sprintf("Take enough rations for the party"),
                    Action: func() {
                        engine.AddFood(partySize)
                        s.foodCount -= partySize
                    },
                })

            actions = append(actions,
                util.MenuItem{
                    Text: fmt.Sprintf("Take half of the rations"),
                    Action: func() {
                        amount := s.foodCount / 2
                        engine.AddFood(amount)
                        s.foodCount -= amount
                    },
                })
        }
        actions = append(actions,
            util.MenuItem{
                Text: fmt.Sprintf("Take all of the rations"),
                Action: func() {
                    engine.AddFood(s.foodCount)
                    s.foodCount = 0
                },
            })
    }
    return actions
}
