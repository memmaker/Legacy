package main

import (
    "Legacy/ega"
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/gocoro"
    "Legacy/gridmap"
    "Legacy/ldtk_go"
    "Legacy/recfile"
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
    "path"
    "runtime/pprof"
    "sort"
    "strings"
)

type GridEngine struct {
    // Basic Game Engine
    wantsToQuit bool
    WorldTicks  uint64

    // Rules
    rules *game.Rules

    // Game State & Bookkeeping
    avatar          *game.Actor
    splitControlled *game.Actor
    playerParty     *game.Party
    playerKnowledge *game.PlayerKnowledge
    flags           *game.Flags
    activeEvents    []game.GameEvent
    mapsInMemory    map[string]*gridmap.GridMap[*game.Actor, game.Item, game.Object]

    // combat
    combatManager *CombatState

    // Map
    currentMap     *gridmap.GridMap[*game.Actor, game.Item, game.Object]
    ldtkMapProject *ldtk_go.Project
    spawnPosition  geometry.Point

    // Animation
    animationRoutine gocoro.Coroutine

    // UI
    deviceDPIScale float64
    tileScale      float64
    internalWidth  int
    internalHeight int
    textInput      *renderer.TextInput
    modalElement   Modal
    inputElement   UIWidget
    uiOverlay      map[int]int32

    foodButton         geometry.Point
    goldButton         geometry.Point
    defenseBuffsButton geometry.Point
    offenseBuffsButton geometry.Point

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

    textToPrint          string
    ticksForPrint        int
    printLog             []string
    grayScaleEntityTiles *ebiten.Image
    allowedPartyIcons    []int32
    onMoveSubscribers    []func(newPosition geometry.Point) bool
    isGameOver           bool
}

func (g *GridEngine) GetVisibleMap() geometry.Rect {
    return g.mapWindow.GetVisibleMap()
}

func (g *GridEngine) GetDialogueFromFile(conversationId string) *game.Dialogue {
    filename := path.Join("assets", "dialogues", conversationId+".txt")
    file := mustOpen(filename)
    records := recfile.Read(file)
    _ = file.Close()
    return game.NewDialogueFromRecords(records, g.gridRenderer.AutolayoutArrayToIconPages)
}

func (g *GridEngine) GetActorByInternalName(internalName string) *game.Actor {
    for _, actor := range g.currentMap.Actors() {
        if actor.GetInternalName() == internalName {
            return actor
        }
    }
    return nil
}

func (g *GridEngine) RaiseAsUndeadForParty(pos geometry.Point) {
    downedActor, exists := g.currentMap.TryGetDownedActorAt(pos)
    if !exists {
        return
    }
    g.currentMap.RemoveDownedActor(downedActor)
    undeadActor := game.NewActor("Skeletor", 109)
    g.currentMap.AddActor(undeadActor, pos)
    g.playerParty.AddMember(undeadActor)
}

func (g *GridEngine) GetRegion(regionName string) geometry.Rect {
    return g.currentMap.GetNamedRegion(regionName)
}

func (g *GridEngine) RemoveDoorAt(pos geometry.Point) {
    objectAt := g.currentMap.ObjectAt(pos)
    if _, isDoor := objectAt.(*game.Door); isDoor {
        g.currentMap.RemoveObjectAt(pos)
    }
}

func (g *GridEngine) GetRandomPositionsInRegion(regionName string, count int) []geometry.Point {
    region := g.currentMap.GetNamedRegion(regionName)
    setOfLocations := make(map[geometry.Point]bool)
    for len(setOfLocations) < count {
        setOfLocations[region.GetRandomPoint()] = true
    }
    var result []geometry.Point
    for k := range setOfLocations {
        result = append(result, k)
    }
    return result
}

func (g *GridEngine) GetGridMap() *gridmap.GridMap[*game.Actor, game.Item, game.Object] {
    return g.currentMap
}

