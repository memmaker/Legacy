package main

import (
    "Legacy/bmpfonts"
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/gridmap"
    "Legacy/ldtk_go"
    "Legacy/renderer"
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "path"
    "strings"
)

func (g *GridEngine) Init() {
    g.deviceDPIScale = ebiten.DeviceScaleFactor()

    g.gridRenderer = renderer.NewDualGridRenderer(g.TotalScale(), getCharFontIndex())
    g.gridRenderer.SetFontIndexForBigGrid(getWorldFontIndex())
    g.gridRenderer.SetAtlasMap(
        map[renderer.AtlasName]*ebiten.Image{
            renderer.AtlasCharacters:        g.uiTiles,
            renderer.AtlasWorld:             g.worldTiles,
            renderer.AtlasEntities:          g.entityTiles,
            renderer.AtlasEntitiesGrayscale: g.grayScaleEntityTiles,
        },
    )

    g.gridRenderer.SetBorderDefinition(renderer.GridBorderDefinition{
        HorizontalLineTextureIndex: 13,
        VerticalLineTextureIndex:   128,
        CornerTextureIndex:         2,
        BackgroundTextureIndex:     32,
    })

    g.combatManager = NewCombatState(g)
    g.rules = game.NewRules()

    g.ldtkMapProject, _ = ldtk_go.Open("assets/Legacy.ldtk")
    g.loadPartyIcons()

    g.loadUIOverlay()

    smallScreenSize := g.gridRenderer.GetSmallGridScreenSize()
    g.foodButton = geometry.Point{X: 1, Y: smallScreenSize.Y - 2}
    g.bagButton = geometry.Point{X: smallScreenSize.X - 2, Y: smallScreenSize.Y - 2}
    g.goldButton = geometry.Point{X: smallScreenSize.X - 4, Y: smallScreenSize.Y - 2}

    g.defenseBuffsButton = geometry.Point{X: 15, Y: 0}
    g.offenseBuffsButton = geometry.Point{X: 24, Y: 0}

    g.avatar = game.NewActor("---", 7)
    g.playerParty = game.NewParty(g.avatar)
    g.playerParty.InitWithRules(g.rules)
    g.playerKnowledge = game.NewPlayerKnowledge()
    g.flags = game.NewFlags()

    //g.currentMap = g.loadMap("WorldMap")
    g.currentMap = g.loadMap("Bed_Room")
    g.initMapWindow(g.currentMap.MapWidth, g.currentMap.MapHeight)
    //g.currentMap = g.loadMap("Tauci_Castle")
    g.PlaceParty(g.spawnPosition)
}

