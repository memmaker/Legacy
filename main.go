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
    "image/color"
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
    playerKnowledge *dialogue.PlayerKnowledge

    // Map
    gridmap         *gridmap.GridMap[*game.Actor, game.Item, game.Object]
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

    textToPrint   string
    ticksForPrint int
    spawnPosition geometry.Point
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
        g.move(geometry.Point{X: 1, Y: 0})
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
        g.move(geometry.Point{X: -1, Y: 0})
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
        if g.inputElement != nil {
            g.inputElement.ActionUp()
        } else if g.modalElement != nil {
            g.modalElement.ActionUp()
        } else {
            g.move(geometry.Point{X: 0, Y: -1})
        }
    }
    if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
        if g.inputElement != nil {
            g.inputElement.ActionDown()
        } else if g.modalElement != nil {
            g.modalElement.ActionDown()
        } else {
            g.move(geometry.Point{X: 0, Y: 1})
        }
    }

    if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
        if g.inputElement != nil {
            g.inputElement.ActionConfirm()
        } else if g.modalElement != nil {
            g.modalElement.ActionConfirm()
        } else if transition, ok := g.transitionMap[g.avatar.Pos()]; ok {
            g.loadMap(transition.TargetMap)
            g.PlaceParty(transition.TargetPos)
        } else {
            g.openContextMenu()
        }
    }

    if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
        if g.gridmap.IsItemAt(g.avatar.Pos()) {
            item := g.gridmap.ItemAt(g.avatar.Pos())
            g.PickUpItem(item)
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
        g.StartConversation(game.NewActor("Tim", 22))

    } else if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
        //g.openCharDetails(2)
        g.openMenu([]renderer.MenuItem{
            {Text: "Item 1", Action: func() {
                println("Item 1")
            }},
            {Text: "Item 2", Action: func() {
                println("Item 2")
            }},
        })
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF4) {
        g.openCharDetails(3)
    } else if inpututil.IsKeyJustPressed(ebiten.KeyF5) {
        g.openPartyMenu()
    }

    cellX, cellY := g.gridRenderer.ScreenToSmallCell(ebiten.CursorPosition())

    if cellX != g.lastMousePosX || cellY != g.lastMousePosY {
        g.onMouseMoved(cellX, cellY)
        g.lastMousePosX = cellX
        g.lastMousePosY = cellY
    }
    if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
        g.onMouseClick(cellX, cellY)
    }
}

func (g *GridEngine) move(direction geometry.Point) {
    if g.splitControlled != nil {
        g.gridmap.MoveActor(g.splitControlled, g.splitControlled.Pos().Add(direction))
        g.onAvatarMovedAlone()
    } else {
        g.playerParty.Move(direction)
        g.onPartyMoved()
    }
}
func (g *GridEngine) Draw(screen *ebiten.Image) {
    g.drawUIOverlay(screen)
    g.mapRenderer.Draw(g.playerParty.GetFoV(), screen)

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

func (g *GridEngine) mapLookup(x, y int) (*ebiten.Image, int, color.Color) {
    location := geometry.Point{X: x, Y: y}

    if g.gridmap.IsActorAt(location) {
        return g.entityTiles, g.gridmap.GetActor(location).Icon(), color.White
    }

    if g.gridmap.IsItemAt(location) {
        itemAt := g.gridmap.ItemAt(location)
        return g.entityTiles, itemAt.Icon(), itemAt.TintColor()
    }

    if g.gridmap.IsObjectAt(location) {
        objectAt := g.gridmap.ObjectAt(location)
        return g.entityTiles, objectAt.Icon(), objectAt.TintColor()
    }

    tile := g.gridmap.GetCell(location)
    return g.worldTiles, tile.TileType.DefinedIcon, color.White
}

func (g *GridEngine) openCharDetails(partyIndex int) {
    actor := g.playerParty.GetMember(partyIndex)
    if actor != nil {
        g.showModal(actor.GetDetails())
    }
}

func (g *GridEngine) openPartyInventory() {
    //header := []string{"Inventory", "---------"}
    partyInventory := g.playerParty.GetInventory()
    if len(partyInventory) == 0 {
        g.showModal([]string{"Your party has no items."})
        return
    }
    var menuItems []renderer.MenuItem
    for _, i := range partyInventory {
        item := i
        menuItems = append(menuItems, renderer.MenuItem{
            Text: i.Name(),
            Action: func() {
                g.openMenu(item.GetContextActions(g))
            },
        })
    }
    g.openMenu(menuItems)
}

func (g *GridEngine) showModal(text []string) {
    g.modalElement = renderer.NewScrollableTextWindowWithAutomaticSize(g.gridRenderer, text)
}

func (g *GridEngine) ShowScrollableText(text []string, color color.Color) {
    modal := renderer.NewScrollableTextWindowWithAutomaticSize(g.gridRenderer, text)
    modal.SetTextColor(color)
    g.modalElement = modal
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
        g.inputElement.OnMouseClicked(x, y)
    } else if g.modalElement != nil {
        g.modalElement.ActionConfirm()
    }
}

