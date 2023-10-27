package main

import (
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/gocoro"
    "Legacy/gridmap"
    "Legacy/ldtk_go"
    "Legacy/renderer"
    "errors"
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
    _ "image/png"
    "log"
    "math"
    "os"
    "runtime/pprof"
    "strconv"
)

type Modal interface {
    Draw(screen *ebiten.Image)
    ActionUp()
    ActionDown()
    ActionConfirm()
    ShouldClose() bool
}

type UIWidget interface {
    Draw(screen *ebiten.Image)
    ActionUp()
    ActionDown()
    ActionConfirm()
    OnMouseClicked(x int, y int)
    OnMouseMoved(x int, y int)
    ShouldClose() bool
}

type GridEngine struct {
    // Basic Game Engine
    wantsToQuit bool
    WorldTicks  uint64

    // Game State & Bookkeeping
    avatar          *game.Actor
    splitControlled *game.Actor
    playerParty     *game.Party
    playerKnowledge *game.PlayerKnowledge
    flags           *game.Flags
    mapsInMemory    map[string]*gridmap.GridMap[*game.Actor, game.Item, game.Object]

    // Map
    currentMap     *gridmap.GridMap[*game.Actor, game.Item, game.Object]
    ldtkMapProject *ldtk_go.Project
    spawnPosition  geometry.Point

    // Animation
    animationRoutine gocoro.Coroutine

    // UI
    deviceDPIScale     float64
    tileScale          float64
    internalWidth      int
    internalHeight     int
    modalElement       Modal
    inputElement       UIWidget
    uiOverlay          map[int]int
    gridRenderer       *renderer.DualGridRenderer
    mapRenderer        *renderer.MapRenderer
    mapWindow          *renderer.MapWindow
    lastMousePosX      int
    lastMousePosY      int
    contextActions     []renderer.MenuItem
    lastSelectedAction func()

    // Textures
    worldTiles  *ebiten.Image
    entityTiles *ebiten.Image
    uiTiles     *ebiten.Image

    textToPrint   string
    ticksForPrint int
}

func (g *GridEngine) DamageAvatar(amount int) {
    g.GetAvatar().Damage(amount)
    // TODO: visual indicator?
}

func (g *GridEngine) AddLockpicks(amount int) {
    g.playerParty.AddLockpicks(amount)
}

func (g *GridEngine) PartyHasLockpick() bool {
    return g.playerParty.GetLockpicks() > 0
}

func (g *GridEngine) RemoveLockpick() {
    g.playerParty.RemoveLockpicks(1)
}

func (g *GridEngine) ManaSpent(caster *game.Actor, cost int) {
    caster.RemoveMana(cost)
    // TODO: remove HP from mother nature here..
    g.ManaSpentInWorld(caster.Pos(), cost)
}

func (g *GridEngine) ShowDrinkPotionMenu(potion *game.Potion) {
    var menuItems []renderer.MenuItem
    for _, m := range g.playerParty.GetMembers() {
        member := m
        menuItems = append(menuItems, renderer.MenuItem{
            Text: fmt.Sprintf("%s", member.Name()),
            Action: func() {
                if !potion.IsEmpty() {
                    g.DrinkPotion(potion, member)
                }

            },
        })
    }
    g.openMenu(menuItems)
}

func (g *GridEngine) Flags() *game.Flags {
    return g.flags
}

func (g *GridEngine) IsPlayerControlled(holder game.ItemHolder) bool {
    for _, member := range g.playerParty.GetMembers() {
        if member == holder {
            return true
        }
    }
    return g.playerParty == holder
}

