package game

import (
    "Legacy/geometry"
    "Legacy/gridmap"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
    "math/rand"
)

type GameEvent interface {
    Update()
    IsOver() bool
}

func CreateGameEvent(engine Engine, name string) GameEvent {
    switch name {
    case "break_tauci_cellar_door":
        return NewEncounterBreakTauciCellarDoor(engine)
    case "grow_grass_event":
        return NewFlashWordEvent(engine, "USE MAGIC")
    case "early_exit":
        return NewEarlyExitEvent(engine)
    case "picked_up_nom_de_plume":
        return NewPickedUpNomDePlumeEvent(engine)
    case "imprisonment_tauci":
        return NewImprisonmentTauciEvent(engine)
    case "tauci_rat_king_knocking":
        // unlock the and show some text..
        engine.UnlockDoorsByKeyName("rat_king")
        engine.ShowScrollableText([]string{"Whatever screeching voices you heard before, they are gone now. The silence is short-lived, however.", "’Come on in', a high-pitched voice says. 'We have been waiting for you.'", "There is a loud cackle, as the door is unlocked."}, color.White, true)
    case "rat_battle_start":
        return NewTauciRatBattleStartEvent(engine)
    case "tauci_criminal_offense":
        neededNPCName := "tauci_front_guard"
        neededDialogueFile := "tauci_criminal_offense"
        //engine.GetParty().resetToPreviousPositions() // make the player bounce back from the trigger region
        if !tryStartDialogue(engine, neededNPCName, neededDialogueFile) {
            engine.ShowScrollableText([]string{"The people are crying out for justice.", "But the guards are nowhere to be found."}, color.White, true)
        }
    case "wood_orc_prisoner":
        triggerRegion := name
        neededNPCName := "orc_leader"
        neededDialogueFile := "wood_orc_prisoner"
        engine.GetParty().resetToPreviousPositions() // make the player bounce back from the trigger region
        startDialogueAtLocation(engine, neededNPCName, triggerRegion, neededDialogueFile)
    }
    return nil
}

func NewTauciRatBattleStartEvent(engine Engine) GameEvent {
    event := TauciRatBattleStartEvent{
        engine: engine,
    }
    event.Init()
    return &event
}

func NewImprisonmentTauciEvent(engine Engine) GameEvent {
    // just a little stuff to do:
    // gather all items of the party and remove them
    // put them into a chest
    // teleport the party to the prison
    // repair all doors on the prison map
    //engine.CloseConversation()
    party := engine.GetParty()

    // teleport the party to the prison
    engine.TransitionToNamedLocation("Tauci_Prison_Level_1", "prison_spawn")

    partyInventory := party.RemoveAllItems()
    chest := engine.GetChestByInternalName("tauci_prisoner_belongings")
    if chest == nil {
        println("ERR: Chest not found")
        return nil
    }
    chest.AddFixedLoot(flatten(partyInventory))
    chest.ResetLock()

    engine.ResetAllLockedDoorsOnMap("Tauci_Prison_Level_1")

    engine.ShowScrollableText([]string{"You are now in a prison cell of the Tauci empire.", "The guards took all your belongings and threw them into a chest."}, color.White, true)
    return nil
}

func flatten(inventory [][]Item) []Item {
    result := make([]Item, 0)
    for _, v := range inventory {
        result = append(result, v...)
    }
    return result
}
func tryStartDialogue(engine Engine, neededNPCName string, neededDialogueFile string) bool {
    neededNPC := engine.GetActorByInternalName(neededNPCName)
    if neededNPC != nil && neededNPC.IsAlive() {
        dialogueFromFile := engine.GetDialogueFromFile(neededDialogueFile)
        engine.StartConversation(
            neededNPC,
            dialogueFromFile,
        )
        return true
    }
    return false
}
func startDialogueAtLocation(engine Engine, neededNPCName string, triggerRegion string, neededDialogueFile string) {
    neededNPC := engine.GetActorByInternalName(neededNPCName)
    triggerAtLoc, isInRegion := engine.GetGridMap().GetNamedTriggerAt(neededNPC.Pos())
    isAtLocationOfTrigger := isInRegion && triggerAtLoc.Name == triggerRegion
    if neededNPC != nil && neededNPC.IsAlive() && isAtLocationOfTrigger {
        dialogueFromFile := engine.GetDialogueFromFile(neededDialogueFile)
        engine.StartConversation(
            neededNPC,
            dialogueFromFile,
        )
    }
}

type EncounterPickedUpNomDePlume struct {
    engine           Engine
    drawStartPos     geometry.Point
    availableIndices []int
    text             string
    tickCounter      uint64
}

