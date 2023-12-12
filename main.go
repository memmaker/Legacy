package main

import (
    "Legacy/dungen"
    "Legacy/ega"
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/gocoro"
    "Legacy/gridmap"
    "Legacy/ldtk_go"
    "Legacy/recfile"
    "Legacy/renderer"
    "Legacy/ui"
    "Legacy/util"
    "bufio"
    "errors"
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
    _ "image/png"
    "log"
    "math"
    "os"
    "path"
    "runtime/pprof"
    "sort"
    "strings"
    "time"
)

type EngineConfiguration struct {
    AlwaysCenter bool
}

type GridEngine struct {
    // Basic Game Engine
    wantsToQuit bool
    WorldTicks  uint64

    // Rules
    rules *game.Rules

    // config
    config EngineConfiguration

    // Game State & Bookkeeping
    avatar *game.Actor

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
    animator         *renderer.Animator
    animationRoutine gocoro.Coroutine
    movementRoutine  gocoro.Coroutine

    // UI
    deviceDPIScale float64
    tileScale      float64
    internalWidth  int
    internalHeight int

    modalStack        []Modal
    conversationModal *ui.ConversationModal

    currentTooltip           ui.Tooltip
    ticksUntilTooltipAppears int

    uiOverlay map[int]int32

    foodButton geometry.Point
    goldButton geometry.Point
    bagButton  geometry.Point

    defenseBuffsButton geometry.Point
    offenseBuffsButton geometry.Point

    gridRenderer  *renderer.DualGridRenderer
    mapRenderer   *renderer.MapRenderer
    mapWindow     *renderer.MapWindow
    lastMousePosX int
    lastMousePosY int

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

    isGameOver              bool
    wheelYVelocity          float64
    allTransitions          map[string][]NamedLocation
    worldTime               game.WorldTime
    lastInteractionWasMouse bool
    altIsPressed            bool
    levelHooks              game.LevelHooks
    isSneaking              bool
    debugInfoMode           bool
    overlayPositions        map[geometry.Point]color.Color
}

func testDungeonGenerator() {
    gen := dungen.NewAccretionGenerator()

    // get input from keyboard
    scanner := bufio.NewScanner(os.Stdin)
    text := ""
    for text != "q" {
        dunMap := gen.Generate(32, 16)
        dunMap.Print()
        scanner.Scan()
        text = scanner.Text()
    }
}
func main() {
    /*
       testDungeonGenerator()
       return

    */

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
        movementRoutine:      gocoro.NewCoroutine(),
        overlayPositions:     make(map[geometry.Point]color.Color),
    }
    ebiten.SetWindowTitle(gameTitle)
    ebiten.SetWindowSize(scaledScreenWidth, scaledScreenHeight)
    ebiten.SetScreenClearedEveryFrame(true)

    gridEngine.animator = renderer.NewAnimator(func(pos geometry.Point) (bool, geometry.Point) {
        isOnScreen := gridEngine.IsMapPosOnScreen(pos)
        if isOnScreen {
            return true, gridEngine.MapToScreenCoordinates(pos)
        }
        return false, geometry.Point{}
    })

    gridEngine.Init()

    if err := ebiten.RunGameWithOptions(gridEngine, &ebiten.RunGameOptions{
        GraphicsLibrary: ebiten.GraphicsLibraryOpenGL,
    }); err != nil && !errors.Is(err, ebiten.Termination) {
        log.Fatal(err)
    }
}
func (g *GridEngine) MeleeAttackWithFixedDamage(attacker *game.Actor, target *game.Actor, damage int) {
    g.combatManager.MeleeAttack(attacker, target)
}

func (g *GridEngine) IsSneaking() bool {
    return g.isSneaking
}

func (g *GridEngine) TryMoveNPCOnPath(actor *game.Actor, dest geometry.Point) {
    if !actor.IsSleeping() && actor.IsAlive() {
        if g.currentMap.IsCurrentlyPassable(dest) {
            g.currentMap.MoveActor(actor, dest)
        }
    }
}

func (g *GridEngine) UnlockDoorsByKeyName(keyName string) {
    for _, o := range g.currentMap.Objects() {
        if door, isDoor := o.(*game.Door); isDoor {
            if door.NeedsKey() && door.GetKeyName() == keyName {
                door.Unlock()
            }
        }
    }
}