func (g *GridEngine) TeleportTo(mirrorText string) {
    if mirrorText == "" {
        return
    }
    mapName, locationName := g.rules.GetTargetLocation(mirrorText)
    if mapName == "" || locationName == "" {
        return
    }
    g.transition(mapName, locationName)
}

func (g *GridEngine) GetBreakingToolName() string {
    return g.playerParty.GetNameOfBreakingTool()
}

func (g *GridEngine) ProdActor(prodder *game.Actor, victim *game.Actor) {
    if !prodder.IsRightNextTo(victim) {
        return
    }
    // check if dest is free
    direction := victim.Pos().Sub(prodder.Pos())
    dest := victim.Pos().Add(direction)
    if !g.currentMap.IsCurrentlyPassable(dest) {
        return
    }

    g.currentMap.MoveActor(victim, dest)
}

func (g *GridEngine) FreezeActorAt(pos geometry.Point, turns int) {
    if !g.currentMap.IsActorAt(pos) {
        return
    }

    actor := g.currentMap.ActorAt(pos)
    actor.SetTintColor(ega.BrightBlue)
    actor.SetTinted(true)

    // TODO, actually freeze them..
}

func (g *GridEngine) CanLevelUp(member *game.Actor) (bool, int) {
    return g.rules.CanLevelUp(member.GetLevel(), member.GetXP())
}

func (g *GridEngine) GetRules() *game.Rules {
    return g.rules
}

func (g *GridEngine) GetPartyEquipment() []game.Item {
    return g.playerParty.GetFlatInventory()
}

func (g *GridEngine) GetAoECircle(source geometry.Point, radius int) []geometry.Point {
    current := source
    var result []geometry.Point
    for x := -radius; x <= radius; x++ {
        for y := -radius; y <= radius; y++ {
            if x*x+y*y <= radius*radius {
                result = append(result, current.Add(geometry.Point{X: x, Y: y}))
            }
        }
    }
    return result
}

func (g *GridEngine) HitAnimation(pos geometry.Point, atlasName renderer.AtlasName, icon int32, tintColor color.Color, whenDone func()) {
    g.combatManager.addHitAnimation(pos, atlasName, icon, tintColor, whenDone)
}

func (g *GridEngine) SpellDamageAt(caster *game.Actor, pos geometry.Point, amount int) {
    if g.currentMap.IsActorAt(pos) {
        actor := g.currentMap.ActorAt(pos)
        g.DeliverSpellDamage(caster, actor, amount)
        if !actor.IsAlive() {
            g.actorDied(actor)
            g.combatManager.findAlliesOfOpponent(actor)
        } else if !g.IsPlayerControlled(actor) {
            g.combatManager.addOpponent(actor)
        }
    }
}

func (g *GridEngine) StartCombat(opponents *game.Actor) {
    g.combatManager.PlayerStartsMeleeAttack(g.GetAvatar(), opponents)
}

func main() {
    // Create a CPU profile file
    cpuProfileFile, err := os.Create("cpu.prof")
    if err != nil {
        panic(err)
    }
    defer cpuProfileFile.Close()

    // start CPU profiling
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
        tileScale:            tileScaleFactor,
        internalWidth:        internalScreenWidth,
        internalHeight:       internalScreenHeight,
        worldTiles:           ebiten.NewImageFromImage(mustLoadImage("assets/world.png")),
        entityTiles:          ebiten.NewImageFromImage(mustLoadImage("assets/entities.png")),
        grayScaleEntityTiles: ebiten.NewImageFromImage(mustLoadImage("assets/entities_gs.png")),
        uiTiles:              ebiten.NewImageFromImage(mustLoadImage("assets/charset.png")),
        uiOverlay:            make(map[int]int32),
        mapsInMemory:         make(map[string]*gridmap.GridMap[*game.Actor, game.Item, game.Object]),
        animationRoutine:     gocoro.NewCoroutine(),
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

    if x == g.offenseBuffsButton.X && y == g.offenseBuffsButton.Y {
        offBuffs := g.playerParty.GetOffenseBuffs()
        if len(offBuffs) > 0 {
            g.ShowFixedFormatText(offBuffs)
        }
        return
    }

    if x == g.defenseBuffsButton.X && y == g.defenseBuffsButton.Y {
        defBuffs := g.playerParty.GetDefenseBuffs()
        if len(defBuffs) > 0 {
            g.ShowFixedFormatText(defBuffs)
        }
        return
    }

    if y == screenSize.Y-2 {
        if x == g.foodButton.X {
            g.TryRestParty()
        } else if x == g.goldButton.X {
            g.openFinanceOverview()
        }
    } else if y == screenSize.Y-1 {
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
    } else if g.inputElement != nil && g.inputElement.OnMouseClicked(x, y) {
        return
    } else if g.modalElement != nil {
        g.modalElement.OnMouseClicked(x, y)
    }
}

