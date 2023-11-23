package game

import (
    "Legacy/geometry"
    "Legacy/recfile"
    "Legacy/util"
    "fmt"
    "image/color"
)

type Door struct {
    BaseObject
    key               string
    isLocked          bool
    isMagicallyLocked bool
    lockStrength      DifficultyLevel
    frameStrength     DifficultyLevel
    spellStrength     DifficultyLevel
    isBroken          bool
    listenText        []string
    knockEvent        string
    breakEvent        string
}

func (d *Door) TintColor() color.Color {
    return color.White
}
func (d *Door) OnActorWalkedOn(actor *Actor) {
    if d.isLocked && actor.HasKey(d.key) {
        actor.UsedKey(d.key)
    }
}
func (d *Door) Name() string {
    return "a door"
}
func (d *Door) ToRecordAndType() (recfile.Record, string) {
    return recfile.Record{
        {Name: "name", Value: d.name},
        {Name: "icon", Value: recfile.Int32Str(d.icon)},
        {Name: "pos", Value: d.Pos().Encode()},
        {Name: "isHidden", Value: recfile.BoolStr(d.isHidden)},
        {Name: "key", Value: d.key},
        {Name: "isLocked", Value: recfile.BoolStr(d.isLocked)},
        {Name: "isMagicallyLocked", Value: recfile.BoolStr(d.isMagicallyLocked)},
        {Name: "lockStrength", Value: d.lockStrength.ToString()},
        {Name: "frameStrength", Value: d.frameStrength.ToString()},
        {Name: "spellStrength", Value: d.spellStrength.ToString()},
        {Name: "isBroken", Value: recfile.BoolStr(d.isBroken)},
        //{Name: "listenText", Value: recfile.StringSliceStr(d.listenText)},
        {Name: "knockEvent", Value: d.knockEvent},
        {Name: "breakEvent", Value: d.breakEvent},
    }, "door"
}

func NewDoorFromRecord(record recfile.Record) *Door {
    door := NewDoor()
    for _, field := range record {
        switch field.Name {
        case "name":
            door.name = field.Value
        case "icon":
            door.icon = field.AsInt32()
        case "pos":
            door.SetPos(geometry.MustDecodePoint(field.Value))
        case "isHidden":
            door.isHidden = field.AsBool()
        case "key":
            door.key = field.Value
        case "isLocked":
            door.isLocked = field.AsBool()
        case "isMagicallyLocked":
            door.isMagicallyLocked = field.AsBool()
        case "lockStrength":
            door.lockStrength = DifficultyLevelFromString(field.Value)
        case "frameStrength":
            door.frameStrength = DifficultyLevelFromString(field.Value)
        case "spellStrength":
            door.spellStrength = DifficultyLevelFromString(field.Value)
        case "isBroken":
            door.isBroken = field.AsBool()
        //case "listenText": //TODO
        //    door.listenText = recfile.DecodeStringSlice(field.Value)
        case "knockEvent":
            door.knockEvent = field.Value
        case "breakEvent":
            door.breakEvent = field.Value
        }
    }
    return door
}

