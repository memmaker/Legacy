package main

import (
    "Legacy/bmpfonts"
    "Legacy/ega"
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/gridmap"
    "Legacy/ldtk_go"
    "Legacy/renderer"
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
    "path"
    "strings"
)

func (g *GridEngine) Init() {
    g.deviceDPIScale = ebiten.DeviceScaleFactor()

    g.gridRenderer = renderer.NewDualGridRenderer(g.uiTiles, g.entityTiles, g.TotalScale(), g.getFontIndex())

    g.gridRenderer.SetBorderDefinition(renderer.GridBorderDefinition{
        HorizontalLineTextureIndex: 13,
        VerticalLineTextureIndex:   128,
        CornerTextureIndex:         2,
        BackgroundTextureIndex:     32,
    })

    g.ldtkMapProject, _ = ldtk_go.Open("assets/Legacy.ldtk")

    g.loadUIOverlay()

    g.avatar = game.NewActor("Avatar", 7)
    g.playerParty = game.NewParty(g.avatar)
    g.playerKnowledge = game.NewPlayerKnowledge()
    g.flags = game.NewFlags()

    g.currentMap = g.loadMap("WorldMap")
    g.PlaceParty(g.spawnPosition)
}

func (g *GridEngine) loadUIOverlay() {
    screenW := 40 // in 8x8 cells
    uiMap := g.ldtkMapProject.LevelByIdentifier("UI_Overlay")
    uiLayer := uiMap.LayerByIdentifier("UI_Layer")
    for _, tile := range uiLayer.Tiles {
        cellX, cellY := uiLayer.ToGridPosition(tile.Position[0], tile.Position[1])
        cellIndex := cellY*screenW + cellX
        g.uiOverlay[cellIndex] = tile.ID
    }
}

type TextConfig struct {
    fontIndex map[rune]uint16
    atlas     *ebiten.Image
    scale     float64
    maxLength int
}

func (t TextConfig) GetMaxLength() int {
    return t.maxLength
}

func (t TextConfig) GetTileSize() (int, int) {
    return 8, 8
}

func (t TextConfig) GetScreenOffset() geometry.Point {
    return geometry.Point{X: 0, Y: 0}
}

func (t TextConfig) GetTextureIndexFor(c rune) (*ebiten.Image, int) {
    if _, ok := t.fontIndex[c]; !ok {
        return nil, -1
    }
    return t.atlas, int(t.fontIndex[c])
}

func (t TextConfig) GetScale() (float64, float64) {
    return t.scale, t.scale
}

func (g *GridEngine) getFontIndex() map[rune]uint16 {
    indexOfSmallA := int(97)
    indexOfZero := int(48)
    return bmpfonts.NewIndexFromDescription(bmpfonts.AtlasDescription{
        IndexOfCapitalA: 65,
        IndexOfSmallA:   &indexOfSmallA,
        IndexOfZero:     &indexOfZero,
        Chains: []bmpfonts.SpecialCharacterChain{
            {StartIndex: 32, Characters: []rune{' ', '!', '"', '#', '$', '%', '&', '’', '(', ')', '*', '+', ',', '-', '.', '/'}},
            {StartIndex: 58, Characters: []rune{':', ';', '<', '=', '>', '?', '@'}},
            {StartIndex: 91, Characters: []rune{'[', '\\', ']', '^', '_', '`'}},
            {StartIndex: 123, Characters: []rune{'{', '|', '}', '~', '^'}},
            {StartIndex: 1, Characters: []rune{'ӽ'}},  // red gem
            {StartIndex: 15, Characters: []rune{'Ө'}}, // blue gem
            {StartIndex: 8, Characters: []rune{'ө'}},  // grey gem
            {StartIndex: 3, Characters: []rune{'█'}},  // grey block
            {StartIndex: 45, Characters: []rune{'—'}},
            {StartIndex: 39, Characters: []rune{'\''}},
            {StartIndex: 131, Characters: []rune{'🍗', '🪙'}}, // food/gold
        },
    })
}