func (e *EncounterPickedUpNomDePlume) Update() {
    e.tickCounter++
    if e.tickCounter >= uint64(ebiten.ActualTPS()) {
        e.tickCounter = 0
        if len(e.availableIndices) > 0 {
            e.drawChar()
        }
    }
}

func (e *EncounterPickedUpNomDePlume) IsOver() bool {
    return len(e.availableIndices) == 0
}

func (e *EncounterPickedUpNomDePlume) drawChar() {
    randomIndex := rand.Intn(len(e.availableIndices))
    index := e.availableIndices[randomIndex]
    e.availableIndices = append(e.availableIndices[:randomIndex], e.availableIndices[randomIndex+1:]...)
    charToDraw := e.text[index]
    drawPos := geometry.Point{X: e.drawStartPos.X + index, Y: e.drawStartPos.Y}
    e.engine.DrawCharInWorld(rune(charToDraw), drawPos)
}

func NewPickedUpNomDePlumeEvent(engine Engine) GameEvent {
    textarea := engine.GetRegion("bedroom_textarea")
    newText := "+++CELADOR+++"
    availableIndices := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
    return &EncounterPickedUpNomDePlume{
        engine:           engine,
        drawStartPos:     textarea.Min,
        availableIndices: availableIndices,
        text:             newText,
    }
}

type EncounterEarlyExit struct {
    engine     Engine
    ticksAlive uint64
    isOver     bool
}

func (e *EncounterEarlyExit) Name() string {
    return "early_exit"
}

func (e *EncounterEarlyExit) start() {
    if e.engine.GetMapName() != "Bed_Room" {
        println("Early exit event started on wrong map")
        return
    }
    avatar := e.engine.GetAvatar()
    doorPos := avatar.Pos()
    e.engine.RemoveDoorAt(doorPos)
    e.engine.PlayerMovement(geometry.Point{X: -1, Y: 0})
    e.engine.SetWallAt(doorPos)
}

func (e *EncounterEarlyExit) Update() {
    e.ticksAlive++
    if e.ticksAlive > uint64(ebiten.ActualTPS()*1) && !e.isOver {
        e.engine.ShowScrollableText([]string{"That's definitely not how physics worked yesterday.", "Is this a bad dream?"}, color.White, true)
        e.isOver = true
    }
}

func (e *EncounterEarlyExit) IsOver() bool {
    return e.isOver
}

func NewEarlyExitEvent(engine Engine) GameEvent {
    e := &EncounterEarlyExit{
        engine: engine,
    }
    e.start()
    return e
}

type FlashWordEvent struct {
    engine         Engine
    tilesToRestore map[geometry.Point]int32
    tickCounter    uint64
    gridMap        *gridmap.GridMap[*Actor, Item, Object]
    isOver         bool
}

func (e *FlashWordEvent) Update() {
    if e.isOver {
        return
    }
    e.tickCounter++
    if e.tickCounter >= uint64(ebiten.ActualTPS()/2.0) {
        e.tickCounter = 0
        e.restore()
        e.isOver = true
    }
}

func (e *FlashWordEvent) IsOver() bool {
    return e.isOver
}

func (e *FlashWordEvent) overwrite(word string, visibleMapTiles geometry.Rect) {
    randomX := rand.Intn(visibleMapTiles.Size().X - len(word))
    randomY := rand.Intn(visibleMapTiles.Size().Y)

    for i, c := range []rune(word) {
        pos := geometry.Point{X: visibleMapTiles.Min.X + randomX + i, Y: visibleMapTiles.Min.Y + randomY}
        oldIcon := e.gridMap.GetTileIconAt(pos)
        e.tilesToRestore[pos] = oldIcon
        e.engine.DrawCharInWorld(c, pos)
    }
}

func (e *FlashWordEvent) restore() {
    for pos, icon := range e.tilesToRestore {
        e.gridMap.SetTileIcon(pos, icon)
    }
}

func NewFlashWordEvent(engine Engine, word string) GameEvent {
    visibleMapTiles := engine.GetVisibleMap()
    e := &FlashWordEvent{
        engine:         engine,
        gridMap:        engine.GetGridMap(),
        tilesToRestore: make(map[geometry.Point]int32),
    }
    e.overwrite(word, visibleMapTiles)
    return e
}

func NewEncounterBreakTauciCellarDoor(engine Engine) GameEvent {
    return &EncounterBreakTauciCellarDoor{
        engine:        engine,
        wasOnMap:      engine.GetMapName(),
        tickStartedOn: engine.CurrentTick(),
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
    frontGuard := e.engine.GetActorByInternalName("tauci_front_guard")
    if frontGuard == nil || !frontGuard.IsAlive() {
        println("Front guard not found")
        return
    }
    dialogue := e.engine.GetDialogueFromFile("tauci_cellar_door_break")
    e.engine.StartConversation(frontGuard, dialogue)
}
