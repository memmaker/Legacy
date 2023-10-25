package game

import (
    "Legacy/renderer"
    "fmt"
    "image/color"
)

type Door struct {
    BaseObject
    key      string
    isLocked bool
    strength float64
}

func (d *Door) TintColor() color.Color {
    return color.White
}

func (d *Door) Name() string {
    return "a door"
}

func NewDoor() *Door {
    return &Door{
        BaseObject: BaseObject{
            icon: 183,
        },
    }
}
func NewLockedDoor(key string, strength float64) *Door {
    return &Door{
        BaseObject: BaseObject{
            icon: 184,
        },
        key:      key,
        strength: strength,
        isLocked: true,
    }
}

func (d *Door) Description() []string {
    if d.isLocked {
        return []string{
            "You see a door.",
            "It appears to be locked.",
        }
    }
    return []string{
        "You see a door.",
        "It is unlocked.",
    }
}

func (d *Door) Icon() int {
    if !d.isLocked {
        return 183
    }
    return 184
}
func (d *Door) IsWalkable(person *Actor) bool {
    if !d.isLocked {
        return true
    }
    return person.HasKey(d.key)
}

func (d *Door) IsTransparent() bool {
    return !d.isLocked
}

func (d *Door) GetContextActions(engine Engine) []renderer.MenuItem {
    actions := d.BaseObject.GetContextActions(engine, d)
    if d.isLocked && engine.GetAvatar().HasKey(d.key) {
        actions = append(actions, renderer.MenuItem{
            Text: fmt.Sprintf("Unlock \"%s\"", d.Name()),
            Action: func() {
                d.isLocked = false
            },
        })
    }
    return actions
}