func (g *GridEngine) loadUIOverlay() {
    screenW := 40 // in 8x8 cells
    uiMap := g.ldtkMapProject.LevelByIdentifier("UI_Overlay")
    uiLayer := uiMap.LayerByIdentifier("UI_Layer")
    for _, tile := range uiLayer.Tiles {
        cellX, cellY := uiLayer.ToGridPosition(tile.Position[0], tile.Position[1])
        cellIndex := cellY*screenW + cellX
        g.uiOverlay[cellIndex] = int32(tile.ID)
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

func getCharFontIndex() map[rune]uint16 {
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
            {StartIndex: 131, Characters: []rune{'🍗', '🪙'}},                          // food/gold
            {StartIndex: 147, Characters: []rune{'Ⅰ', 'Ⅱ', 'Ⅲ', 'Ⅳ'}},                // roman numerals 1-4
            {StartIndex: 153, Characters: []rune{'Ä', 'Ö', 'Ü', 'ß', 'ä', 'ö', 'ü'}}, // umlaute
        },
    })
}
func getWorldFontIndex() map[rune]uint16 {
    return bmpfonts.NewIndexFromDescription(bmpfonts.AtlasDescription{
        IndexOfCapitalA: 49,
        Chains: []bmpfonts.SpecialCharacterChain{
            {StartIndex: 80, Characters: []rune{' '}},
            {StartIndex: 80, Characters: []rune{'+'}},
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
    zoneLayer := currentMap.LayerByIdentifier("Zones")

    loadedMap := gridmap.NewEmptyMap[*game.Actor, game.Item, game.Object](environmentLayer.CellWidth, environmentLayer.CellHeight, 9)
    loadedMap.SetName(mapName)

    for _, metaEntity := range metaLayer.Entities {
        posX, posY := metaLayer.ToGridPosition(metaEntity.Position[0], metaEntity.Position[1])
        gridPos := geometry.Point{X: posX, Y: posY}

        g.handleMetaEntity(loadedMap, metaEntity, gridPos)
    }

    for _, entity := range zoneLayer.Entities {
        posX, posY := zoneLayer.ToGridPosition(entity.Position[0], entity.Position[1])
        gridPos := geometry.Point{X: posX, Y: posY}
        name := entity.PropertyByIdentifier("Name").AsString()
        widthInCells, heightInCells := zoneLayer.ToGridPosition(entity.Width, entity.Height)
        zoneRect := geometry.NewRect(gridPos.X, gridPos.Y, gridPos.X+widthInCells, gridPos.Y+heightInCells)
        loadedMap.AddNamedRegion(name, zoneRect)

        enterTriggerProp := entity.PropertyByIdentifier("TriggerOnEnter")
        if enterTriggerProp.IsNull() {
            continue
        }
        eventName := enterTriggerProp.AsString()
        g.subscribeToMovement(func(pos geometry.Point) bool {
            if g.GetMapName() == mapName && zoneRect.Contains(pos) {
                g.TriggerEvent(eventName)
                return true
            }
            return false
        })
    }

    for _, tile := range environmentLayer.Tiles {
        posX, posY := environmentLayer.ToGridPosition(tile.Position[0], tile.Position[1])
        pos := geometry.Point{X: posX, Y: posY}
        enums := worldTileset.EnumsForTile(tile.ID)

        walkable := !enums.Contains("IsBlockingMovement")
        transparent := !enums.Contains("IsBlockingView")

        specialTile := gridmap.SpecialTileNone
        if enums.Contains("IsForest") && mapName == "WorldMap" {
            specialTile = gridmap.SpecialTileForest
        } else if enums.Contains("IsMountain") {
            if mapName == "WorldMap" {
                specialTile = gridmap.SpecialTileMountain
            } else {
                walkable = false
                transparent = false
            }
        } else if enums.Contains("IsSwamp") && mapName == "WorldMap" {
            specialTile = gridmap.SpecialTileSwamp
        } else if enums.Contains("IsWater") {
            specialTile = gridmap.SpecialTileWater
        } else if enums.Contains("IsBed") {
            specialTile = gridmap.SpecialTileBed
        } else if enums.Contains("IsBreakable") {
            if tile.ID == 108 {
                specialTile = gridmap.SpecialTileBreakableGems
            } else if tile.ID == 109 {
                specialTile = gridmap.SpecialTileBreakableGold
            } else if tile.ID >= 75 || tile.ID <= 79 {
                specialTile = gridmap.SpecialTileBreakableGlass
            } else {
                specialTile = gridmap.SpecialTileBreakable
            }
        }
        loadedMap.SetCell(pos, gridmap.MapCell[*game.Actor, game.Item, game.Object]{
            TileType: gridmap.Tile{
                DefinedIcon:   int32(tile.ID),
                IsWalkable:    walkable,
                IsTransparent: transparent,
                Special:       specialTile,
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

        var discoveryMessage []string
        if !entity.PropertyByIdentifier("OnDiscovery").IsNull() {
            discoveryMessage = strings.Split(entity.PropertyByIdentifier("OnDiscovery").AsString(), "\n")
        }

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
            npc = game.NewActorFromFile(mustOpen(npcFilename), int32(textureIndex), g.gridRenderer.AutolayoutArrayToIconPages)
        } else {
            npc = game.NewActor(name, int32(textureIndex))
        }
        npc.SetIconFrames(iconFrames)
        npc.SetInternalName(name)
        npc.SetDiscoveryMessage(isHidden, discoveryMessage)

        // vendor inventory
        vendorProp := entity.PropertyByIdentifier("AddVendorInventory")
        vendorItemLevelProp := entity.PropertyByIdentifier("VendorItemLevel")

        if !vendorProp.IsNull() && !vendorItemLevelProp.IsNull() {
            lootType := game.Loot(strings.ToLower(vendorProp.Value.(string)))
            lootLevel := vendorItemLevelProp.AsInt()
            vendorItems := g.CreateItemsForVendor(lootType, lootLevel)
            npc.SetVendorInventory(vendorItems)
        }

        npcLevel := entity.PropertyByIdentifier("Level").AsInt()
        if npcLevel > npc.GetLevel() {
            for npc.GetLevel() < npcLevel {
                g.rules.LevelUp(npc)
            }
        }

        loadedMap.AddActor(npc, pos)
    }

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
    nameOfLocation := metaEntity.PropertyByIdentifier("NameOfLocation").AsString()
    loadedMap.SetNamedLocation(nameOfLocation, gridPos)

    if nameOfLocation == "player_spawn" {
        g.spawnPosition = gridPos
    } else {
        targetProp := metaEntity.PropertyByIdentifier("Target")
        if targetProp.IsNull() {
            return
        }
        ref := targetProp.Value.(map[string]interface{})
        levelId := ref["levelIid"].(string)
        entityId := ref["entityIid"].(string)
        targetLevel := g.ldtkMapProject.LevelByIID(levelId)
        refEntity := g.ldtkMapProject.EntityByIID(entityId)

        mapName := targetLevel.Identifier
        destLoc := refEntity.PropertyByIdentifier("NameOfLocation").AsString()

        if destLoc != "" && mapName != "" {
            loadedMap.SetTransitionAt(gridPos, gridmap.Transition{
                TargetMap:      mapName,
                TargetLocation: destLoc,
            })
        }
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
    g.onViewedActorMoved(startPos)
}
func (g *GridEngine) PlacePartyBackOnCurrentMap() {
    g.playerParty.SetGridMap(g.currentMap)
    for _, member := range g.playerParty.GetMembers() {
        g.currentMap.AddActor(member, member.Pos())
    }
    g.onViewedActorMoved(g.playerParty.GetMember(0).Pos())
}
func (g *GridEngine) getItemFromEntity(entity *ldtk_go.Entity) game.Item {
    var returnedItem game.Item
    switch entity.Identifier {
    case "Scroll":
        returnedItem = g.getScrollFromEntity(entity)
    case "Key":
        returnedItem = g.getKeyFromEntity(entity)
    case "Candle":
        returnedItem = game.NewCandle(entity.PropertyByIdentifier("IsLit").AsBool())
    case "Left_Torch":
        returnedItem = game.NewLeftTorch(entity.PropertyByIdentifier("IsLit").AsBool())
    case "Right_Torch":
        returnedItem = game.NewRightTorch(entity.PropertyByIdentifier("IsLit").AsBool())
    case "Potion":
        returnedItem = game.NewPotion()
    }

    pickupTriggerProp := entity.PropertyByIdentifier("TriggerOnPickup")
    if pickupTriggerProp != nil && !pickupTriggerProp.IsNull() {
        pickupEventName := pickupTriggerProp.AsString()
        returnedItem.SetPickupEvent(pickupEventName)
    }

    return returnedItem
}

func (g *GridEngine) getObjectFromEntity(entity *ldtk_go.Entity) game.Object {
    switch entity.Identifier {
    case "Door":
        return g.getDoorFromEntity(entity)
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
    case "Mirror":
        return game.NewMirror(entity.PropertyByIdentifier("IsMagical").AsBool(), entity.PropertyByIdentifier("IsBroken").AsBool())
    case "Tombstone":
        return g.getTombstoneFromEntity(entity)
    }
    return nil
}

func (g *GridEngine) getDoorFromEntity(entity *ldtk_go.Entity) *game.Door {
    var onListenText []string
    if !entity.PropertyByIdentifier("OnListenText").IsNull() {
        onListenText = strings.Split(entity.PropertyByIdentifier("OnListenText").AsString(), "\n")
    }
    door := game.NewDoor()
    door.SetListenText(onListenText)
    return door
}

func (g *GridEngine) getLockedDoorFromEntity(entity *ldtk_go.Entity) game.Object {
    key := entity.PropertyByIdentifier("Key").AsString()
    strength := entity.PropertyByIdentifier("Strength").AsFloat64()
    var onBreakEvent string
    if !entity.PropertyByIdentifier("OnBreakEvent").IsNull() {
        onBreakEvent = entity.PropertyByIdentifier("OnBreakEvent").AsString()
    }
    var onListenText []string
    if !entity.PropertyByIdentifier("OnListenText").IsNull() {
        onListenText = strings.Split(entity.PropertyByIdentifier("OnListenText").AsString(), "\n")
    }
    door := game.NewLockedDoor(key, strength)
    door.SetBreakEvent(onBreakEvent)
    door.SetListenText(onListenText)
    return door
}

func (g *GridEngine) getMagicallyLockedDoorFromEntity(entity *ldtk_go.Entity) game.Object {
    var onListenText []string
    if !entity.PropertyByIdentifier("OnListenText").IsNull() {
        onListenText = strings.Split(entity.PropertyByIdentifier("OnListenText").AsString(), "\n")
    }
    strength := entity.PropertyByIdentifier("Strength").AsFloat64()
    door := game.NewMagicallyLockedDoor(strength)
    door.SetListenText(onListenText)
    return door
}
func (g *GridEngine) getScrollFromEntity(entity *ldtk_go.Entity) game.Item {
    title := entity.PropertyByIdentifier("Title").AsString()
    filename := entity.PropertyByIdentifier("Filename").AsString()
    isHidden := entity.PropertyByIdentifier("IsHidden").AsBool()
    var spellName string
    spellProp := entity.PropertyByIdentifier("SpellName")
    if spellProp != nil && !spellProp.IsNull() {
        spellName = spellProp.AsString()
    }
    var discoveryMessage []string
    if !entity.PropertyByIdentifier("OnDiscovery").IsNull() {
        discoveryMessage = strings.Split(entity.PropertyByIdentifier("OnDiscovery").AsString(), "\n")
    }

    scroll := game.NewScroll(title, filename)
    scroll.SetDiscoveryMessage(isHidden, discoveryMessage)
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

    var discoveryMessage []string
    if !entity.PropertyByIdentifier("OnDiscovery").IsNull() {
        discoveryMessage = strings.Split(entity.PropertyByIdentifier("OnDiscovery").AsString(), "\n")
    }

    newKey := game.NewKeyFromImportance(name, key, importance)
    newKey.SetDiscoveryMessage(isHidden, discoveryMessage)
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
    lootTypeList := entity.PropertyByIdentifier("LootType").Value

    var lootList []game.Loot
    for _, lootTypeRaw := range lootTypeList.([]interface{}) {
        lootType := game.Loot(strings.ToLower(lootTypeRaw.(string)))
        lootList = append(lootList, lootType)
    }

    var fixedLoot []game.Item
    if !entity.PropertyByIdentifier("FixedLoot").IsNull() {
        for _, fixedLootRaw := range entity.PropertyByIdentifier("FixedLoot").Value.([]interface{}) {
            itemString := fixedLootRaw.(string)
            fixedLoot = append(fixedLoot, game.NewItemFromString(itemString))
        }
    }

    var discoveryMessage []string
    if !entity.PropertyByIdentifier("OnDiscovery").IsNull() {
        discoveryMessage = strings.Split(entity.PropertyByIdentifier("OnDiscovery").AsString(), "\n")
    }
    chest := game.NewChest(lootLevel, lootList)
    chest.SetLockedWithKey(needsKey)
    chest.SetDiscoveryMessage(isHidden, discoveryMessage)
    if len(fixedLoot) > 0 {
        chest.SetFixedLoot(fixedLoot)
    }
    return chest
}

func (g *GridEngine) getFireplaceFromEntity(entity *ldtk_go.Entity) game.Object {
    foodCount := entity.PropertyByIdentifier("FoodCount").AsInt()
    return game.NewFireplace(foodCount)
}