func (g *GridEngine) loadMap(mapName string) *gridmap.GridMap[*game.Actor, game.Item, game.Object] {
    worldTileset := g.ldtkMapProject.TilesetByIdentifier("World")

    currentMap := g.ldtkMapProject.LevelByIdentifier(mapName)
    environmentLayer := currentMap.LayerByIdentifier("Environment")
    metaLayer := currentMap.LayerByIdentifier("Meta")
    itemLayer := currentMap.LayerByIdentifier("Items")
    objectLayer := currentMap.LayerByIdentifier("Objects")
    npcLayer := currentMap.LayerByIdentifier("NPCs")

    loadedMap := gridmap.NewEmptyMap[*game.Actor, game.Item, game.Object](environmentLayer.CellWidth, environmentLayer.CellHeight, 9)
    loadedMap.SetName(mapName)
    for _, metaEntity := range metaLayer.Entities {
        posX, posY := metaLayer.ToGridPosition(metaEntity.Position[0], metaEntity.Position[1])
        gridPos := geometry.Point{X: posX, Y: posY}

        g.handleMetaEntity(loadedMap, metaEntity, gridPos)
    }

    for _, tile := range environmentLayer.Tiles {
        posX, posY := environmentLayer.ToGridPosition(tile.Position[0], tile.Position[1])
        pos := geometry.Point{X: posX, Y: posY}
        enums := worldTileset.EnumsForTile(tile.ID)
        loadedMap.SetCell(pos, gridmap.MapCell[*game.Actor, game.Item, game.Object]{
            TileType: gridmap.Tile{
                DefinedIcon:   tile.ID,
                IsWalkable:    !enums.Contains("IsBlockingMovement"),
                IsTransparent: !enums.Contains("IsBlockingView"),
            },
        })
    }

    for _, entity := range objectLayer.Entities {
        posX, posY := objectLayer.ToGridPosition(entity.Position[0], entity.Position[1])
        pos := geometry.Point{X: posX, Y: posY}
        mapObject := g.getObjectFromEntity(entity)
        if mapObject != nil {
            loadedMap.AddObject(mapObject, pos)
        } else {
            println(fmt.Sprintf("ERROR: could not create object from entity %v", entity))
        }
    }

    for _, entity := range itemLayer.Entities {
        posX, posY := itemLayer.ToGridPosition(entity.Position[0], entity.Position[1])
        pos := geometry.Point{X: posX, Y: posY}
        item := g.getItemFromEntity(entity)
        if item != nil {
            loadedMap.AddItem(item, pos)
        } else {
            println(fmt.Sprintf("ERROR: could not create item from entity %v", entity))
        }
    }

    for _, entity := range npcLayer.Entities {
        posX, posY := npcLayer.ToGridPosition(entity.Position[0], entity.Position[1])
        pos := geometry.Point{X: posX, Y: posY}
        name := entity.PropertyByIdentifier("Name").AsString()
        npcTile := entity.PropertyByIdentifier("Icon").Value.(map[string]interface{})
        isHidden := entity.PropertyByIdentifier("IsHidden").AsBool()
        iconFrames := entity.PropertyByIdentifier("IconFrames").AsInt()

        tileset := g.ldtkMapProject.TilesetByUID(int(npcTile["tilesetUid"].(float64)))
        tilesetWidth := tileset.Width / tileset.GridSize
        atlasX := int(npcTile["x"].(float64) / float64(tileset.GridSize))
        atlasY := int(npcTile["y"].(float64) / float64(tileset.GridSize))
        textureIndex := atlasY*tilesetWidth + atlasX
        npcFilename := path.Join("assets", "npc", name+".txt")
        // NOVA
        if g.NovaPlays() && name == "hungry_caterpillar" {
            textureIndex = 188
        }
        var npc *game.Actor
        if doesFileExist(npcFilename) {
            npc = game.NewActorFromFile(mustOpen(npcFilename), textureIndex)
        } else {
            npc = game.NewActor(name, textureIndex)
        }
        npc.SetIconFrames(iconFrames)
        npc.SetInternalName(name)
        npc.SetHidden(isHidden)
        loadedMap.AddActor(npc, pos)
    }

    mapWidth := environmentLayer.CellWidth
    mapHeight := environmentLayer.CellHeight

    g.initMapWindow(mapWidth, mapHeight)

    return loadedMap
}

