package game

import (
    "Legacy/geometry"
    "Legacy/gridmap"
    "Legacy/recfile"
    "Legacy/util"
    "fmt"
    "image/color"
)

type Vehicle struct {
    BaseObject
    name                 string
    canTraverseWater     bool
    canTraverseLand      bool
    canTraverseMountains bool
    minutesPerStep       int
}

func (v *Vehicle) TintColor() color.Color {
    return color.White
}

func (v *Vehicle) Name() string {
    return v.name
}
func (v *Vehicle) ToRecordAndType() (recfile.Record, string) {
    return recfile.Record{
        {Name: "name", Value: v.name},
        {Name: "icon", Value: recfile.Int32Str(v.icon)},
        {Name: "pos", Value: v.Pos().Encode()},
        {Name: "isHidden", Value: recfile.BoolStr(v.isHidden)},
    }, "Vehicle"
}

func NewVehicleFromRecord(record recfile.Record) *Vehicle {
    vehicle := &Vehicle{}
    for _, field := range record {
        switch field.Name {
        case "name":
            vehicle.name = field.Value
        case "icon":
            vehicle.icon = field.AsInt32()
        case "pos":
            vehicle.SetPos(geometry.MustDecodePoint(field.Value))
        case "isHidden":
            vehicle.isHidden = field.AsBool()
        }
    }
    return vehicle
}
func NewBalloon() *Vehicle {
    return &Vehicle{
        BaseObject: BaseObject{
            icon: 6,
        },
        name:                 "Balloon",
        canTraverseMountains: true,
        canTraverseLand:      true,
        canTraverseWater:     true,
        minutesPerStep:       5,
    }
}

func NewShip() *Vehicle {
    return &Vehicle{
        BaseObject: BaseObject{
            icon: 47,
        },
        name:                 "Ship",
        canTraverseMountains: false,
        canTraverseLand:      false,
        canTraverseWater:     true,
        minutesPerStep:       10,
    }
}

func (v *Vehicle) Description() []string {
    return []string{
        "Somebody has built a Vehicle here.",
        "It is called:",
        v.name,
    }
}

func (v *Vehicle) Icon(uint64) int32 {
    return v.icon
}
func (v *Vehicle) IsWalkable(person *Actor) bool {
    return true
}

func (v *Vehicle) IsTransparent() bool {
    return true
}

func (v *Vehicle) GetContextActions(engine Engine) []util.MenuItem {
    actions := v.BaseObject.GetContextActions(engine, v)
    party := engine.GetParty()
    if party.IsInVehicle() {
        actions = append(actions, util.MenuItem{
            Text: fmt.Sprintf("Exit %s", v.name),
            Action: func() {
                v.Exit(party)
            },
        })
    } else {
        actions = append(actions, util.MenuItem{
            Text: fmt.Sprintf("Enter %s", v.name),
            Action: func() {
                v.Enter(party)
            },
        })
    }

    return actions
}

func (v *Vehicle) Enter(party *Party) {
    party.EnterVehicle(v)
}

func (v *Vehicle) Exit(party *Party) {
    party.TryExitVehicle()
}

func (v *Vehicle) GetMinutesPerStep() int {
    return v.minutesPerStep
}

func (v *Vehicle) CanMoveTo(cell gridmap.MapCell[*Actor, Item, Object]) bool {
    if cell.Object != nil {
        return false
    }

    if cell.TileType.IsWater() {
        return v.canTraverseWater
    }

    if cell.TileType.IsMountain() {
        return v.canTraverseMountains
    }

    if cell.TileType.IsLand() {
        return v.canTraverseLand
    }
    return cell.TileType.IsWalkable
}