func (g *GridEngine) GetChestByInternalName(internalName string) *game.Chest {
    for _, o := range g.currentMap.Objects() {
        if chest, isChest := o.(*game.Chest); isChest {
            if chest.GetInternalName() == internalName {
                return chest
            }
        }
    }
    return nil
}

func (g *GridEngine) ResetAllLockedDoorsOnMap(mapName string) {
    if !g.isMapInMemory(mapName) {
        return
    }
    namedMap := g.mapsInMemory[mapName]
    for _, object := range namedMap.Objects() {
        if door, isDoor := object.(*game.Door); isDoor {
            if door.NeedsKey() {
                door.Reset()
            }
        }
    }
}

func (g *GridEngine) HasSkill(skill game.SkillName) bool {
    return g.avatar.GetSkills().HasSkill(skill)
}

func (g *GridEngine) GetRelativeDifficulty(skill game.SkillName, difficulty game.DifficultyLevel) game.DifficultyLevel {
    avatar := g.GetAvatar()
    skills := avatar.GetSkills()
    if !skills.HasSkill(skill) {
        return game.DifficultyLevelImpossible
    }
    skillLevel := skills.GetLevel(skill)
    return g.rules.GetRelativeDifficulty(skillLevel, difficulty)
}
func (g *GridEngine) SkillCheckAvatar(skill game.SkillName, difficulty game.DifficultyLevel) bool {
    return g.SkillCheck(g.GetAvatar(), skill, difficulty)
}
func (g *GridEngine) SkillCheckAvatarVs(skill game.SkillName, antagonist *game.Actor, attribute game.AttributeName) bool {
    return g.SkillCheck(g.GetAvatar(), skill, antagonist.GetAbsoluteDifficultyByAttribute(attribute))
}
func (g *GridEngine) SkillCheckVs(actor *game.Actor, skill game.SkillName, antagonist *game.Actor, attribute game.AttributeName) bool {
    return g.SkillCheck(actor, skill, antagonist.GetAbsoluteDifficultyByAttribute(attribute))
}
func (g *GridEngine) SkillCheck(actor *game.Actor, skill game.SkillName, difficulty game.DifficultyLevel) bool {
    skills := actor.GetSkills()
    if !skills.HasSkill(skill) {
        return game.RollChance(0.01)
    }
    skillLevel := skills.GetLevel(skill)
    return g.rules.RollSkillCheck(skillLevel, difficulty)
}

func (g *GridEngine) OnMouseWheel(x int, y int, dy float64) bool {
    return false
}

func (g *GridEngine) GetWorldTime() game.WorldTime {
    return g.worldTime
}

func (g *GridEngine) AdvanceWorldTime(days, hours, minutes int) {
    g.worldTime = g.worldTime.WithAddedDays(days).WithAddedMinutes(hours*game.MinutesPerHour + minutes)
}

func (g *GridEngine) AdvanceWorldTimeWithMessage(days, hours, minutes int) {
    g.worldTime = g.worldTime.WithAddedDays(days).WithAddedMinutes(hours*game.MinutesPerHour + minutes)
    g.printTimePassedMessage(days, hours, minutes)
}

func (g *GridEngine) printTimePassedMessage(days int, hours int, minutes int) {
    if days == 0 {
        if hours == 0 {
            g.Print(fmt.Sprintf("%d minutes passed", minutes))
        } else {
            if minutes == 0 {
                g.Print(fmt.Sprintf("%d hours passed", hours))
            } else {
                g.Print(fmt.Sprintf("%d hours, %d minutes passed", hours, minutes))
            }
        }
    } else {
        if hours == 0 {
            if minutes == 0 {
                g.Print(fmt.Sprintf("%d days passed", days))
            } else {
                g.Print(fmt.Sprintf("%d days, %d minutes passed", days, minutes))
            }
        } else {
            if minutes == 0 {
                g.Print(fmt.Sprintf("%d days, %d hours passed", days, hours))
            } else {
                g.Print(fmt.Sprintf("%d days, %d hours, %d minutes passed", days, hours, minutes))
            }
        }
    }
}

func (g *GridEngine) PlayerStartsOffensiveSpell(caster *game.Actor, spell *game.Spell) {
    if !spell.CanPayCost(g, caster) {
        g.Print("Not enough mana!")
        return
    }
    g.combatManager.PlayerUsesActiveSkill(caster, spell)
}