func (g *GridEngine) Update() error {
    if g.wantsToQuit {
        return ebiten.Termination
    }

    if g.ticksForPrint > 0 {
        g.ticksForPrint--
    }
    g.handleInput()

    if g.animationRoutine.Running() {
        g.animationRoutine.Update()
    }

    g.WorldTicks++

    return nil
}
func (g *GridEngine) RunAnimationScript(script func(exe *gocoro.Execution)) error {
    return g.animationRoutine.Run(script)
}
func (g *GridEngine) onMove(direction geometry.Point) {
    g.flags.IncrementFlag("steps_taken")

    if g.splitControlled != nil {
        g.currentMap.MoveActor(g.splitControlled, g.splitControlled.Pos().Add(direction))
        g.onAvatarMovedAlone()
    } else {
        g.playerParty.Move(direction)
        g.onPartyMoved()
    }

    if g.flags.GetFlag("steps_taken") == 1 {
        g.onVeryFirstStep()
    }
}
func (g *GridEngine) Draw(screen *ebiten.Image) {
    g.drawUIOverlay(screen)
    g.mapRenderer.Draw(g.playerParty.GetFoV(), screen, g.CurrentTick())

    if g.ticksForPrint > 0 {
        g.drawPrintMessage(screen)
    } else {
        g.drawStatusBar(screen)
    }

    //g.textRenderer.Draw(screen)
    if g.modalElement != nil {
        if g.modalElement.ShouldClose() {
            g.modalElement = nil
            g.updateContextActions()
        } else {
            g.modalElement.Draw(screen)
        }
    }
    if g.inputElement != nil {
        if g.inputElement.ShouldClose() {
            if gridMenu, ok := g.inputElement.(*renderer.GridMenu); ok {
                lastAction := gridMenu.GetLastAction()
                g.lastSelectedAction = lastAction
            }
            g.inputElement = nil
            g.updateContextActions()
        } else {
            g.inputElement.Draw(screen)
        }
    }
}

func main() {
    // Create a CPU profile file
    cpuProfileFile, err := os.Create("cpu.prof")
    if err != nil {
        panic(err)
    }
    defer cpuProfileFile.Close()

    // Start CPU profiling
    if err := pprof.StartCPUProfile(cpuProfileFile); err != nil {
        panic(err)
    }
    defer pprof.StopCPUProfile()

    gameTitle := "Legacy"
    internalScreenWidth, internalScreenHeight := 320, 200 // fixed render Size for this project
    tileScaleFactor := 2.0
    deviceScale := ebiten.DeviceScaleFactor()
    totalScale := tileScaleFactor * deviceScale
    //scaleToFullscreen := false

    scaledScreenWidth := int(math.Floor(float64(internalScreenWidth) * totalScale))
    scaledScreenHeight := int(math.Floor(float64(internalScreenHeight) * totalScale))

    gridEngine := &GridEngine{
        tileScale:        tileScaleFactor,
        internalWidth:    internalScreenWidth,
        internalHeight:   internalScreenHeight,
        worldTiles:       ebiten.NewImageFromImage(mustLoadImage("assets/MergedWorld.png")),
        entityTiles:      ebiten.NewImageFromImage(mustLoadImage("assets/entities.png")),
        uiTiles:          ebiten.NewImageFromImage(mustLoadImage("assets/charset-out.png")),
        uiOverlay:        make(map[int]int),
        mapsInMemory:     make(map[string]*gridmap.GridMap[*game.Actor, game.Item, game.Object]),
        animationRoutine: gocoro.NewCoroutine(),
    }
    ebiten.SetWindowTitle(gameTitle)
    ebiten.SetWindowSize(scaledScreenWidth, scaledScreenHeight)
    ebiten.SetScreenClearedEveryFrame(true)
    gridEngine.Init()
    if err := ebiten.RunGameWithOptions(gridEngine, &ebiten.RunGameOptions{
        GraphicsLibrary: ebiten.GraphicsLibraryOpenGL,
    }); err != nil && !errors.Is(err, ebiten.Termination) {
        log.Fatal(err)
    }
}

func (g *GridEngine) CurrentTick() uint64 {
    return g.WorldTicks
}

func (g *GridEngine) Reset() {
    g.WorldTicks = 0
}

