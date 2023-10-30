package game

import (
    "Legacy/renderer"
    "image/color"
    "math/rand"
)

type Door struct {
    BaseObject
    key               string
    isLocked          bool
    isMagicallyLocked bool
    lockStrength      float64
    frameStrength     float64
    isBroken          bool
    listenText        []string
    knockEvent        string
    breakEvent        string
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
        key:          key,
        lockStrength: strength,
        isLocked:     true,
    }
}
func NewMagicallyLockedDoor(strength float64) *Door {
    return &Door{
        BaseObject: BaseObject{
            icon: 185,
        },
        lockStrength:      strength,
        isLocked:          true,
        isMagicallyLocked: true,
    }
}

func (d *Door) Description() []string {
    if d.isBroken {
        return []string{
            "You see a broken door.",
        }
    }
    if d.isLocked {
        if d.isMagicallyLocked {
            return []string{
                "You see a door.",
                "It has a blue shimmer.",
                "It's locked tight.",
            }
        }
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

func (d *Door) Icon(uint64) int {
    if d.isBroken {
        return 192
    }
    if d.isMagicallyLocked {
        return 185
    }
    if d.isLocked {
        return 184
    }
    return 183
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
    if !d.isBroken {
        actions = append(actions, renderer.MenuItem{
            Text: "Listen",
            Action: func() {
                if !d.isBroken {
                    if d.listenText != nil && len(d.listenText) > 0 {
                        engine.ShowColoredText(d.listenText, color.White, true)
                    } else {
                        engine.Print("You hear nothing.")
                    }
                }
            },
        })
        actions = append(actions, renderer.MenuItem{
            Text: "Knock",
            Action: func() {
                if !d.isBroken {
                    if d.knockEvent != "" {
                        engine.TriggerEvent(d.knockEvent)
                    } else {
                        engine.Print("Knocking yields no response.")
                    }
                }
            },
        })
    }
    if d.isLocked && !d.isMagicallyLocked && !d.isBroken {
        actions = append(actions, renderer.MenuItem{
            Text: "Break",
            Action: func() {
                if d.isLocked && !d.isMagicallyLocked && !d.isBroken {
                    if rand.Float64() > d.frameStrength {
                        d.isBroken = true
                        d.isLocked = false
                        engine.DamageAvatar(8)
                        engine.Print("You broke the door.")
                        if d.breakEvent != "" {
                            engine.TriggerEvent(d.breakEvent)
                        }
                    } else {
                        engine.DamageAvatar(10)
                        engine.Print("You failed to break the door.")
                    }
                }
            },
        })
        if engine.PartyHasKey(d.key) {
            actions = append(actions, renderer.MenuItem{
                Text: "Unlock",
                Action: func() {
                    if d.isLocked && !d.isMagicallyLocked && !d.isBroken && engine.PartyHasKey(d.key) {
                        d.isLocked = false
                        engine.Print("You unlocked the door.")
                    }
                },
            })
        } else if engine.PartyHasLockpick() {
            actions = append(actions, renderer.MenuItem{
                Text: "Pick lock",
                Action: func() {
                    if d.isLocked && !d.isMagicallyLocked && !d.isBroken && engine.PartyHasLockpick() {
                        if rand.Float64() > d.lockStrength {
                            d.isLocked = false
                            engine.Print("You picked the lock.")
                        } else {
                            // broke
                            engine.RemoveLockpick()
                            engine.Print("Your lockpick broke.")
                        }
                    }
                },
            })
        }
    }
    return actions
}

func (d *Door) SetListenText(text []string) {
    d.listenText = text
}

func (d *Door) SetBreakEvent(event string) {
    d.breakEvent = event
}