func (g *GridEngine) OnCommand(command ui.CommandType) bool {
    if g.IsInCombat() {
        return false
    }
    switch command {
    case ui.PlayerCommandUp:
        g.ActionUp()
    case ui.PlayerCommandDown:
        g.ActionDown()
    case ui.PlayerCommandLeft:
        g.ActionLeft()
    case ui.PlayerCommandRight:
        g.ActionRight()
    case ui.PlayerCommandConfirm:
        g.ActionConfirm()
    case ui.PlayerCommandCancel:
        g.ActionCancel()
    }
    return true
}

func (g *GridEngine) GetParty() *game.Party {
    return g.playerParty
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

func (g *GridEngine) RaiseAsUndeadAt(caster *game.Actor, pos geometry.Point) {
    undeadActor, exists := g.currentMap.TryGetDownedActorAt(pos)
    if !exists {
        return
    }
    undeadActor.SetIcon(109)
    undeadActor.SetHealth(1)
    g.AddStatusEffect(undeadActor, game.StatusUndead(), 1)
    g.currentMap.SetActorToNormal(undeadActor)
    //g.currentMap.AddActor(undeadActor, pos)
    if g.IsPlayerControlled(caster) {
        g.playerParty.AddMember(undeadActor)
    } else {
        undeadActor.SetCombatFaction(caster.GetCombatFaction())
    }
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
    mapName, locationName := g.rules.GetTargetTravelLocation(mirrorText)
    if mapName == "" || locationName == "" {
        return
    }
    g.TransitionToNamedLocation(mapName, locationName)
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
    if victim.IsAlive() {
        g.currentMap.MoveActor(victim, dest)
    } else {
        g.currentMap.MoveDownedActor(victim, dest)
    }
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

func (g *GridEngine) CombatHitAnimation(pos geometry.Point, atlasName renderer.AtlasName, icon int32, tintColor color.Color, whenDone func()) {
    g.combatManager.animator.AddDefaultHitAnimation(pos, atlasName, icon, tintColor, whenDone)
}

func (g *GridEngine) TileAnimation(pos geometry.Point, atlasName renderer.AtlasName, icon int32, tintColor color.Color, whenDone func()) {
    g.animator.AddDefaultHitAnimation(pos, atlasName, icon, tintColor, whenDone)
}

func (g *GridEngine) FixedDamageAt(caster *game.Actor, pos geometry.Point, amount int) {
    if g.currentMap.IsActorAt(pos) {
        actor := g.currentMap.ActorAt(pos)
        g.DeliverSpellDamage(caster, actor, amount)
        g.combatManager.OnCombatAction(caster, actor)
        if !actor.IsAlive() {
            g.actorDied(actor)
        }
    }
}
func (g *GridEngine) PlayerTriesBackstab(opponent *game.Actor) {
    if !opponent.IsAlive() {
        g.Print(fmt.Sprintf("'%s' is already dead.", opponent.Name()))
        return
    }
    if g.SkillCheckAvatarVs(game.PhysicalSkillBackstab, opponent, game.Perception) {
        g.Kill(opponent)
    } else {
        g.Print("Your attack was noticed!")
        g.onCriminalOffense(opponent)
    }
}
func (g *GridEngine) PlayerStartsCombat(opponent *game.Actor) {
    if !opponent.IsAlive() {
        g.Print(fmt.Sprintf("'%s' is already dead.", opponent.Name()))
        return
    }
    g.combatManager.MeleeAttack(g.GetAvatar(), opponent)
}

func (g *GridEngine) EnemyStartsCombat(opponent *game.Actor) {
    if opponent.IsSleeping() || !opponent.IsAlive() {
        return
    }
    g.combatManager.EnemyStartsCombat(opponent, g.GetAvatar())
}

func (g *GridEngine) Reset() {
    g.WorldTicks = 0
}

func (g *GridEngine) QuitGame() {
    g.wantsToQuit = true
}

func (g *GridEngine) handleTooltip(tooltip ui.Tooltip) {
    g.handleTooltipWithDelay(tooltip, 0.5)
}

func (g *GridEngine) handleTooltipWithDelay(tooltip ui.Tooltip, delay float64) {
    if tooltip.IsNull() {
        g.currentTooltip = nil
    } else {
        g.currentTooltip = tooltip
        g.ticksUntilTooltipAppears = int(ebiten.ActualTPS() * delay)
    }
}

func (g *GridEngine) MapToScreenCoordinates(pos geometry.Point) geometry.Point {
    return g.mapWindow.GetScreenGridPositionFromMapGridPosition(pos)
}

// OnMouseMoved receives the coordinates as character cells
func (g *GridEngine) OnMouseMoved(x int, y int) (bool, ui.Tooltip) {
    screenSize := g.gridRenderer.GetSmallGridScreenSize()
    oneFourth := screenSize.X / 4
    if y == screenSize.Y-1 {
        // each 1/4 of the screen is a different UI
        if x < oneFourth {
            if x == 0 {
                member := g.playerParty.GetMember(0)
                return true, ui.NewTextTooltip(g.gridRenderer, []string{fmt.Sprintf("HP: %d/%d", member.GetHealth(), member.GetMaxHealth())}, geometry.Point{X: x, Y: y})
            } else {
                return false, ui.NoTooltip{}
            }
        } else if x < oneFourth*2 {
            if x == oneFourth && len(g.playerParty.GetMembers()) > 1 {
                member := g.playerParty.GetMember(1)
                return true, ui.NewTextTooltip(g.gridRenderer, []string{fmt.Sprintf("HP: %d/%d", member.GetHealth(), member.GetMaxHealth())}, geometry.Point{X: x, Y: y})
            } else {
                return false, ui.NoTooltip{}
            }
        } else if x < oneFourth*3 && len(g.playerParty.GetMembers()) > 2 {
            if x == oneFourth*2 {
                member := g.playerParty.GetMember(2)
                return true, ui.NewTextTooltip(g.gridRenderer, []string{fmt.Sprintf("HP: %d/%d", member.GetHealth(), member.GetMaxHealth())}, geometry.Point{X: x, Y: y})
            } else {
                return false, ui.NoTooltip{}
            }
        } else {
            if x == oneFourth*3 && len(g.playerParty.GetMembers()) > 3 {
                member := g.playerParty.GetMember(3)
                return true, ui.NewTextTooltip(g.gridRenderer, []string{fmt.Sprintf("HP: %d/%d", member.GetHealth(), member.GetMaxHealth())}, geometry.Point{X: x, Y: y})
            } else {
                return false, ui.NoTooltip{}
            }
        }
    }
    return false, ui.NoTooltip{}
}

// OnMouseClicked receives the coordinates as character cells
func (g *GridEngine) OnMouseClicked(x int, y int) bool {
    screenSize := g.gridRenderer.GetSmallGridScreenSize()
    oneFourth := screenSize.X / 4

    if x == g.offenseBuffsButton.X && y == g.offenseBuffsButton.Y {
        /*
           offBuffs := g.playerParty.GetOffenseBuffs()
           if len(offBuffs) > 0 {
               g.ShowFixedFormatText(offBuffs)
           }

        */
        return true
    }

    if x == g.defenseBuffsButton.X && y == g.defenseBuffsButton.Y {

        /*defBuffs := g.playerParty.GetDefenseBuffs()

          if len(defBuffs) > 0 {
              g.ShowFixedFormatText(defBuffs)
          }
        */
        return true
    }

    if y == screenSize.Y-2 {
        if x == g.foodButton.X {
            g.TryRestParty()
            return true
        } else if x == g.goldButton.X {
            g.openFinanceOverview()
            return true
        } else if x == g.bagButton.X {
            g.openExtendedInventory()
            return true
        }
    } else if y == screenSize.Y-1 {
        // each 1/4 of the screen is a different UI
        if x < oneFourth {
            if x == 0 {
                g.openCharMainStats(0)
            } else {
                g.OpenEquipmentDetails(0)
            }
        } else if x < oneFourth*2 {
            if x == oneFourth {
                g.openCharMainStats(1)
            } else {
                g.OpenEquipmentDetails(1)
            }
        } else if x < oneFourth*3 {
            if x == oneFourth*2 {
                g.openCharMainStats(2)
            } else {
                g.OpenEquipmentDetails(2)
            }
        } else {
            if x == oneFourth*3 {
                g.openCharMainStats(3)
            } else {
                g.OpenEquipmentDetails(3)
            }
        }
        return true
    }
    return false
}

func (g *GridEngine) onViewedActorMoved(newPosition geometry.Point) {
    if g.config.AlwaysCenter {
        g.mapWindow.CenterOn(newPosition)
    } else {
        g.mapWindow.EnsurePositionIsInview(newPosition, 5)
    }
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

    xBeforeText := xOffsetForCenter - 1
    xAfterText := xOffsetForCenter + textLen
    g.gridRenderer.DrawColoredString(screen, xOffsetForCenter, yPos, g.textToPrint, color.White)
    if upper {
        g.gridRenderer.DrawOnSmallGrid(screen, xBeforeText, yPos, 16)
        g.gridRenderer.DrawOnSmallGrid(screen, xAfterText, yPos, 17)
    }
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
    }
}

func (g *GridEngine) TryPickpocketItem(item game.Item, victim *game.Actor) {
    g.flags.IncrementFlag("pickpocket_attempts")

    if g.SkillCheckAvatarVs(game.ThievingSkillPickpocket, victim, game.Perception) {
        g.PickPocketItem(item, victim)
        g.Print(fmt.Sprintf("You stole \"%s\"", item.Name()))
        g.flags.IncrementFlag("pickpocket_successes")
    } else {
        g.Print(fmt.Sprintf("You were caught stealing \"%s\"", item.Name()))
        g.onCriminalOffense(victim)
    }
}

func (g *GridEngine) onCriminalOffense(victim *game.Actor) {
    offenseEvent := g.rules.GetCriminalOffenseEvent(g.currentMap.GetName())
    if offenseEvent != "" {
        g.TriggerEvent(offenseEvent)
    } else {
        g.EnemyStartsCombat(victim)
    }
}

func (g *GridEngine) TryPlantItem(item game.Item, victim *game.Actor) {
    g.flags.IncrementFlag("plant_attempts")
    if g.SkillCheckAvatarVs(game.ThievingSkillPickpocket, victim, game.Perception) {
        g.PlantItem(item, victim)
        g.Print(fmt.Sprintf("You planted \"%s\"", item.Name()))
        g.flags.IncrementFlag("plant_successes")
    } else {
        g.Print(fmt.Sprintf("You were caught planting \"%s\"", item.Name()))
        g.onCriminalOffense(victim)
    }
}

func (g *GridEngine) createPotions(level int) []game.Item {
    var potions []game.Item
    for i := 0; i < level; i++ {
        potions = append(potions, game.NewPotion())
    }
    return potions
}

func (g *GridEngine) createArmorForVendor(level, amount int) []game.Item {
    var armor []game.Item
    for i := 0; i < amount; i++ {
        armor = append(armor, game.NewRandomArmorForVendor(level))
    }
    return armor
}
func (g *GridEngine) createArmorForLoot(level, amount int) []game.Item {
    var armor []game.Item
    for i := 0; i < amount; i++ {
        armor = append(armor, game.NewRandomArmor(level))
    }
    return armor
}

func (g *GridEngine) createWeaponsForVendor(level int, amount int) []game.Item {
    var weapons []game.Item
    for i := 0; i < amount; i++ {
        weapons = append(weapons, game.NewRandomWeaponForVendor(level))
    }
    return weapons
}

func (g *GridEngine) createWeaponsForLoot(level int, amount int) []game.Item {
    var weapons []game.Item
    for i := 0; i < amount; i++ {
        weapons = append(weapons, game.NewRandomWeapon(level))
    }
    return weapons
}

func (g *GridEngine) EquipItem(actor *game.Actor, a game.Equippable) {
    actor.Equip(a)
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
        // award xp for the Kill
        g.AddXP(actor.GetXPForKilling())
        println(fmt.Sprintf("'%s' died at %s", actor.Name(), actor.Pos().String()))
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
func (g *GridEngine) topModal() Modal {
    return g.modalStack[len(g.modalStack)-1]
}
func (g *GridEngine) IsWindowOpen() bool {
    return g.IsModalOpen() || g.IsInConversation()
}

func (g *GridEngine) IsModalOpen() bool {
    return g.modalStack != nil && len(g.modalStack) > 0
}

func (g *GridEngine) addToJournal(source string, text []string) {
    g.Print("Added to journal.")
    g.playerKnowledge.AddJournalEntry(source, text, g.CurrentTick())
}

func (g *GridEngine) openJournal() {
    if g.playerKnowledge.IsJournalEmpty() {
        g.ShowText([]string{"You don't have any journal entries."})
        return
    }
    g.OpenMenu([]util.MenuItem{
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
                g.OpenMenu(g.toJournalMenu(sources))
            },
        },
    })

}