func (g *GridEngine) PickUpItem(item game.Item) {
    g.playerParty.AddItem(item)
    g.gridmap.RemoveItem(item)
    g.Print(fmt.Sprintf("Taken \"%s\"", item.Name()))
}

func (g *GridEngine) DropItem(item game.Item) {
    g.playerParty.RemoveItem(item)
    destPos := g.avatar.Pos()
    if g.TryPlaceItem(item, destPos) {
        g.Print(fmt.Sprintf("Dropped \"%s\"", item.Name()))
    }
}

func (g *GridEngine) TryPlaceItem(item game.Item, destPos geometry.Point) bool {
    if g.gridmap.IsItemAt(destPos) {
        freeCells := g.gridmap.GetFreeCellsForDistribution(g.avatar.Pos(), 1, func(p geometry.Point) bool {
            return g.gridmap.Contains(p) && g.gridmap.IsWalkableFor(p, g.avatar) && !g.gridmap.IsItemAt(p)
        })
        if len(freeCells) > 0 {
            destPos = freeCells[0]
        } else {
            g.Print(fmt.Sprintf("No space to drop \"%s\"", item.Name()))
            return false
        }
    }

    g.gridmap.AddItem(item, destPos)
    return true
}
func (g *GridEngine) openSpeechWindow(speaker *game.Actor, text []string, onLastPage func()) {
    g.modalElement = renderer.NewMultiPageWindow(g.gridRenderer, 3, speaker.Icon(), text, onLastPage)
}

func (g *GridEngine) openMenu(items []renderer.MenuItem) {
    gridMenu := renderer.NewGridMenu(g.gridRenderer, items)
    gridMenu.SetAutoClose()
    g.inputElement = gridMenu
    g.inputElement.OnMouseMoved(g.lastMousePosX, g.lastMousePosY)
}

func (g *GridEngine) openConversationMenu(topLeft geometry.Point, items []renderer.MenuItem) {
    g.inputElement = renderer.NewGridDialogueMenu(g.gridRenderer, topLeft, items)
    g.inputElement.OnMouseMoved(g.lastMousePosX, g.lastMousePosY)
}