func NewDoor() *Door {
    return &Door{
        BaseObject: BaseObject{
            icon: 183,
        },
    }
}
func NewLockedDoor(key string, lockStrength DifficultyLevel) *Door {
    return &Door{
        BaseObject: BaseObject{
            icon: 184,
        },
        key:          key,
        lockStrength: lockStrength,
        isLocked:     true,
    }
}
func NewMagicallyLockedDoor(spellStrength DifficultyLevel) *Door {
    return &Door{
        BaseObject: BaseObject{
            icon: 185,
        },
        spellStrength:     spellStrength,
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

func (d *Door) Icon(uint64) int32 {
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
    if person == nil {
        return false
    }
    return person.HasKey(d.key)
}

func (d *Door) IsTransparent() bool {
    return !d.isLocked
}

func (d *Door) GetContextActions(engine Engine) []util.MenuItem {
    actions := d.BaseObject.GetContextActions(engine, d)
    party := engine.GetParty()
    if !d.isBroken {
        actions = append(actions, util.MenuItem{
            Text: "Listen",
            Action: func() {
                if !d.isBroken {
                    if d.listenText != nil && len(d.listenText) > 0 {
                        engine.ShowScrollableText(d.listenText, color.White, true)
                    } else {
                        engine.Print("You hear nothing.")
                    }
                }
            },
        })
        actions = append(actions, util.MenuItem{
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
        if !d.isLocked && d.key != "" && party.HasKey(d.key) {
            actions = append(actions, util.MenuItem{
                Text: "Lock",
                Action: func() {
                    if !d.isLocked && d.key != "" && party.HasKey(d.key) {
                        d.isLocked = true
                        party.UsedKey(d.key)
                        engine.Print("You locked the door.")
                    }
                },
            })
        }
        if !d.isLocked && d.key != "" && party.GetLockpicks() > 0 {
            skill := ThievingSkillLockpicking
            difficulty := d.lockStrength
            actions = append(actions, util.MenuItem{
                Text: fmt.Sprintf("Lock (lockpick) - %s", engine.GetRelativeDifficulty(skill, difficulty).ToString()),
                Action: func() {
                    if !d.isLocked && d.key != "" && party.GetLockpicks() > 0 {
                        if engine.SkillCheckAvatar(skill, difficulty) {
                            d.isLocked = true
                            engine.Print("You locked the door.")
                        } else {
                            engine.Print("Your lockpick broke.")
                            engine.RemoveLockpick()
                        }
                    }
                },
            })
        }
    }

    if d.isLocked && !d.isMagicallyLocked && !d.isBroken {
        skill := PhysicalSkillTackle
        difficulty := d.frameStrength
        actions = append(actions, util.MenuItem{
            Text: fmt.Sprintf("Break - %s", engine.GetRelativeDifficulty(skill, difficulty).ToString()),
            Action: func() {
                if d.isLocked && !d.isMagicallyLocked && !d.isBroken {
                    if engine.SkillCheckAvatar(skill, difficulty) {
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
        if breakingTool := engine.GetBreakingToolName(); breakingTool != "" {
            skillForTool := PhysicalSkillTools
            difficultyForTool := d.frameStrength.ReducedBy(1)
            actions = append(actions, util.MenuItem{
                Text: fmt.Sprintf("Break (%s) - %s", breakingTool, engine.GetRelativeDifficulty(skillForTool, difficultyForTool).ToString()),
                Action: func() {
                    if d.isLocked && !d.isMagicallyLocked && !d.isBroken {
                        if engine.SkillCheckAvatar(skillForTool, difficultyForTool) {
                            d.isBroken = true
                            d.isLocked = false
                            engine.Print("You broke the door.")
                            if d.breakEvent != "" {
                                engine.TriggerEvent(d.breakEvent)
                            }
                        } else {
                            engine.Print("You failed to break the door.")
                        }
                    }
                },
            })
        }
        if party.HasKey(d.key) {
            actions = append(actions, util.MenuItem{
                Text: "Unlock",
                Action: func() {
                    if d.isLocked && !d.isMagicallyLocked && !d.isBroken && party.HasKey(d.key) {
                        d.isLocked = false
                        engine.GetParty().UsedKey(d.key)
                        engine.Print("You unlocked the door.")
                    }
                },
            })
        } else if party.GetLockpicks() > 0 {
            skillForPick := ThievingSkillLockpicking
            difficultyForPick := d.lockStrength
            actions = append(actions, util.MenuItem{
                Text: fmt.Sprintf("Pick lock - %s", engine.GetRelativeDifficulty(skillForPick, difficultyForPick).ToString()),
                Action: func() {
                    if d.isLocked && !d.isMagicallyLocked && !d.isBroken && party.GetLockpicks() > 0 {
                        if engine.SkillCheckAvatar(skillForPick, difficultyForPick) {
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

func (d *Door) SetFrameStrength(level DifficultyLevel) {
    d.frameStrength = level
}
