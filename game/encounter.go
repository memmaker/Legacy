package game

import "Legacy/renderer"

type Encounter interface {
    Name() string
    Start()
    Update()
    IsOver() bool
}

func GetEncounter(engine Engine, name string) Encounter {
    switch name {
    case "break_tauci_cellar_door":
        return NewEncounterBreakTauciCellarDoor(engine)
    }
    return nil
}

func NewEncounterBreakTauciCellarDoor(engine Engine) Encounter {
    return &EncounterBreakTauciCellarDoor{
        engine: engine,
    }
}

type EncounterBreakTauciCellarDoor struct {
    engine        Engine
    wasOnMap      string
    tickStartedOn uint64
    isOver        bool
}

func (e *EncounterBreakTauciCellarDoor) IsOver() bool {
    return e.isOver
}

func (e *EncounterBreakTauciCellarDoor) Update() {
    if e.isOver {
        return
    }
    ticksElapsed := e.engine.CurrentTick() - e.tickStartedOn
    tenSecondsElapsed := e.engine.TicksToSeconds(ticksElapsed) > 10
    currentMap := e.engine.GetMapName()
    if tenSecondsElapsed && currentMap == e.wasOnMap {
        println("Ticks elapsed: ", ticksElapsed)
        e.isOver = true
        e.guardsArrive()
    }
}

func (e *EncounterBreakTauciCellarDoor) Name() string {
    return "break_tauci_cellar_door"
}

func (e *EncounterBreakTauciCellarDoor) guardsArrive() {
    text := []string{
        "The guards arrive and ask you what you are doing here.",
    }
    guardIcon := int32(132)
    choices := []renderer.MenuItem{
        {
            Text: "I am a friend of the owner.",
            Action: func() {
                e.isOver = true
                //e.engine.openIconWindow(npc.Icon(0), response.Text, func() { g.modalElement = nil })
            },
        },
        {
            Text: "I am a thief.",
            Action: func() {
                e.isOver = true
            },
        },
    }
    e.engine.ShowMultipleChoiceDialogue(guardIcon, text, choices)
}
func (e *EncounterBreakTauciCellarDoor) Start() {
    e.wasOnMap = e.engine.GetMapName()
    e.tickStartedOn = e.engine.CurrentTick()
}