func (g *GridEngine) StartConversation(npc *game.Actor) {
    // NOTE: Conversations can have a line length of 27 chars
    charName := npc.Name()
    filename := strings.ToLower(charName) + ".txt"
    dialogueFilename := path.Join("assets", "dialogue", filename)
    if !doesFileExist(dialogueFilename) {
        g.Print(fmt.Sprintf("\"%s\" has nothing to say.", charName))
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
                response, effect := dialogue.GetResponseAndAddKnowledge(g.playerKnowledge, option)
                g.inputElement = nil
                quitsDialogue := g.handleDialogueEffect(npc, effect)
                if quitsDialogue {
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

func (g *GridEngine) handleDialogueEffect(npc *game.Actor, effect string) bool {
    switch effect {
    case "quits":
        return true
    case "joins":
        g.AddToParty(npc)
        return false
    case "sells":
        g.openVendorMenu(npc)
        return true
    }
    return false
}
func (g *GridEngine) updateContextActions() {

    // NOTE: We need to reverse the dependency here
    // The objects in the world, should provide us with context actions.
    // We should not know about them beforehand.

    loc := g.GetAvatar().Pos()

    g.contextActions = make([]renderer.MenuItem, 0)

    neighborsWithStuff := g.gridmap.NeighborsCardinal(loc, func(p geometry.Point) bool {
        return g.gridmap.Contains(p) && (g.gridmap.IsActorAt(p) || g.gridmap.IsItemAt(p) || g.gridmap.IsObjectAt(p))
    })

    neighborsWithStuff = append(neighborsWithStuff, loc)

    for _, neighbor := range neighborsWithStuff {
        if g.gridmap.IsActorAt(neighbor) {
            actor := g.gridmap.GetActor(neighbor)
            g.contextActions = append(g.contextActions, actor.GetContextActions(g)...)
        }
        if g.gridmap.IsItemAt(neighbor) {
            item := g.gridmap.ItemAt(neighbor)
            g.contextActions = append(g.contextActions, item.GetContextActions(g)...)
        }
        if g.gridmap.IsObjectAt(neighbor) {
            object := g.gridmap.ObjectAt(neighbor)
            g.contextActions = append(g.contextActions, object.GetContextActions(g)...)
        }
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
        if g.gridmap.Contains(neighbor) && g.gridmap.IsActorAt(neighbor) {
            actor := g.gridmap.GetActor(neighbor)
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

func (g *GridEngine) openPartyMenu() {
    partyOptions := []renderer.MenuItem{
        {
            Text: "Inventory",
            Action: func() {
                g.openPartyInventory()
            },
        },
        {
            Text: "Magic",
            Action: func() {
                g.Print("Not implemented yet")
            },
        },
        {
            Text: "Rest",
            Action: func() {
                g.TryRestParty()
            },
        },
        {
            Text: "Attack",
            Action: func() {
                g.Print("Not implemented yet")
            },
        },
    }
    if g.playerParty.HasFollowers() {
        partyOptions = append(partyOptions, renderer.MenuItem{
            Text: "Split",
            Action: func() {
                g.openMenu(g.playerParty.GetSplitActions(g))
            },
        })
        if g.splitControlled != nil {
            partyOptions = append(partyOptions, renderer.MenuItem{
                Text: "Join",
                Action: func() {
                    g.TryJoinParty()
                },
            })
        }
    }

    g.openMenu(partyOptions)
}
func (g *GridEngine) SwitchAvatarTo(actor *game.Actor) {
    if !g.playerParty.IsMember(actor) {
        g.Print(fmt.Sprintf("\"%s\" is not in your party", actor.Name()))
        return
    }
    g.splitControlled = actor
    g.onAvatarMovedAlone()
}
func (g *GridEngine) onPartyMoved() {
    g.mapWindow.CenterOn(g.playerParty.Pos())
    g.updateContextActions()
    g.gridmap.UpdateFieldOfView(g.playerParty.GetFoV(), g.playerParty.Pos())
}

func (g *GridEngine) onAvatarMovedAlone() {
    g.mapWindow.CenterOn(g.GetAvatar().Pos())
    g.updateContextActions()
    g.gridmap.UpdateFieldOfView(g.playerParty.GetFoV(), g.GetAvatar().Pos())
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

func (g *GridEngine) AddToParty(npc *game.Actor) {
    if g.playerParty.IsFull() {
        g.Print(fmt.Sprintf("No room for \"%s\"", npc.Name()))
        return
    } else if g.playerParty.IsMember(npc) {
        g.Print(fmt.Sprintf("\"%s\" is already in your party", npc.Name()))
        return
    }
    g.playerParty.AddMember(npc)
}

func (g *GridEngine) TryJoinParty() {
    for _, member := range g.playerParty.GetMembers() {
        if member == g.avatar {
            continue
        }
        if !member.IsNextTo(g.avatar) {
            g.Print(fmt.Sprintf("\"%s\" is not next to you.", member.Name()))
            return
        }
    }
    g.splitControlled = nil
}

func (g *GridEngine) openVendorMenu(npc *game.Actor) {
    itemsToSell := npc.GetItemsToSell()
    if len(itemsToSell) == 0 {
        g.openSpeechWindow(npc, []string{"Unfortunately, I have nothing left to sell."}, func() {})
        return
    }
    var menuItems []renderer.MenuItem
    for _, i := range itemsToSell {
        offer := i
        itemLine := fmt.Sprintf("%s (%d)", offer.Item.Name(), offer.Price)
        menuItems = append(menuItems, renderer.MenuItem{
            Text: itemLine,
            Action: func() {
                g.TryBuyItem(npc, offer)
            },
        })
    }
    g.openMenu(menuItems)
}

func (g *GridEngine) TryBuyItem(npc *game.Actor, offer game.SalesOffer) {
    if g.playerParty.GetGold() < offer.Price {
        g.openSpeechWindow(npc, []string{"You don't have enough gold."}, func() {})
        return
    }
    npc.RemoveItem(offer.Item)
    npc.AddGold(offer.Price)

    g.playerParty.RemoveGold(offer.Price)
    g.playerParty.AddItem(offer.Item)
    g.openSpeechWindow(npc, []string{"Thank you for your business."}, func() {})
}

func (g *GridEngine) TryRestParty() {
    if g.playerParty.TryRest() {
        g.openSpeechWindow(g.GetAvatar(), []string{"You have eaten some food", "and rested the night.", "Your party has been healed."}, func() {})
    } else {
        g.Print("Not enough food to rest.")
    }
}

func secondsToTicks(seconds float64) int {
    return int(ebiten.ActualTPS() * seconds)
}