func (g *GridEngine) drawStatusBar(screen *ebiten.Image) {
    status := g.playerParty.Status()
    screenSize := g.gridRenderer.GetSmallGridScreenSize()
    y := screenSize.Y - 1
    x := 0
    //divider := '█'
    for i, charStatus := range status {
        x = i * 10
        //g.gridRenderer.DrawOnSmallGrid(screen, x, y, int(g.fontIndex[charStatus.HealthIcon]))
        g.gridRenderer.DrawColoredString(screen, x+1, y, charStatus.Name, charStatus.StatusColor)
        //g.gridRenderer.DrawOnSmallGrid(screen, x+9, y, int(g.fontIndex[divider]))
    }

    foodCount := g.playerParty.GetFood()
    goldCount := g.playerParty.GetGold()
    lockpickCount := g.playerParty.GetLockpicks()

    foodString := strconv.Itoa(foodCount)
    goldString := strconv.Itoa(goldCount)
    lockpickString := strconv.Itoa(lockpickCount)

    foodIcon := 131
    goldIcon := 132
    lockpickIcon := 133

    yPos := screenSize.Y - 2
    xPosFood := 2
    xPosLockpick := xPosFood + len(foodString) + 2
    xPosGold := screenSize.X - 2 - len(goldString)

    g.gridRenderer.DrawColoredString(screen, xPosFood, yPos, foodString, color.White)
    g.gridRenderer.DrawColoredString(screen, xPosLockpick, yPos, lockpickString, color.White)
    g.gridRenderer.DrawColoredString(screen, xPosGold, yPos, goldString, color.White)

    g.gridRenderer.DrawOnSmallGrid(screen, xPosFood-1, yPos, foodIcon)
    g.gridRenderer.DrawOnSmallGrid(screen, xPosLockpick-1, yPos, lockpickIcon)
    g.gridRenderer.DrawOnSmallGrid(screen, xPosGold+len(goldString), yPos, goldIcon)
}

func (g *GridEngine) QuitGame() {
    g.wantsToQuit = true
}

// onMouseMoved receives the coordinates as character cells
func (g *GridEngine) onMouseMoved(x int, y int) {
    if g.inputElement != nil {
        g.inputElement.OnMouseMoved(x, y)
    }
}

// onMouseClick receives the coordinates as character cells
func (g *GridEngine) onMouseClick(x int, y int) {
    screenSize := g.gridRenderer.GetSmallGridScreenSize()
    oneFourth := screenSize.X / 4

    // if it's the last line, we want to open ui
    if y == screenSize.Y-1 {
        // each 1/4 of the screen is a different UI
        if x < oneFourth {
            g.openCharDetails(0)
        } else if x < oneFourth*2 {
            g.openCharDetails(1)
        } else if x < oneFourth*3 {
            g.openCharDetails(2)
        } else {
            g.openCharDetails(3)
        }
    } else if g.inputElement != nil {
        g.inputElement.OnMouseClicked(x, y)
    } else if g.modalElement != nil {
        g.modalElement.ActionConfirm()
    }
}

func (g *GridEngine) openMenu(items []renderer.MenuItem) {
    gridMenu := renderer.NewGridMenu(g.gridRenderer, items)
    gridMenu.SetAutoClose()
    g.inputElement = gridMenu
    g.inputElement.OnMouseMoved(g.lastMousePosX, g.lastMousePosY)
}