func (g *GridEngine) toJournalMenu(sources []string) []util.MenuItem {
    var items []util.MenuItem
    for _, s := range sources {
        source := s
        items = append(items, util.MenuItem{
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

func (g *GridEngine) IsMapPosOnScreen(pos geometry.Point) bool {
    return g.mapWindow.GetVisibleMap().Contains(pos)
}

func (g *GridEngine) openPartyOverView() {
    g.ShowFixedFormatText(g.playerParty.GetPartyOverview())
}

func (g *GridEngine) DeliverMeleeDamage(attacker *game.Actor, victim *game.Actor) {
    damage := g.rules.GetMeleeDamage(attacker, victim)

    if damage > 0 {
        victim.Damage(g, damage)
        g.Print(fmt.Sprintf("%d dmg. to '%s'", damage, victim.Name()))
    } else {
        g.Print(fmt.Sprintf("No dmg. to '%s'", victim.Name()))
    }

    // hit procs
    attacker.OnMeleeHitPerformed(g, victim)
}

func (g *GridEngine) DeliverRangedDamage(attacker *game.Actor, victim *game.Actor) {
    damage := g.rules.GetRangedDamage(attacker, victim)

    if damage > 0 {
        victim.Damage(g, damage)
        g.Print(fmt.Sprintf("%d dmg. to '%s'", damage, victim.Name()))
    } else {
        g.Print(fmt.Sprintf("No dmg. to '%s'", victim.Name()))
    }
    // hit procs
    attacker.OnRangedHitPerformed(g, victim)
}

func (g *GridEngine) DeliverSpellDamage(attacker *game.Actor, victim *game.Actor, amount int) {
    baseSpellDamage := amount
    victimDefense := victim.GetMagicDefense()
    damage := baseSpellDamage - victimDefense
    if damage > 0 {
        victim.Damage(g, damage)
        g.Print(fmt.Sprintf("%d dmg. to '%s'", damage, victim.Name()))
    } else {
        g.Print(fmt.Sprintf("No dmg. to '%s'", victim.Name()))
    }
}

func (g *GridEngine) openDismissMenu() {
    var menuItems []util.MenuItem
    for _, m := range g.playerParty.GetMembers() {
        if m == g.avatar {
            continue
        }
        member := m
        menuItems = append(menuItems, util.MenuItem{
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
    g.OpenMenu(menuItems)
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

func (g *GridEngine) AddStatusEffect(actor *game.Actor, effect game.StatusEffect, stacks int) {
    for i := 0; i < stacks; i++ {
        actor.AddStatusEffect(g, effect)
    }
    if stacks > 1 {
        g.Print(fmt.Sprintf("%s received %s status x%d", actor.Name(), effect.Name(), stacks))
    } else {
        g.Print(fmt.Sprintf("%s received %s status", actor.Name(), effect.Name()))
    }
}

func (g *GridEngine) AskUserForString(prompt string, maxLength int, onConfirm func(text string)) {
    textInput := g.NewTextInputAtY(10, prompt, func(endedWith ui.EndAction, text string) {
        if endedWith == ui.EndActionConfirm {
            onConfirm(text)
        }
    })
    textInput.SetMaxLength(maxLength)
    textInput.CenterHorizontallyAtY(10)
    g.PushModal(textInput)
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
        return g.createArmorForVendor(level, 10)
    case game.LootWeapon:
        return g.createWeaponsForVendor(level, 10)
    case game.LootHealer:
        return g.createPotions(10)
    case game.LootPotions:
        return g.createPotions(10)
    }
    return []game.Item{}
}

func (g *GridEngine) goBackToBed() {
    g.ShowMultipleChoiceDialogue(g.GetAvatar().Icon(0), g.gridRenderer.AutolayoutArrayToIconPages(5, []string{"You are sure your walls and the mirror will be back to normal, if you just go back to sleep"}), []util.MenuItem{
        {
            Text: "Stay awake",
            Action: func() {
                g.CloseConversation()
            },
        },
        {
            Text: "Go back to bed",
            Action: func() {
                g.CloseConversation()
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

func (g *GridEngine) OpenEquipmentDetails(charIndex int) {
    actor := g.playerParty.GetMember(charIndex)
    g.CloseAllModals()
    equipmentWindow := ui.NewEquipmentWindow(g, actor, g.gridRenderer)
    equipmentWindow.SetActionRight(func() {
        g.openCharMainStats(charIndex)
    })
    equipmentWindow.SetActionLeft(func() {
        g.openCharStatusEffects(charIndex)
    })
    g.PushModal(equipmentWindow)
}
func (g *GridEngine) giveAllSpells() {
    allSpellScrolls := game.GetAllSpellScrolls()
    for _, spellScroll := range allSpellScrolls {
        g.playerParty.AddItem(spellScroll)
    }
}
func (g *GridEngine) giveAllArmorsAndWeapons() {
    allTiers := game.GetAllTiers()
    allWeaponTypes := game.GetAllWeaponTypes()
    allWeaponMaterials := game.GetAllWeaponMaterials()
    allArmorTypes := game.GetAllArmorSlots()
    allArmorMaterials := game.GetAllArmorModifiers()

    for _, tier := range allTiers {
        for _, weaponType := range allWeaponTypes {
            for _, weaponMaterial := range allWeaponMaterials {
                g.playerParty.AddItem(game.NewWeapon(tier, weaponType, weaponMaterial))
            }
        }
        for _, armorType := range allArmorTypes {
            for _, armorMaterial := range allArmorMaterials {
                g.playerParty.AddItem(game.NewArmor(tier, armorType, armorMaterial))
            }
        }
    }
}

func (g *GridEngine) IsInConversation() bool {
    return g.conversationModal != nil && !g.conversationModal.ShouldClose()
}

func (g *GridEngine) CloseConversation() {
    g.conversationModal = nil
}

func (g *GridEngine) IsInCombat() bool {
    return g.combatManager != nil && g.combatManager.IsInCombat()
}

func (g *GridEngine) TryMoveAvatarWithPathfinding(pos geometry.Point) {
    currentMap := g.currentMap
    ourActor := g.GetAvatar()
    currentPath := currentMap.GetJPSPath(ourActor.Pos(), pos, func(point geometry.Point) bool {
        return currentMap.Contains(point) && currentMap.IsWalkableFor(point, ourActor)
    })
    // remove the first element, which is the current position
    if len(currentPath) > 0 {
        currentPath = currentPath[1:]
    }
    if len(currentPath) == 0 {
        return
    }
    runMoveScript := func() {
        g.movementRoutine.Run(func(exe *gocoro.Execution) {
            for _, dest := range currentPath {
                if exe.Stopped() {
                    return
                }
                direction := dest.Sub(g.playerParty.Pos())
                g.PlayerMovement(direction)

                _ = exe.YieldTime(200 * time.Millisecond)
            }
        })
    }

    if g.movementRoutine.Running() {
        //g.movementRoutine.Stop()
        //g.movementRoutine.OnFinish = runMoveScript
        g.movementRoutine = gocoro.NewCoroutine()
        runMoveScript()
    } else {
        runMoveScript()
    }
}

func (g *GridEngine) advanceTimeFromMovement() {
    if g.currentMap.GetName() != "WorldMap" {
        g.AdvanceWorldTime(0, 0, g.rules.GetMinutesPerStepInLevels())
    } else {
        g.AdvanceWorldTime(0, 0, g.playerParty.GetMinutesPerStepOnWorldmap())
    }
}

func (g *GridEngine) advanceTimeFromTicks(ticks uint64) {
    delayInSeconds := 60.0
    actualTPS := ebiten.ActualTPS()
    divider := uint64(actualTPS * delayInSeconds)
    if divider == 0 {
        return
    }
    if ticks%divider == 0 {
        g.AdvanceWorldTime(0, 0, 1)
    }
}

func (g *GridEngine) getWeaponFromEntity(entity *ldtk_go.Entity) game.Item {
    itemTier := game.ItemTier(fromEnum(entity.PropertyByIdentifier("ItemTier").AsString()))
    weaponType := game.WeaponType(fromEnum(entity.PropertyByIdentifier("WeaponType").AsString()))
    weaponMaterial := game.WeaponMaterial(fromEnum(entity.PropertyByIdentifier("WeaponMaterial").AsString()))
    returnedItem := game.NewWeapon(itemTier, weaponType, weaponMaterial)
    if !entity.PropertyByIdentifier("CustomName").IsNull() {
        returnedItem.SetName(entity.PropertyByIdentifier("CustomName").AsString())
    }
    return returnedItem
}

func (g *GridEngine) getArmorFromEntity(entity *ldtk_go.Entity) game.Item {
    itemTier := game.ItemTier(fromEnum(entity.PropertyByIdentifier("ItemTier").AsString()))
    armorType := game.ArmorSlot(fromEnum(entity.PropertyByIdentifier("ArmorType").AsString()))
    armorModifier := game.ArmorModifier(fromEnum(entity.PropertyByIdentifier("ArmorMaterial").AsString()))
    returnedItem := game.NewArmor(itemTier, armorType, armorModifier)
    if !entity.PropertyByIdentifier("CustomName").IsNull() {
        returnedItem.SetName(entity.PropertyByIdentifier("CustomName").AsString())
    }
    return returnedItem
}

func (g *GridEngine) Kill(opponent *game.Actor) {
    if !opponent.IsAlive() {
        return
    }
    opponent.Damage(g, opponent.GetHealth())
    g.actorDied(opponent)
}

func (g *GridEngine) checkPlantHooks(item game.Item, owner *game.Actor) {
    plantHooks := g.levelHooks.PlantHooks
    for i := len(plantHooks) - 1; i >= 0; i-- {
        hook := plantHooks[i]
        if hook.Applies(item, owner) {
            hook.Action(item, owner)
            if hook.Consume {
                g.levelHooks.PlantHooks = append(g.levelHooks.PlantHooks[:i], g.levelHooks.PlantHooks[i+1:]...)
            }
        }
    }
}

func (g *GridEngine) openSkillEditor() {
    var menuItems []util.MenuItem
    for _, skill := range game.GetAllSkillNames() {
        skillName := skill
        currentValue := g.GetAvatar().GetSkills().GetSkillLevel(skillName)
        menuItems = append(menuItems, util.MenuItem{
            Text: fmt.Sprintf("%s (%d)", skillName, currentValue),
            ActionLeft: func() string {
                skills := g.GetAvatar().GetSkills()
                val := skills.GetSkillLevel(skillName)
                if val == 0 {
                    return fmt.Sprintf("%s (0)", skillName)
                }
                skills.DecrementSkill(skillName)
                return fmt.Sprintf("%s (%d)", skillName, skills.GetSkillLevel(skillName))
            },
            ActionRight: func() string {
                skills := g.GetAvatar().GetSkills()
                val := skills.GetSkillLevel(skillName)
                if val == 4 {
                    return fmt.Sprintf("%s (4)", skillName)
                }
                skills.IncrementSkill(skillName)
                return fmt.Sprintf("%s (%d)", skillName, skills.GetSkillLevel(skillName))
            },
        })
    }
    g.OpenMenu(menuItems)
}

func (g *GridEngine) openAttributeEditor() {
    var menuItems []util.MenuItem
    for _, a := range game.GetAllAttributeNames() {
        attributeName := a
        currentValue := g.GetAvatar().GetAttributes().GetAttributeBaseValue(attributeName)
        menuItems = append(menuItems, util.MenuItem{
            Text: fmt.Sprintf("%s (%d)", attributeName, currentValue),
            ActionLeft: func() string {
                attributes := g.GetAvatar().GetAttributes()
                val := attributes.GetAttribute(attributeName)
                if val == 0 {
                    return fmt.Sprintf("%s (0)", attributeName)
                }
                attributes.Decrement(attributeName)
                return fmt.Sprintf("%s (%d)", attributeName, attributes.GetAttribute(attributeName))
            },
            ActionRight: func() string {
                attributes := g.GetAvatar().GetAttributes()
                val := attributes.GetAttribute(attributeName)
                if val == 10 {
                    return fmt.Sprintf("%s (10)", attributeName)
                }
                attributes.Increment(attributeName)
                return fmt.Sprintf("%s (%d)", attributeName, attributes.GetAttribute(attributeName))
            },
        })
    }
    g.OpenMenu(menuItems)
}

func fromEnum(asString string) string {
    asString = strings.ToLower(asString)
    return strings.ReplaceAll(asString, "_", " ")
}

func secondsToTicks(seconds float64) int {
    return int(ebiten.ActualTPS() * seconds)
}
