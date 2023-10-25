package main

import (
    "Legacy/bmpfonts"
    "Legacy/dialogue"
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
    g.playerKnowledge = dialogue.NewPlayerKnowledge()

    g.loadMap("WorldMap")
    g.PlaceParty(g.spawnPosition)

    //pk := dialogue.NewPlayerKnowledge()

    /*
       dialogueFilename := path.Join("assets", "dialogue", "tim.txt")
       dialogueFile := mustOpen(dialogueFilename)
       loadedDialogue := dialogue.NewDialogueFromFile(dialogueFile)
       options := loadedDialogue.GetOptions(pk)
       response, _ := loadedDialogue.GetResponseAndAddKnowledge(pk, options[0])
       println(response)

    */

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
        },
    })
}

type Transition struct {
    TargetMap string
    TargetPos geometry.Point
}

func (g *GridEngine) loadMap(identifier string) {
    worldTileset := g.ldtkMapProject.TilesetByIdentifier("World")

    currentMap := g.ldtkMapProject.LevelByIdentifier(identifier)
    environmentLayer := currentMap.LayerByIdentifier("Environment")
    transitionLayer := currentMap.LayerByIdentifier("Transitions")
    itemLayer := currentMap.LayerByIdentifier("Items")
    objectLayer := currentMap.LayerByIdentifier("Objects")
    npcLayer := currentMap.LayerByIdentifier("NPCs")

    g.transitionMap = make(map[geometry.Point]Transition)

    for _, transitionEntity := range transitionLayer.Entities {
        posX, posY := transitionLayer.ToGridPosition(transitionEntity.Position[0], transitionEntity.Position[1])
        gridPos := geometry.Point{X: posX, Y: posY}
        destPosArray := transitionEntity.PropertyByIdentifier("DestinationPosition").AsArray()
        mapName := transitionEntity.PropertyByIdentifier("MapName").AsString()
        transitionPos := geometry.Point{
            X: int(destPosArray[0].(float64)),
            Y: int(destPosArray[1].(float64)),
        }

        if mapName == "_player_spawn_" {
            g.spawnPosition = gridPos
        } else {
            g.transitionMap[gridPos] = Transition{
                TargetMap: mapName,
                TargetPos: transitionPos,
            }
        }
    }

    g.gridmap = gridmap.NewEmptyMap[*game.Actor, game.Item, game.Object](environmentLayer.CellWidth, environmentLayer.CellHeight, 9)

    for _, tile := range environmentLayer.Tiles {
        posX, posY := environmentLayer.ToGridPosition(tile.Position[0], tile.Position[1])
        pos := geometry.Point{X: posX, Y: posY}
        enums := worldTileset.EnumsForTile(tile.ID)
        g.gridmap.SetCell(pos, gridmap.MapCell[*game.Actor, game.Item, game.Object]{
            TileType: gridmap.Tile{
                DefinedIcon:   tile.ID,
                IsWalkable:    !enums.Contains("Wall") && !enums.Contains("Water"),
                IsTransparent: !enums.Contains("Wall"),
            },
        })
    }

    for _, entity := range objectLayer.Entities {
        posX, posY := objectLayer.ToGridPosition(entity.Position[0], entity.Position[1])
        pos := geometry.Point{X: posX, Y: posY}
        mapObject := g.getObjectFromEntity(entity)
        g.gridmap.AddObject(mapObject, pos)
    }

    for _, entity := range itemLayer.Entities {
        posX, posY := itemLayer.ToGridPosition(entity.Position[0], entity.Position[1])
        pos := geometry.Point{X: posX, Y: posY}
        item := g.getItemFromEntity(entity)
        if item != nil {
            g.gridmap.AddItem(item, pos)
        }
    }

    for _, entity := range npcLayer.Entities {
        posX, posY := npcLayer.ToGridPosition(entity.Position[0], entity.Position[1])
        pos := geometry.Point{X: posX, Y: posY}
        name := entity.PropertyByIdentifier("Name").AsString()
        npcTile := entity.PropertyByIdentifier("Icon").Value.(map[string]interface{})
        tileset := g.ldtkMapProject.TilesetByUID(int(npcTile["tilesetUid"].(float64)))
        tilesetWidth := tileset.Width / tileset.GridSize
        atlasX := int(npcTile["x"].(float64) / float64(tileset.GridSize))
        atlasY := int(npcTile["y"].(float64) / float64(tileset.GridSize))
        textureIndex := atlasY*tilesetWidth + atlasX
        npc := game.NewActor(name, textureIndex)
        g.gridmap.AddActor(npc, pos)
    }

    g.mapWindow = renderer.NewMapWindow(
        geometry.Point{X: 8, Y: 8},
        geometry.Point{X: 19, Y: 11},
        geometry.Point{X: environmentLayer.CellWidth, Y: environmentLayer.CellHeight},
        g.mapLookup,
    )
    g.mapRenderer = renderer.NewRenderer(g.gridRenderer, g.mapWindow)
}

func (g *GridEngine) PlaceParty(startPos geometry.Point) {
    g.playerParty.SetGridMap(g.gridmap)
    g.gridmap.AddActor(g.avatar, startPos)

    if g.playerParty.HasFollowers() {
        followerCount := len(g.playerParty.GetMembers()) - 1
        freeCells := g.gridmap.GetFreeCellsForDistribution(startPos, followerCount, func(p geometry.Point) bool {
            return g.gridmap.Contains(p) && g.gridmap.IsCurrentlyPassable(p)
        })
        if len(freeCells) < followerCount {
            println(fmt.Sprintf("ERROR: not enough free cells for followers at %v", startPos))
        } else {
            for i, follower := range g.playerParty.GetMembers()[1:] {
                followerPos := freeCells[i]
                g.gridmap.AddActor(follower, followerPos)
            }
        }
    }

    g.onPartyMoved()
}

func (g *GridEngine) getItemFromEntity(entity *ldtk_go.Entity) game.Item {
    switch entity.Identifier {
    case "Book":
        return g.getBookFromEntity(entity)
    case "Key":
        return g.getKeyFromEntity(entity)
    }
    return nil
}

func (g *GridEngine) getObjectFromEntity(entity *ldtk_go.Entity) game.Object {
    switch entity.Identifier {
    case "Door":
        return game.NewDoor()
    case "Locked_Door":
        return g.getLockedDoorFromEntity(entity)
    }
    return nil
}

func (g *GridEngine) getLockedDoorFromEntity(entity *ldtk_go.Entity) game.Object {
    key := entity.PropertyByIdentifier("Key").AsString()
    strength := entity.PropertyByIdentifier("Strength").AsFloat64()
    door := game.NewLockedDoor(key, strength)
    return door
}
func (g *GridEngine) getBookFromEntity(entity *ldtk_go.Entity) game.Item {
    title := entity.PropertyByIdentifier("Title").AsString()
    filename := entity.PropertyByIdentifier("Filename").AsString()
    bookPath := path.Join("assets", "books", filename)
    book := game.NewBook(title, bookPath)
    return book
}

func (g *GridEngine) getKeyFromEntity(entity *ldtk_go.Entity) game.Item {
    name := entity.PropertyByIdentifier("Name").AsString()
    key := entity.PropertyByIdentifier("Key").AsString()
    importance := entity.PropertyByIdentifier("Importance").AsInt()
    return game.NewKey(name, key, keyColor(importance))
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
