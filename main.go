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
    "math/rand"
    "os"
    "runtime/pprof"
)

type GridEngine struct {
    // Basic Game Engine
    wantsToQuit bool
    WorldTicks  uint64

    // Game State & Bookkeeping
    avatar           *game.Actor
    splitControlled  *game.Actor
    playerParty      *game.Party
    playerKnowledge  *game.PlayerKnowledge
    flags            *game.Flags
    currentEncounter game.Encounter
    mapsInMemory     map[string]*gridmap.GridMap[*game.Actor, game.Item, game.Object]

    // combat
    combatManager *CombatState

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
    lastShownText      []string

    // Textures
    worldTiles  *ebiten.Image
    entityTiles *ebiten.Image
    uiTiles     *ebiten.Image

    textToPrint   string
    ticksForPrint int
}

func (g *GridEngine) StartCombat(opponents *game.Actor) {
    g.combatManager.PlayerStartsCombat(g.GetAvatar(), opponents)
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

func (g *GridEngine) Reset() {
    g.WorldTicks = 0
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

func (g *GridEngine) MapToScreenCoordinates(pos geometry.Point) geometry.Point {
    return g.mapWindow.GetScreenGridPositionFromMapGridPosition(pos)
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
    g.RemoveItem(potion)

    potion.SetEmpty()

    g.Print(fmt.Sprintf("%s drank \"%s\"", member.Name(), potion.Name()))
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
func (g *GridEngine) growBloodAt(pos geometry.Point) {
    currentCell := g.currentMap.GetCell(pos)
    currentTile := currentCell.TileType
    if currentTile.Special == gridmap.SpecialTileNone {
        bloodTile := currentTile.WithIcon(104)
        g.currentMap.SetTile(pos, bloodTile)
    }
}

func (g *GridEngine) chooseTarget(onTargetChose func(target geometry.Point)) {
    // TODO
}

func (g *GridEngine) TryPickpocketItem(item game.Item, victim *game.Actor) {
    g.flags.IncrementFlag("pickpocket_attempts")
    chanceOfSuccess := 1.0
    if rand.Float64() < chanceOfSuccess {
        g.PickPocketItem(item, victim)
        g.Print(fmt.Sprintf("You stole \"%s\"", item.Name()))
        g.flags.IncrementFlag("pickpocket_successes")
    } else {
        g.Print(fmt.Sprintf("You were caught stealing \"%s\"", item.Name()))
        // TODO: consequences, encounter depending on the attitude of the victim
        // and the reputation of the party
    }
}

func (g *GridEngine) createPotions(level int) []game.Item {
    var potions []game.Item
    for i := 0; i < level; i++ {
        potions = append(potions, game.NewPotion())
    }
    return potions
}

func (g *GridEngine) createArmor(level, amount int) []game.Item {
    var armor []game.Item
    for i := 0; i < amount; i++ {
        armor = append(armor, game.NewRandomArmor(level))
    }
    return armor
}

func (g *GridEngine) EquipArmor(actor *game.Actor, a *game.Armor) {
    g.playerParty.RemoveItem(a)
    actor.Equip(a)
    g.playerParty.AddItem(a)
}

func (g *GridEngine) RunAnimationScript(script func(exe *gocoro.Execution)) error {
    return g.animationRoutine.Run(script)
}

func (g *GridEngine) actorDied(actor *game.Actor) {
    g.growBloodAt(actor.Pos())
    g.dropActorInventory(actor)
    g.currentMap.SetActorToDowned(actor)
}

func (g *GridEngine) dropActorInventory(actor *game.Actor) {
    chest := game.NewFixedChest(actor.DropInventory())
    dest := actor.Pos()

    freeNeighbor := g.currentMap.GetFreeCellsForDistribution(actor.Pos(), 1, func(p geometry.Point) bool {
        return g.currentMap.Contains(p) && g.currentMap.IsCurrentlyPassable(p) && !g.currentMap.IsDownedActorAt(p) && !g.currentMap.IsObjectAt(p) && !g.currentMap.IsItemAt(p)
    })
    if len(freeNeighbor) == 0 {
        g.currentMap.AddObject(chest, dest)
    } else {
        g.currentMap.AddObject(chest, freeNeighbor[0])
    }
}

func (g *GridEngine) showGameOver() {
    g.ShowText([]string{"Game Over"})
}

func secondsToTicks(seconds float64) int {
    return int(ebiten.ActualTPS() * seconds)
}
