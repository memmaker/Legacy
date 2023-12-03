package game

import (
    "Legacy/geometry"
    "Legacy/recfile"
    "Legacy/util"
    "image/color"
)

type Direction string

func (d Direction) ToPoint() geometry.Point {
    switch d {
    case DirectionNorth:
        return geometry.Point{X: 0, Y: -1}
    case DirectionSouth:
        return geometry.Point{X: 0, Y: 1}
    case DirectionEast:
        return geometry.Point{X: 1, Y: 0}
    case DirectionWest:
        return geometry.Point{X: -1, Y: 0}
    }
    return geometry.Point{}
}

const (
    DirectionNone  Direction = "None"
    DirectionNorth Direction = "North"
    DirectionSouth Direction = "South"
    DirectionEast  Direction = "East"
    DirectionWest  Direction = "West"
)

type Fireplace struct {
    BaseObject
    direction            Direction
    isAtOriginalPosition bool
    isFound              bool
}

func (s *Fireplace) TintColor() color.Color {
    return color.White
}

func (s *Fireplace) Name() string {
    return "a Fireplace"
}

func (s *Fireplace) Searched() (bool, []string) {
    if !s.isFound && s.hasHiddenMechanism() {
        s.isFound = true
        return true, []string{
            "You discover a small mechanism on the side of the fireplace.",
        }
    }
    return false, []string{
        "Nothing interesting.",
    }
}

func (s *Fireplace) ToRecordAndType() (recfile.Record, string) {
    return recfile.Record{
        {Name: "name", Value: s.name},
        {Name: "icon", Value: recfile.Int32Str(s.icon)},
        {Name: "pos", Value: s.Pos().Encode()},
        {Name: "isHidden", Value: recfile.BoolStr(s.isHidden)},
        {Name: "isFound", Value: recfile.BoolStr(s.isFound)},
        {Name: "direction", Value: string(s.direction)},
        {Name: "isAtOriginalPosition", Value: recfile.BoolStr(s.isAtOriginalPosition)},
    }, "Fireplace"
}

func NewFireplaceFromRecord(record recfile.Record) *Fireplace {
    fireplace := NewFireplace(DirectionNone)
    for _, field := range record {
        switch field.Name {
        case "name":
            fireplace.name = field.Value
        case "icon":
            fireplace.icon = field.AsInt32()
        case "pos":
            fireplace.SetPos(geometry.MustDecodePoint(field.Value))
        case "isHidden":
            fireplace.isHidden = field.AsBool()
        case "isFound":
            fireplace.isFound = field.AsBool()
        case "direction":
            fireplace.direction = Direction(field.Value)
        case "isAtOriginalPosition":
            fireplace.isAtOriginalPosition = field.AsBool()
        }
    }
    return fireplace
}

func NewFireplace(moveInDirection Direction) *Fireplace {
    icon := int32(241)
    return &Fireplace{
        BaseObject: BaseObject{
            icon:       icon,
            GameObject: GameObject{},
        },
        direction:            moveInDirection,
        isAtOriginalPosition: true,
        isFound:              false,
    }
}

func (s *Fireplace) Description() []string {
    return []string{
        "A fireplace. It's warm and cozy.",
    }
}

func (s *Fireplace) Icon(tick uint64) int32 {
    icon := s.icon
    frameOffset := util.GetLoopingFrameFromTick(tick, 0.2, 2)
    if s.hasHiddenMechanism() && s.isFound {
        icon = 243
    }

    return icon + frameOffset
}

func (s *Fireplace) hasHiddenMechanism() bool {
    return s.direction != DirectionNone
}
func (s *Fireplace) IsWalkable(person *Actor) bool {
    return false
}

func (s *Fireplace) IsTransparent() bool {
    return false
}

func (s *Fireplace) GetContextActions(engine Engine) []util.MenuItem {
    actions := s.BaseObject.GetContextActions(engine, s)
    if s.hasHiddenMechanism() && !s.isHidden {
        actions = append(actions, util.MenuItem{
            Text: "Push plate",
            Action: func() {
                if !s.hasHiddenMechanism() || !s.isFound {
                    return
                }
                if s.isAtOriginalPosition {
                    s.moveAway(engine)
                } else {
                    s.moveBack(engine)
                }
            },
        })
    }
    return actions
}

func (s *Fireplace) moveAway(engine Engine) {
    moveOffset := s.direction.ToPoint()
    newPos := s.Pos().Add(moveOffset)
    currentMap := engine.GetGridMap()
    currentMap.MoveObject(s, newPos)
    s.isAtOriginalPosition = false
}

func (s *Fireplace) moveBack(engine Engine) {
    moveOffset := s.direction.ToPoint().Mul(-1)
    newPos := s.Pos().Add(moveOffset)
    currentMap := engine.GetGridMap()
    if !currentMap.IsCurrentlyPassable(newPos) {
        return
    }

    currentMap.MoveObject(s, newPos)
    s.isAtOriginalPosition = true
}