func (g *GridEngine) initMapWindow(mapWidth int, mapHeight int) {
    g.mapWindow = renderer.NewMapWindow(
        geometry.Point{X: 8, Y: 8},
        geometry.Point{X: 19, Y: 11},
        geometry.Point{X: mapWidth, Y: mapHeight},
        g.mapLookup,
    )
    g.mapRenderer = renderer.NewRenderer(g.gridRenderer, g.mapWindow)
}

func (g *GridEngine) handleMetaEntity(loadedMap *gridmap.GridMap[*game.Actor, game.Item, game.Object], entity *ldtk_go.Entity, gridPos geometry.Point) {
    switch entity.Identifier {
    case "Transition":
        g.handleTransition(loadedMap, entity, gridPos)
    case "Secret_Door":
        loadedMap.SetSecretDoorAt(gridPos)
    }
}

func (g *GridEngine) handleTransition(loadedMap *gridmap.GridMap[*game.Actor, game.Item, game.Object], metaEntity *ldtk_go.Entity, gridPos geometry.Point) {
    destPosArray := metaEntity.PropertyByIdentifier("DestinationPosition").AsArray()
    mapName := metaEntity.PropertyByIdentifier("MapName").AsString()
    transitionPos := geometry.Point{
        X: int(destPosArray[0].(float64)),
        Y: int(destPosArray[1].(float64)),
    }

    if mapName == "_player_spawn_" {
        g.spawnPosition = gridPos
    } else {
        loadedMap.SetTransitionAt(gridPos, gridmap.Transition{
            TargetMap: mapName,
            TargetPos: transitionPos,
        })
    }
}

func (g *GridEngine) NovaPlays() bool {
    return strings.ToLower(g.avatar.Name()) == "nova"
}
func (g *GridEngine) RemovePartyFromMap(loadedMap *gridmap.GridMap[*game.Actor, game.Item, game.Object]) {
    for _, actor := range g.playerParty.GetMembers() {
        loadedMap.RemoveActor(actor)
    }
}
func (g *GridEngine) PlaceParty(startPos geometry.Point) {
    g.playerParty.SetGridMap(g.currentMap)
    g.currentMap.AddActor(g.avatar, startPos)

    if g.playerParty.HasFollowers() {
        followerCount := len(g.playerParty.GetMembers()) - 1
        freeCells := g.currentMap.GetFreeCellsForDistribution(startPos, followerCount, func(p geometry.Point) bool {
            return g.currentMap.Contains(p) && g.currentMap.IsCurrentlyPassable(p)
        })
        if len(freeCells) < followerCount {
            println(fmt.Sprintf("ERROR: not enough free cells for followers at %v", startPos))
        } else {
            for i, follower := range g.playerParty.GetMembers()[1:] {
                followerPos := freeCells[i]
                g.currentMap.AddActor(follower, followerPos)
            }
        }
    }

    g.onPartyMoved()
}

func (g *GridEngine) getItemFromEntity(entity *ldtk_go.Entity) game.Item {
    switch entity.Identifier {
    case "Scroll":
        return g.getScrollFromEntity(entity)
    case "Key":
        return g.getKeyFromEntity(entity)
    case "Candle":
        return game.NewCandle(entity.PropertyByIdentifier("IsLit").AsBool())
    case "Potion":
        return game.NewPotion()
    }
    return nil
}