func (g *GridEngine) onViewedActorMoved(newPosition geometry.Point) {
    g.mapWindow.CenterOn(newPosition)
    g.updateContextActions()
    g.currentMap.UpdateFieldOfView(g.playerParty.GetFoV(), newPosition)
}

func (g *GridEngine) drawPrintMessage(screen *ebiten.Image, upper bool) {
    screenSize := g.gridRenderer.GetSmallGridScreenSize()
    offsetY := 1
    if upper {
        offsetY = 2
    }
    yPos := screenSize.Y - offsetY
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
func (g *GridEngine) SetWallAt(pos geometry.Point) {
    currentCell := g.currentMap.GetCell(pos)
    currentTile := currentCell.TileType
    if currentTile.Special == gridmap.SpecialTileNone {
        grassTile := currentTile.WithIcon(80).WithIsWalkable(false).WithIsTransparent(false)
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

func (g *GridEngine) breakTileAt(loc geometry.Point) {
    currentCell := g.currentMap.GetCell(loc)
    currentTile := currentCell.TileType
    if currentTile.IsBreakable() {
        debrisTile := currentTile.WithIcon(currentTile.GetDebrisTile()).
            WithIsWalkable(true).
            WithIsTransparent(true).
            WithSpecial(gridmap.SpecialTileNone)
        g.currentMap.SetTile(loc, debrisTile)
        if currentTile.Special == gridmap.SpecialTileBreakableGold {
            goldItem := game.NewPseudoItemFromTypeAndAmount(game.PseudoItemTypeGold, 200)
            g.currentMap.AddItem(goldItem, loc)
        }
        g.updateContextActions()
    }
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

func (g *GridEngine) createWeapons(level int, amount int) []game.Item {
    var weapons []game.Item
    for i := 0; i < amount; i++ {
        weapons = append(weapons, game.NewRandomWeapon(level))
    }
    return weapons
}

func (g *GridEngine) EquipItem(actor *game.Actor, a game.Wearable) {
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
    if !g.IsPlayerControlled(actor) {
        g.currentMap.SetActorToDowned(actor) // IS THIS A GOOD IDEA?
        g.Print(fmt.Sprintf("'%s' died", actor.Name()))
        // award xp for the kill
        g.AddXP(actor.GetXPForKilling())
    }
}

func (g *GridEngine) AddXP(xp int) {
    g.playerParty.AddXPForEveryone(xp)
    g.Print(fmt.Sprintf("%d XP awarded", xp))
}

func (g *GridEngine) dropActorInventory(actor *game.Actor) {
    droppedItems := actor.DropInventory()
    if len(droppedItems) == 0 {
        return
    }
    chest := game.NewFixedChest(droppedItems)
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

func (g *GridEngine) setGameOver() {
    g.isGameOver = true
    g.ShowText([]string{"Game Over"})
}

func (g *GridEngine) IsWindowOpen() bool {
    return g.modalElement != nil || g.inputElement != nil
}

func (g *GridEngine) addToJournal(npc *game.Actor, text []string) {
    g.Print("Added to journal.")
    g.playerKnowledge.AddJournalEntry(npc.Name(), text, g.CurrentTick())
}

func (g *GridEngine) openJournal() {
    if g.playerKnowledge.IsJournalEmpty() {
        g.ShowText([]string{"You don't have any journal entries."})
        return
    }
    g.openMenu([]renderer.MenuItem{
        {
            Text: "Chronological",
            Action: func() {
                g.ShowFixedFormatText(g.playerKnowledge.GetChronologicalJournal())
            },
        },
        {
            Text: "By Source",
            Action: func() {
                sources := g.playerKnowledge.GetJournalSources()
                g.openMenu(g.toJournalMenu(sources))
            },
        },
    })

}

func (g *GridEngine) toJournalMenu(sources []string) []renderer.MenuItem {
    var items []renderer.MenuItem
    for _, s := range sources {
        source := s
        items = append(items, renderer.MenuItem{
            Text: source,
            Action: func() {
                g.ShowFixedFormatText(g.playerKnowledge.GetJournalBySource(source))
            },
        })
    }
    return items
}

func (g *GridEngine) ScreenToMap(screenX int, screenY int) geometry.Point {
    bigGrid := g.gridRenderer.GetScaledBigGridSize()
    smallGrid := g.gridRenderer.GetScaledSmallGridSize()
    screenGridPos := geometry.Point{X: (screenX - smallGrid) / bigGrid, Y: (screenY - smallGrid) / bigGrid}
    return g.mapWindow.GetMapGridPositionFromScreenGridPosition(screenGridPos)
}

func (g *GridEngine) openPartyOverView() {
    g.ShowFixedFormatText(g.playerParty.GetPartyOverview())
}

func (g *GridEngine) DeliverCombatDamage(attacker *game.Actor, victim *game.Actor) {
    damage := g.rules.GetMeleeDamage(attacker, victim)

    if damage > 0 {
        victim.Damage(damage)
        g.Print(fmt.Sprintf("%d dmg. to '%s'", damage, victim.Name()))
    } else {
        g.Print(fmt.Sprintf("No dmg. to '%s'", victim.Name()))
    }

}

func (g *GridEngine) DeliverSpellDamage(attacker *game.Actor, victim *game.Actor, amount int) {
    baseSpellDamage := amount
    victimDefense := victim.GetMagicDefense()
    damage := baseSpellDamage - victimDefense
    if damage > 0 {
        victim.Damage(damage)
        g.Print(fmt.Sprintf("%d dmg. to '%s'", damage, victim.Name()))
    } else {
        g.Print(fmt.Sprintf("No dmg. to '%s'", victim.Name()))
    }
}

func (g *GridEngine) openDismissMenu() {
    var menuItems []renderer.MenuItem
    for _, m := range g.playerParty.GetMembers() {
        if m == g.avatar {
            continue
        }
        member := m
        menuItems = append(menuItems, renderer.MenuItem{
            Text: member.Name(),
            Action: func() {
                if g.IsPlayerControlled(member) {
                    if g.GetAvatar() == member {
                        g.SwitchAvatarTo(g.avatar)
                    }
                    g.playerParty.RemoveMember(member)
                    g.Print(fmt.Sprintf("'%s' left the party.", member.Name()))
                }
            },
        })
    }
    g.openMenu(menuItems)
}

func (g *GridEngine) openFinanceOverview() {
    g.ShowFixedFormatText(g.playerParty.GetFinanceOverview(g))
}

func (g *GridEngine) getAllLoadedMaps() map[string]*gridmap.GridMap[*game.Actor, game.Item, game.Object] {
    mapsInMemory := g.mapsInMemory
    mapsInMemory[g.currentMap.GetName()] = g.currentMap
    return mapsInMemory
}

func (g *GridEngine) openPrintLog() {
    g.ShowFixedFormatText(g.printLog)
}

func (g *GridEngine) AddSkill(avatar *game.Actor, skill string) {
    avatar.GetSkills().IncrementSkill(game.SkillName(skill))
    g.Print(fmt.Sprintf("'%s' learned '%s'", avatar.Name(), skill))
}

func (g *GridEngine) AddBuff(actor *game.Actor, name string, buffType game.BuffType, strength int) {
    actor.AddBuff(name, buffType, strength)
    g.Print(fmt.Sprintf("%s received %s", actor.Name(), name))
}

func (g *GridEngine) AskUserForString(prompt string, maxLength int, onConfirm func(text string)) {
    textInput := g.gridRenderer.NewTextInputAtY(10, prompt, func(endedWith renderer.EndAction, text string) {
        if endedWith == renderer.EndActionConfirm {
            onConfirm(text)
        }
    })
    textInput.SetMaxLength(maxLength)
    textInput.CenterHorizontallyAtY(10)
    g.textInput = textInput
}

func (g *GridEngine) loadPartyIcons() {
    tileset := g.ldtkMapProject.TilesetByIdentifier("Entities")
    for tileID, enums := range tileset.Enums {
        if enums.Contains("IsAllowedAsPartyIcon") {
            g.allowedPartyIcons = append(g.allowedPartyIcons, int32(tileID))
        }
    }
    sort.SliceStable(g.allowedPartyIcons, func(i, j int) bool {
        return g.allowedPartyIcons[i] < g.allowedPartyIcons[j]
    })
}

func (g *GridEngine) CreateItemsForVendor(lootType game.Loot, level int) []game.Item {
    switch lootType {
    case game.LootArmor:
        return g.createArmor(level, 10)
    case game.LootWeapon:
        return g.createWeapons(level, 10)
    case game.LootHealer:
        return g.createPotions(10)
    case game.LootPotions:
        return g.createPotions(10)
    }
    return []game.Item{}
}

func (g *GridEngine) subscribeToMovement(subscriber func(pos geometry.Point) bool) {
    g.onMoveSubscribers = append(g.onMoveSubscribers, subscriber)
}

func (g *GridEngine) goBackToBed() {
    g.ShowMultipleChoiceDialogue(g.GetAvatar().Icon(0), g.gridRenderer.AutolayoutArrayToIconPages(5, []string{"You are sure your walls and the mirror will be back to normal, if you just go back to sleep"}), []renderer.MenuItem{
        {
            Text: "Stay awake",
            Action: func() {
                g.inputElement = nil
                g.modalElement = nil
            },
        },
        {
            Text: "Go back to bed",
            Action: func() {
                g.inputElement = nil
                g.modalElement = nil
                g.setGameOver()
            },
        },
    })
}
func (g *GridEngine) drawTextInWorld(worldPos geometry.Point, text string) {
    for i, c := range text {
        drawPos := worldPos.Add(geometry.Point{X: i, Y: 0})
        g.DrawCharInWorld(c, drawPos)
    }
}

func (g *GridEngine) DrawCharInWorld(c rune, drawPos geometry.Point) {
    fontIndex := getWorldFontIndex()
    tileIndex := fontIndex[c]
    g.currentMap.SetTileIcon(drawPos, int32(tileIndex))
}

func (g *GridEngine) SetMapIcon(icon int32, mapPos geometry.Point) {
    g.currentMap.SetTileIcon(mapPos, icon)
}

func (g *GridEngine) getTombstoneFromEntity(entity *ldtk_go.Entity) game.Object {
    isHoly := entity.PropertyByIdentifier("IsHoly").AsBool()
    inscriptionProperty := entity.PropertyByIdentifier("Inscription")
    if inscriptionProperty == nil || inscriptionProperty.IsNull() {
        return game.NewTombstone(isHoly)
    }
    tombstone := game.NewTombstone(isHoly)
    tombstone.SetDescription(strings.Split(inscriptionProperty.AsString(), "\n"))
    return tombstone
}
func secondsToTicks(seconds float64) int {
    return int(ebiten.ActualTPS() * seconds)
}
