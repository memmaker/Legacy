package main

import (
    "Legacy/dialogue"
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/gridmap"
    "Legacy/ldtk_go"
    "Legacy/renderer"
    "errors"
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/inpututil"
    _ "image/png"
    "log"
    "math"
    "os"
    "path"
    "runtime/pprof"
    "strings"
)

type Modal interface {
    Draw(screen *ebiten.Image)
    ActionUp()
    ActionDown()
    ActionConfirm() bool
}

type UIWidget interface {
    Draw(screen *ebiten.Image)
    ActionUp()
    ActionDown()
    ActionConfirm() bool
    OnMouseClicked(x int, y int) bool
    OnMouseMoved(x int, y int)
}

type GridEngine struct {
    // Basic Game Engine
    wantsToQuit bool
    WorldTicks  uint64

    // Game State & Bookkeeping
    player          *game.Actor
    playerParty     *game.Party
    playerKnowledge *dialogue.PlayerKnowledge

    // Map
    gridmap         *gridmap.GridMap[*game.Actor, game.Item, *game.Object]
    transitionLayer *ldtk_go.Layer
    transitionMap   map[geometry.Point]Transition
    ldtkMapProject  *ldtk_go.Project

    // UI
    deviceDPIScale float64
    tileScale      float64
    internalWidth  int
    internalHeight int
    modalElement   Modal
    inputElement   UIWidget
    uiOverlay      map[int]int
    gridRenderer   *renderer.DualGridRenderer
    mapRenderer    *renderer.MapRenderer
    mapWindow      *renderer.MapWindow
    lastMousePosX  int
    lastMousePosY  int
    contextActions []renderer.MenuItem

    // Textures
    worldTiles  *ebiten.Image
    entityTiles *ebiten.Image
    uiTiles     *ebiten.Image
}

func (g *GridEngine) Update() error {
    if g.wantsToQuit {
        return ebiten.Termination
    }

    g.handleInput()
    // These are global and don't need any focus..
    //g.checkRecorderControls()

    //g.Audio.Update()
    // do we need this?
    //g.Input.Update()

    //g.UserInterface.Update(g.Input)
    g.WorldTicks++
    /*
       if !g.UserInterface.IsBlocking() {
           g.UpdateScheduledCalls()

       }
    */
    /*
       g.UserInterface.Draw(g.Console)

       g.Animator.Update()
       g.Animator.Draw(g.Console)
    */
    return nil
}