func (g *GridEngine) updateContextActions() {

    // NOTE: We need to reverse the dependency here
    // The objects in the world, should provide us with context actions.
    // We should not know about them beforehand.

    loc := g.GetAvatar().Pos()

    g.contextActions = make([]renderer.MenuItem, 0)

    neighborsWithStuff := g.currentMap.NeighborsCardinal(loc, func(p geometry.Point) bool {
        return g.currentMap.Contains(p) && (g.currentMap.IsActorAt(p) || g.currentMap.IsItemAt(p) || g.currentMap.IsObjectAt(p))
    })

    neighborsWithStuff = append(neighborsWithStuff, loc)

    var actorsNearby []*game.Actor
    var uniqueItemsNearby []game.Item
    var allItemsNearby []game.Item
    var objectsNearby []game.Object

    for _, neighbor := range neighborsWithStuff {
        if g.currentMap.IsActorAt(neighbor) {
            actor := g.currentMap.GetActor(neighbor)
            if !actor.IsHidden() {
                actorsNearby = append(actorsNearby, actor)
            }
        }
        if g.currentMap.IsItemAt(neighbor) {
            item := g.currentMap.ItemAt(neighbor)
            if !item.IsHidden() {
                allItemsNearby = append(allItemsNearby, item)
                if len(uniqueItemsNearby) == 0 {
                    uniqueItemsNearby = append(uniqueItemsNearby, item)
                } else { // can we stack it?
                    for _, otherItem := range uniqueItemsNearby {
                        if !otherItem.CanStackWith(item) {
                            uniqueItemsNearby = append(uniqueItemsNearby, item)
                        }
                    }
                }
            }
        }
        if g.currentMap.IsObjectAt(neighbor) {
            object := g.currentMap.ObjectAt(neighbor)
            if !object.IsHidden() {
                objectsNearby = append(objectsNearby, object)
            }
        }
    }

    for _, actor := range actorsNearby {
        g.contextActions = append(g.contextActions, actor.GetContextActions(g)...)
    }

    if len(allItemsNearby) > 1 {
        g.contextActions = append(g.contextActions, renderer.MenuItem{
            Text: "Pick up all",
            Action: func() {
                for _, item := range allItemsNearby {
                    g.PickUpItem(item)
                }
            },
        })
    }
    for _, item := range uniqueItemsNearby {
        g.contextActions = append(g.contextActions, item.GetContextActions(g)...)
    }

    for _, object := range objectsNearby {
        g.contextActions = append(g.contextActions, object.GetContextActions(g)...)
    }

    // extended range for talking to NPCs
    twoRangeCardinalRelative := []geometry.Point{
        {X: 0, Y: -2},
        {X: 0, Y: 2},
        {X: -2, Y: 0},
        {X: 2, Y: 0},
    }

    for _, relative := range twoRangeCardinalRelative {
        neighbor := loc.Add(relative)
        if g.currentMap.Contains(neighbor) && g.currentMap.IsActorAt(neighbor) {
            actor := g.currentMap.GetActor(neighbor)
            g.contextActions = append(g.contextActions, actor.GetContextActions(g)...)
        }
    }
}

func (g *GridEngine) openContextMenu() {
    if len(g.contextActions) == 0 {
        return
    }
    if len(g.contextActions) == 1 {
        g.contextActions[0].Action()
        return
    }
    g.openMenu(g.contextActions)
}
func (g *GridEngine) openDebugMenu() {
    g.openMenu([]renderer.MenuItem{
        {
            Text: "Toggle NoClip",
            Action: func() {
                noClipActive := g.currentMap.ToggleNoClip()
                g.Print(fmt.Sprintf("DEBUG(No Clip): %t", noClipActive))
            },
        },
        {
            Text: "impulse 9",
            Action: func() {
                g.playerParty.AddFood(100)
                g.playerParty.AddGold(100)
                g.playerParty.AddLockpicks(100)
                g.Print("DEBUG(impulse 9): Added 100 food, gold and lockpicks")
            },
        },
        {
            Text: "Set some flags",
            Action: func() {
                g.flags.SetFlag("can_talk_to_ghosts", 1)
                g.Print("DEBUG(Flags): Set flag \"can_talk_to_ghosts\"")
            },
        },
        {
            Text: "Set Avatar Name to 'Nova'",
            Action: func() {
                g.avatar.SetName("Nova")
                g.Print("DEBUG(Avatar): Set name to 'Nova'")
            },
        },
        {
            Text: "Show all Flags",
            Action: func() {
                g.ShowScrollableText(g.flags.GetDebugInfo(), color.White)
            },
        },
    })
}
func (g *GridEngine) onPartyMoved() {
    g.mapWindow.CenterOn(g.playerParty.Pos())
    g.updateContextActions()
    g.currentMap.UpdateFieldOfView(g.playerParty.GetFoV(), g.playerParty.Pos())
}

func (g *GridEngine) onAvatarMovedAlone() {
    g.mapWindow.CenterOn(g.GetAvatar().Pos())
    g.updateContextActions()
    g.currentMap.UpdateFieldOfView(g.playerParty.GetFoV(), g.GetAvatar().Pos())
}