func (g *GridEngine) getObjectFromEntity(entity *ldtk_go.Entity) game.Object {
    switch entity.Identifier {
    case "Door":
        return game.NewDoor()
    case "Locked_Door":
        return g.getLockedDoorFromEntity(entity)
    case "Magically_Locked_Door":
        return g.getMagicallyLockedDoorFromEntity(entity)
    case "Shrine":
        return g.getShrineFromEntity(entity)
    case "Chest":
        return g.getChestFromEntity(entity)
    case "Fireplace":
        return g.getFireplaceFromEntity(entity)
    }
    return nil
}

func (g *GridEngine) getLockedDoorFromEntity(entity *ldtk_go.Entity) game.Object {
    key := entity.PropertyByIdentifier("Key").AsString()
    strength := entity.PropertyByIdentifier("Strength").AsFloat64()
    door := game.NewLockedDoor(key, strength)
    return door
}

func (g *GridEngine) getMagicallyLockedDoorFromEntity(entity *ldtk_go.Entity) game.Object {
    strength := entity.PropertyByIdentifier("Strength").AsFloat64()
    door := game.NewMagicallyLockedDoor(strength)
    return door
}
func (g *GridEngine) getScrollFromEntity(entity *ldtk_go.Entity) game.Item {
    title := entity.PropertyByIdentifier("Title").AsString()
    filename := entity.PropertyByIdentifier("Filename").AsString()
    isHidden := entity.PropertyByIdentifier("IsHidden").AsBool()
    spellName := entity.PropertyByIdentifier("SpellName").AsString()

    bookPath := path.Join("assets", "scrolls", filename)
    scroll := game.NewScroll(title, bookPath)
    scroll.SetHidden(isHidden)
    if spellName != "" {
        spell := game.NewSpellFromName(spellName)
        if spell != nil {
            scroll.SetSpell(spell)
        }
    }
    return scroll
}

func (g *GridEngine) getKeyFromEntity(entity *ldtk_go.Entity) game.Item {
    name := entity.PropertyByIdentifier("Name").AsString()
    key := entity.PropertyByIdentifier("Key").AsString()
    importance := entity.PropertyByIdentifier("Importance").AsInt()
    isHidden := entity.PropertyByIdentifier("IsHidden").AsBool()
    newKey := game.NewKey(name, key, keyColor(importance))
    newKey.SetHidden(isHidden)
    return newKey
}

func (g *GridEngine) getShrineFromEntity(entity *ldtk_go.Entity) game.Object {
    name := entity.PropertyByIdentifier("Name").AsString()
    principle := entity.PropertyByIdentifier("Principle").Value.(string)
    return game.NewShrine(name, game.Principle(principle))
}

func (g *GridEngine) getChestFromEntity(entity *ldtk_go.Entity) game.Object {
    needsKey := ""
    needsKeyValue := entity.PropertyByIdentifier("NeedsKey").Value
    if needsKeyValue != nil {
        needsKey = needsKeyValue.(string)
    }
    isHidden := entity.PropertyByIdentifier("IsHidden").AsBool()
    lootLevel := entity.PropertyByIdentifier("LootLevel").AsInt()
    lootType := entity.PropertyByIdentifier("LootType").Value.(string)
    chest := game.NewChest(lootLevel, game.Loot(lootType))
    chest.SetLockedWithKey(needsKey)
    chest.SetHidden(isHidden)
    return chest
}

func (g *GridEngine) getFireplaceFromEntity(entity *ldtk_go.Entity) game.Object {
    foodCount := entity.PropertyByIdentifier("FoodCount").AsInt()
    return game.NewFireplace(foodCount)
}

func keyColor(importance int) color.Color {
    switch importance {
    case 1:
        return ega.BrightBlack
    case 2:
        return ega.White
    case 3:
        return ega.BrightWhite
    case 4:
        return ega.Yellow
    case 5:
        return ega.BrightYellow
    }
    return ega.White
}
