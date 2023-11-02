package game

import (
    "Legacy/geometry"
    "Legacy/recfile"
    "Legacy/renderer"
    "fmt"
    "image/color"
)

type FirePlace struct {
    BaseObject
    foodCount int
}

func (s *FirePlace) TintColor() color.Color {
    return color.White
}

func (s *FirePlace) Name() string {
    return "a fireplace"
}

func (s *FirePlace) ToRecordAndType() (recfile.Record, string) {
    return recfile.Record{
        {Name: "name", Value: s.name},
        {Name: "icon", Value: recfile.Int32Str(s.icon)},
        {Name: "pos", Value: s.Pos().Encode()},
        {Name: "isHidden", Value: recfile.BoolStr(s.isHidden)},
        {Name: "foodCount", Value: recfile.IntStr(s.foodCount)},
    }, "fireplace"
}

func NewFireplaceFromRecord(record recfile.Record) *FirePlace {
    firePlace := NewFireplace(0)
    for _, field := range record {
        switch field.Name {
        case "name":
            firePlace.name = field.Value
        case "icon":
            firePlace.icon = field.AsInt32()
        case "pos":
            firePlace.SetPos(geometry.MustDecodePoint(field.Value))
        case "isHidden":
            firePlace.isHidden = field.AsBool()
        case "foodCount":
            firePlace.foodCount = field.AsInt()
        }
    }
    return firePlace
}

func NewFireplace(foodCount int) *FirePlace {
    return &FirePlace{
        BaseObject: BaseObject{
            icon: 187,
        },
        foodCount: foodCount,
    }
}

func (s *FirePlace) Description() []string {
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

func (s *FirePlace) Icon(uint64) int32 {
    return s.icon
}
func (s *FirePlace) IsWalkable(person *Actor) bool {
    return false
}

func (s *FirePlace) IsTransparent() bool {
    return true
}

func (s *FirePlace) GetContextActions(engine Engine) []renderer.MenuItem {
    actions := s.BaseObject.GetContextActions(engine, s)
    if s.foodCount > 0 {
        partySize := engine.GetPartySize()
        actions = append(actions,
            renderer.MenuItem{
                Text: fmt.Sprintf("Take 1 ration"),
                Action: func() {
                    s.foodCount--
                    engine.Print("You take 1 ration from the fire.")
                    engine.AddFood(1)
                },
            })
        if s.foodCount >= partySize {
            actions = append(actions,
                renderer.MenuItem{
                    Text: fmt.Sprintf("Take enough rations for the party"),
                    Action: func() {
                        engine.Print(fmt.Sprintf("You take %d rations from the fire.", partySize))
                        engine.AddFood(partySize)
                        s.foodCount -= partySize
                    },
                })

            actions = append(actions,
                renderer.MenuItem{
                    Text: fmt.Sprintf("Take half of the rations"),
                    Action: func() {
                        amount := s.foodCount / 2
                        engine.Print(fmt.Sprintf("You take %d rations from the fire.", amount))
                        engine.AddFood(amount)
                        s.foodCount -= amount
                    },
                })
        }
        actions = append(actions,
            renderer.MenuItem{
                Text: fmt.Sprintf("Take all of the rations"),
                Action: func() {
                    engine.Print("You take all rations from the fire.")
                    engine.AddFood(s.foodCount)
                    s.foodCount = 0
                },
            })
    }
    return actions
}