func (g *GridEngine) GetTextFile(filename string) []string {
    return readLines(filename)
}

func (g *GridEngine) GetAvatar() *game.Actor {
    if g.splitControlled != nil {
        return g.splitControlled
    }
    return g.avatar
}

func (g *GridEngine) Print(text string) {
    g.textToPrint = text
    g.ticksForPrint = secondsToTicks(2)
}

func (g *GridEngine) drawPrintMessage(screen *ebiten.Image) {
    screenSize := g.gridRenderer.GetSmallGridScreenSize()
    yPos := screenSize.Y - 1
    width := screenSize.X
    textLen := len(g.textToPrint)
    xOffsetForCenter := (width - textLen) / 2

    g.gridRenderer.DrawColoredString(screen, xOffsetForCenter, yPos, g.textToPrint, color.White)
}

func (g *GridEngine) DrinkPotion(potion *game.Potion, member *game.Actor) {

    member.AddMana(10)
    g.growGrassAt(member.Pos())

    if potion.IsHeld() {
        g.playerParty.RemoveItem(potion)
    } else {
        g.currentMap.RemoveItem(potion)
    }
    potion.SetEmpty()

    g.Print(fmt.Sprintf("%s drank \"%s\"", member.Name(), potion.Name()))
}

func (g *GridEngine) transition(transition gridmap.Transition) {

    currentMapName := g.currentMap.GetName()
    nextMapName := transition.TargetMap

    // remove the party from the current map
    g.RemovePartyFromMap(g.currentMap)

    // save it
    g.mapsInMemory[currentMapName] = g.currentMap

    var nextMap *gridmap.GridMap[*game.Actor, game.Item, game.Object]
    var isInMemory bool
    // check if the next map is already loaded
    if nextMap, isInMemory = g.mapsInMemory[nextMapName]; !isInMemory {
        // if not, load it from ldtk
        nextMap = g.loadMap(transition.TargetMap)
    } else {
        g.initMapWindow(nextMap.MapWidth, nextMap.MapHeight)
    }

    // set the new map
    g.currentMap = nextMap

    // add the party to the new map
    g.PlaceParty(transition.TargetPos)
}

func (g *GridEngine) onVeryFirstStep() {
    err := g.RunAnimationScript(g.animateDimensionGateDisappears)
    if err != nil {
        println(err.Error())
    }
}

func (g *GridEngine) animateDimensionGateDisappears(exe *gocoro.Execution) {
    //gatePos := g.spawnPosition

}

func (g *GridEngine) ManaSpentInWorld(pos geometry.Point, cost int) {
    // turn the tile into grass
    g.growGrassAt(pos)

    g.flags.IncrementFlagBy("damage_to_mother_nature", cost)
}

func (g *GridEngine) growGrassAt(pos geometry.Point) {
    currentCell := g.currentMap.GetCell(pos)
    currentTile := currentCell.TileType
    if currentTile.Special == gridmap.SpecialTileNone {
        grassTile := currentTile.WithIcon(4)
        g.currentMap.SetTile(pos, grassTile)
    }
}

func (g *GridEngine) openSpellMenu() {
    var menuItems []renderer.MenuItem
    for _, s := range g.playerParty.GetSpells() {
        spell := s
        label := fmt.Sprintf("%s (%d)", spell.Name(), spell.ManaCost())
        menuItems = append(menuItems, renderer.MenuItem{
            Text: label,
            Action: func() {
                if spell.IsTargeted() {
                    g.chooseTarget(func(target geometry.Point) {
                        spell.CastOnTarget(g, g.GetAvatar(), target)
                    })
                } else {
                    spell.Cast(g, g.GetAvatar())
                }
            },
        })
    }
    g.openMenu(menuItems)
}

func (g *GridEngine) chooseTarget(onTargetChose func(target geometry.Point)) {
    // TODO
}

func secondsToTicks(seconds float64) int {
    return int(ebiten.ActualTPS() * seconds)
}