func (g *GridEngine) handleInput() {
    if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
        g.gridmap.MoveActor(g.player, g.player.Pos().Add(geometry.Point{X: 1}))
        //g.mapWindow.Scroll(geometry.Point{1, 0})
        g.onPlayerMoved()
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
        g.gridmap.MoveActor(g.player, g.player.Pos().Add(geometry.Point{X: -1}))
        //g.mapWindow.Scroll(geometry.Point{-1, 0})
        g.onPlayerMoved()
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
        if g.inputElement != nil {
            g.inputElement.ActionUp()
        } else if g.modalElement != nil {
            g.modalElement.ActionUp()
        } else {
            g.gridmap.MoveActor(g.player, g.player.Pos().Add(geometry.Point{Y: -1}))
            g.onPlayerMoved()
        }
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
        if g.inputElement != nil {
            g.inputElement.ActionDown()
        } else if g.modalElement != nil {
            g.modalElement.ActionDown()
        } else {
            g.gridmap.MoveActor(g.player, g.player.Pos().Add(geometry.Point{Y: 1}))
            g.onPlayerMoved()
        }
    }

    if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
        if g.inputElement != nil {
            if g.inputElement.ActionConfirm() {
                g.inputElement = nil
            }
        } else if g.modalElement != nil {
            if g.modalElement.ActionConfirm() {
                g.modalElement = nil
            }
        } else if g.gridmap.IsItemAt(g.player.Pos()) {
            item := g.gridmap.ItemAt(g.player.Pos())
            item.Use()
        } else if transition, ok := g.transitionMap[g.player.Pos()]; ok {
            g.loadMap(transition.TargetMap, transition.TargetPos)
        }
    }

    if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
        if g.gridmap.IsItemAt(g.player.Pos()) {
            item := g.gridmap.ItemAt(g.player.Pos())
            g.pickUpItem(item)
        }
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyI) {
        g.openPartyInventory()
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
        if g.inputElement != nil {
            g.inputElement = nil
        }
        if g.modalElement != nil {
            g.modalElement = nil
        }
    }

    if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
        g.openCharDetails(0)
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
        //g.openCharDetails(1)
        g.startConversation(game.NewActor("Tim", geometry.Point{X: 3, Y: 3}, 22))

    } else if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
        //g.openCharDetails(2)
        g.openMenu(geometry.Point{X: 3, Y: 2}, []renderer.MenuItem{
            {Text: "Item 1", Action: func() {
                println("Item 1")
            }},
            {Text: "Item 2", Action: func() {
                println("Item 2")
            }},
        })
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF4) {
        //g.openCharDetails(3)
        g.openContextMenu()
    }
    mousePosX, mousePosY := ebiten.CursorPosition()
    cellX := int(math.Floor(float64(mousePosX) / float64(g.gridRenderer.GetScaledSmallGridSize())))
    cellY := int(math.Floor(float64(mousePosY) / float64(g.gridRenderer.GetScaledSmallGridSize())))

    if cellX != g.lastMousePosX || cellY != g.lastMousePosY {
        g.onMouseMoved(cellX, cellY)
        g.lastMousePosX = cellX
        g.lastMousePosY = cellY
    }
    if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
        g.onMouseClick(cellX, cellY)
    }
}
func (g *GridEngine) Draw(screen *ebiten.Image) {
    g.drawUIOverlay(screen)
    g.mapRenderer.Draw(screen)
    g.drawStatusBar(screen)
    //g.textRenderer.Draw(screen)
    if g.modalElement != nil {
        g.modalElement.Draw(screen)
    }
    if g.inputElement != nil {
        g.inputElement.Draw(screen)
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
        tileScale:      tileScaleFactor,
        internalWidth:  internalScreenWidth,
        internalHeight: internalScreenHeight,
        worldTiles:     ebiten.NewImageFromImage(mustLoadImage("assets/MergedWorld.png")),
        entityTiles:    ebiten.NewImageFromImage(mustLoadImage("assets/entities.png")),
        uiTiles:        ebiten.NewImageFromImage(mustLoadImage("assets/charset-out.png")),
        uiOverlay:      make(map[int]int),
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

func (g *GridEngine) mapLookup(x, y int) (*ebiten.Image, int) {
    location := geometry.Point{X: x, Y: y}

    if g.gridmap.IsActorAt(location) {
        return g.entityTiles, g.gridmap.GetActor(location).Icon()
    }

    if g.gridmap.IsItemAt(location) {
        return g.entityTiles, g.gridmap.ItemAt(location).Icon()
    }

    tile := g.gridmap.GetCell(location)
    return g.worldTiles, tile.TileType.DefinedIcon
}

func (g *GridEngine) openCharDetails(partyIndex int) {
    actor := g.playerParty.GetMember(partyIndex)
    if actor != nil {
        g.showModal(actor.GetDetails())
    }
}

func (g *GridEngine) openPartyInventory() {
    header := []string{"Inventory", "---------"}
    details := g.playerParty.GetInventoryDetails()
    g.showModal(append(header, details...))
}

func (g *GridEngine) showModal(text []string) {
    g.modalElement = renderer.NewScrollableTextWindowWithAutomaticSize(g.gridRenderer, text)
}

func (g *GridEngine) TotalScale() float64 {
    return g.tileScale * g.deviceDPIScale
}
func (g *GridEngine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    panic("should use layoutf")
}

func (g *GridEngine) LayoutF(outsideWidth, outsideHeight float64) (screenWidth, screenHeight float64) {
    g.deviceDPIScale = ebiten.DeviceScaleFactor()
    totalScale := g.tileScale * g.deviceDPIScale
    return float64(g.internalWidth) * totalScale, float64(g.internalHeight) * totalScale
}

func (g *GridEngine) CurrentTick() uint64 {
    return g.WorldTicks
}

func (g *GridEngine) Reset() {
    g.WorldTicks = 0
}

func (g *GridEngine) drawUIOverlay(screen *ebiten.Image) {
    screenW := 40 // in 8x8 cells
    for cellIndex, tileIndex := range g.uiOverlay {
        cellX := cellIndex % screenW
        cellY := cellIndex / screenW
        g.gridRenderer.DrawOnSmallGrid(screen, cellX, cellY, tileIndex)
    }
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
        if g.inputElement.OnMouseClicked(x, y) {
            g.inputElement = nil
        }
    } else if g.modalElement != nil {
        if g.modalElement.ActionConfirm() {
            g.modalElement = nil
        }
    }
}

func (g *GridEngine) pickUpItem(item game.Item) {
    g.playerParty.AddItem(item)
    g.gridmap.RemoveItem(item)
}

func (g *GridEngine) openSpeechWindow(speaker *game.Actor, text []string, onLastPage func()) {
    g.modalElement = renderer.NewMultiPageWindow(g.gridRenderer, 3, speaker.Icon(), text, onLastPage)
}

func (g *GridEngine) openMenu(topLeft geometry.Point, items []renderer.MenuItem) {
    g.inputElement = renderer.NewGridMenu(g.gridRenderer, topLeft, items)
    g.inputElement.OnMouseMoved(g.lastMousePosX, g.lastMousePosY)
}

func (g *GridEngine) openConversationMenu(topLeft geometry.Point, items []renderer.MenuItem) {
    g.inputElement = renderer.NewGridDialogueMenu(g.gridRenderer, topLeft, items)
    g.inputElement.OnMouseMoved(g.lastMousePosX, g.lastMousePosY)
}

func (g *GridEngine) startConversation(npc *game.Actor) {
    // NOTE: Conversations can have a line length of 27 chars
    charName := npc.Name()
    filename := strings.ToLower(charName) + ".txt"
    dialogueFilename := path.Join("assets", "dialogue", filename)
    if !doesFileExist(dialogueFilename) {
        return
    }
    dialogueFile := mustOpen(dialogueFilename)
    loadedDialogue := dialogue.NewDialogueFromFile(dialogueFile)

    options := loadedDialogue.GetOptions(g.playerKnowledge)
    g.openSpeechWindow(npc, npc.LookDescription(), func() {
        g.openConversationMenu(geometry.Point{X: 3, Y: 13}, g.toMenuItems(npc, loadedDialogue, options))
    })
}

func (g *GridEngine) toMenuItems(npc *game.Actor, dialogue *dialogue.Dialogue, options []string) []renderer.MenuItem {
    var items []renderer.MenuItem
    for _, o := range options {
        option := o
        items = append(items, renderer.MenuItem{
            Text: option,
            Action: func() {
                response, endsConversation := dialogue.GetResponseAndAddKnowledge(g.playerKnowledge, option)
                g.inputElement = nil
                if endsConversation {
                    g.modalElement = nil
                } else {
                    g.openSpeechWindow(npc, response, func() {
                        newOptions := dialogue.GetOptions(g.playerKnowledge)
                        g.openConversationMenu(geometry.Point{X: 3, Y: 13}, g.toMenuItems(npc, dialogue, newOptions))
                    })
                }
            },
        })
    }
    return items
}

func (g *GridEngine) updateContextActions() {

    // NOTE: We need to reverse the dependency here
    // The objects in the world, should provide us with context actions.
    // We should not know about them beforehand.

    loc := g.player.Pos()
    g.contextActions = make([]renderer.MenuItem, 0)
    if g.gridmap.IsItemAt(loc) {
        item := g.gridmap.ItemAt(loc)
        g.contextActions = append(g.contextActions, renderer.MenuItem{
            Text: fmt.Sprintf("Pick up \"%s\"", item.ShortDescription()),
            Action: func() {
                g.pickUpItem(item)
            },
        })
    }
    neighborsWithActors := g.gridmap.NeighborsCardinal(loc, func(p geometry.Point) bool {
        return g.gridmap.IsActorAt(p)
    })

    for _, neighbor := range neighborsWithActors {
        actor := g.gridmap.GetActor(neighbor)
        g.contextActions = append(g.contextActions, renderer.MenuItem{
            Text: fmt.Sprintf("Talk to \"%s\"", actor.Name()),
            Action: func() {
                g.startConversation(actor)
            },
        })
    }

}

func (g *GridEngine) openContextMenu() {
    if len(g.contextActions) == 0 {
        return
    }
    g.openMenu(geometry.Point{X: 3, Y: 13}, g.contextActions)
}

func (g *GridEngine) onPlayerMoved() {
    g.mapWindow.CenterOn(g.player.Pos())
    g.updateContextActions()
}
