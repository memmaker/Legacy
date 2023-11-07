package main

import (
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/gridmap"
    "Legacy/recfile"
    "fmt"
    "os"
    "path"
    "strconv"
)

// what do we need to save?
// the state of the party
// the state of all visited maps

func loadPartyState(sourcePath string) (*game.Party, string) {
    fmt.Println("Loading party state from " + sourcePath)
    filename := path.Join(sourcePath, "party.rec")
    f, err := os.Open(filename)
    if err != nil {
        fmt.Println("Error opening party state file: " + err.Error())
        return nil, ""
    }
    defer f.Close()
    partyRecords := recfile.ReadMulti(f)
    return recordsToParty(partyRecords)
}
func saveAllMaps(mapsInMemory map[string]*gridmap.GridMap[*game.Actor, game.Item, game.Object], destinationPath string) {
    mapPaths := path.Join(destinationPath, "maps")
    _ = os.MkdirAll(mapPaths, 0755)
    for mapName, gridMap := range mapsInMemory {
        saveMap(mapName, gridMap, mapPaths)
    }
}
func loadAllMaps(sourcePath string) map[string]*gridmap.GridMap[*game.Actor, game.Item, game.Object] {
    mapsInMemory := make(map[string]*gridmap.GridMap[*game.Actor, game.Item, game.Object])
    mapPaths := path.Join(sourcePath, "maps")
    mapFiles, _ := os.ReadDir(mapPaths)
    for _, dirEntry := range mapFiles {
        if dirEntry.IsDir() {
            continue
        }
        mapName := dirEntry.Name()
        // check for .rec files
        if path.Ext(mapName) != ".rec" {
            continue
        }
        mapName = mapName[:len(mapName)-4]
        mapMetaFilename := path.Join(mapPaths, dirEntry.Name())
        mapBinFilename := path.Join(mapPaths, mapName+".bin")
        mapsInMemory[mapName] = loadMap(mapMetaFilename, mapBinFilename)
    }
    return mapsInMemory
}

func loadMap(metaFilename string, binFilename string) *gridmap.GridMap[*game.Actor, game.Item, game.Object] {
    fmt.Println("Loading map " + metaFilename)
    f, err := os.Open(metaFilename)
    if err != nil {
        fmt.Println("Error opening map file: " + err.Error())
        return nil
    }
    mapRecords := recfile.ReadMulti(f)
    f.Close()
    coreInfo := mapRecords["coreInfo"][0]
    width, height := getMapSize(coreInfo)

    // open the bin file
    f, err = os.Open(binFilename)
    if err != nil {
        fmt.Println("Error opening map file: " + err.Error())
        return nil
    }
    // create a map and read the tiles
    gridMap := gridmap.NewEmptyMap[*game.Actor, game.Item, game.Object](width, height, 12)
    gridMap.ReadTiles(f)
    f.Close()

    placeMapObjects(gridMap, mapRecords)

    return gridMap
}

func placeMapObjects(gridMap *gridmap.GridMap[*game.Actor, game.Item, game.Object], records map[string][]recfile.Record) {
    for mapObjectType, objectRecords := range records {
        switch mapObjectType {
        case "coreInfo":
            continue
        case "actors":
            for _, actorRecord := range objectRecords {
                actor := game.NewActorFromRecord(actorRecord)
                gridMap.AddActor(actor, actor.Pos())
            }
        case "downedActors":
            for _, actorRecord := range objectRecords {
                actor := game.NewActorFromRecord(actorRecord)
                gridMap.AddDownedActor(actor, actor.Pos())
            }
        case "items":
            for _, itemRecord := range objectRecords {
                item := game.NewItemFromRecord(itemRecord)
                gridMap.AddItem(item, item.Pos())
            }
        case "transitions":
            for _, transitionRecord := range objectRecords {
                transitionPos := geometry.MustDecodePoint(transitionRecord[0].Value)
                targetMap := transitionRecord[1].Value
                targetPos := transitionRecord[2].Value
                gridMap.SetTransitionAt(transitionPos, gridmap.Transition{TargetMap: targetMap, TargetLocation: targetPos})
            }
        case "secretDoors":
            for _, secretDoorRecord := range objectRecords {
                secretDoorPos := geometry.MustDecodePoint(secretDoorRecord[0].Value)
                gridMap.SetSecretDoorAt(secretDoorPos)
            }
        default:
            for _, objectRecord := range objectRecords {
                object := game.NewObjectFromRecord(objectRecord, mapObjectType)
                gridMap.AddObject(object, object.Pos())
            }
        }
    }
}

func getMapSize(coreInfo recfile.Record) (int, int) {
    var mapWidth, mapHeight int
    for _, field := range coreInfo {
        switch field.Name {
        case "width":
            mapWidth = field.AsInt()
        case "height":
            mapHeight = field.AsInt()
        }
    }
    return mapWidth, mapHeight
}

func saveMap(name string, gridMap *gridmap.GridMap[*game.Actor, game.Item, game.Object], destinationPath string) {
    fmt.Println("Saving map " + name + " to " + destinationPath)
    mapBinFilename := path.Join(destinationPath, name+".bin")
    mapMetaFilename := path.Join(destinationPath, name+".rec")
    f, err := os.Create(mapBinFilename)
    if err != nil {
        fmt.Println("Error creating map file: " + err.Error())
        return
    }
    gridMap.WriteTiles(f)
    f.Close()

    actorCount := len(gridMap.Actors())
    downedActorCount := len(gridMap.DownedActors())
    itemCount := len(gridMap.Items())
    objectCount := len(gridMap.Objects())

    println(fmt.Sprintf("Saving %d actors, %d downed actors, %d items, and %d objects", actorCount, downedActorCount, itemCount, objectCount))
    var actorsOnMap []recfile.Record
    var downedActorsOnMap []recfile.Record
    for _, actor := range gridMap.Actors() {
        actorsOnMap = append(actorsOnMap, actor.ToRecord())
    }

    for _, actor := range gridMap.DownedActors() {
        downedActorsOnMap = append(downedActorsOnMap, actor.ToRecord())
    }

    var itemsOnMap []recfile.Record
    for _, item := range gridMap.Items() {
        itemsOnMap = append(itemsOnMap, game.ItemToRecord(item))
    }

    objectsOnMap := make(map[string][]recfile.Record)
    for _, object := range gridMap.Objects() {
        record, typeName := object.ToRecordAndType()
        objectsOnMap[typeName] = append(objectsOnMap[typeName], record)
    }

    var coreInfo []recfile.Record
    var transitionInfos []recfile.Record
    var secretDoorInfos []recfile.Record
    coreInfo = append(coreInfo, recfile.Record{
        recfile.Field{Name: "mapName", Value: gridMap.GetName()},
        recfile.Field{Name: "width", Value: recfile.IntStr(gridMap.MapWidth)},
        recfile.Field{Name: "height", Value: recfile.IntStr(gridMap.MapHeight)},
    })

    for pos, transition := range gridMap.Transitions() {
        transitionInfos = append(transitionInfos, recfile.Record{
            recfile.Field{Name: "pos", Value: pos.Encode()},
            recfile.Field{Name: "targetMap", Value: transition.TargetMap},
            recfile.Field{Name: "targetPos", Value: transition.TargetLocation},
        })
    }

    for pos, _ := range gridMap.SecretDoors() {
        secretDoorInfos = append(secretDoorInfos, recfile.Record{
            recfile.Field{Name: "pos", Value: pos.Encode()},
        })
    }

    metaFile, err := os.Create(mapMetaFilename)
    if err != nil {
        fmt.Println("Error creating map file: " + err.Error())
        return
    }
    mapRecords := map[string][]recfile.Record{
        "coreInfo":     coreInfo,
        "actors":       actorsOnMap,
        "downedActors": downedActorsOnMap,
        "items":        itemsOnMap,
        "transitions":  transitionInfos,
        "secretDoors":  secretDoorInfos,
    }
    // merge the objectsOnMap into mapRecords
    for typeName, records := range objectsOnMap {
        mapRecords[typeName] = records
    }
    writeMetaErr := recfile.WriteMulti(metaFile, mapRecords)
    if writeMetaErr != nil {
        return
    }
    metaFile.Close()

}
func saveExtendedState(*game.Flags, *game.PlayerKnowledge, string) {
    // TODO
}

func loadExtendedState(directory string) (*game.Flags, *game.PlayerKnowledge) {
    // TODO
    return game.NewFlags(), game.NewPlayerKnowledge()
}
func savePartyState(party *game.Party, destinationPath string) bool {
    partyRecords := partyToRecords(party)
    // create a file
    // write the party state to the file
    // close the file
    fmt.Println("Saving party state to " + destinationPath)
    filename := path.Join(destinationPath, "party.rec")
    f, err := os.Create(filename)
    if err != nil {
        fmt.Println("Error creating party state file: " + err.Error())
        return false
    }
    writeErr := recfile.WriteMulti(f, partyRecords)
    if writeErr != nil {
        fmt.Println("Error writing party state to file: " + writeErr.Error())
        return false
    }
    errClose := f.Close()
    if errClose != nil {
        fmt.Println("Error closing party state file: " + errClose.Error())
        return false
    }
    return true
}

func recordsToParty(records map[string][]recfile.Record) (*game.Party, string) {
    var members []*game.Actor
    var currentMapName string
    for _, memberRecord := range records["chars"] {
        member := game.NewActorFromRecord(memberRecord)
        members = append(members, member)
    }

    party := game.NewParty(members[0])
    for _, member := range members[1:] {
        party.AddMember(member)
    }

    partyRecord := records["party"][0]
    var previousItem game.Item
    for _, field := range partyRecord {
        switch field.Name {
        case "food":
            party.SetFood(field.AsInt())
        case "gold":
            party.SetGold(field.AsInt())
        case "lockpicks":
            party.SetLockpicks(field.AsInt())
        case "item":
            previousItem = game.NewItemFromString(field.Value)
            party.AddItem(previousItem)
        case "prevItemEquippedBy":
            member := party.GetMember(field.AsInt())
            member.Equip(previousItem)
        case "currentMap":
            currentMapName = field.Value
        }
    }

    return party, currentMapName
}

func partyToRecords(party *game.Party) map[string][]recfile.Record {
    var charRecords []recfile.Record
    var partyRecord []recfile.Record

    partyInfos := recfile.Record{
        recfile.Field{Name: "food", Value: strconv.Itoa(party.GetFood())},
        recfile.Field{Name: "gold", Value: strconv.Itoa(party.GetGold())},
        recfile.Field{Name: "lockpicks", Value: strconv.Itoa(party.GetLockpicks())},
        recfile.Field{Name: "currentMap", Value: party.GetCurrentMapName()},
    }

    for _, key := range party.GetKeys() {
        partyInfos = append(partyInfos, recfile.Field{Name: "item", Value: key.Encode()})
    }

    for _, item := range party.GetFlatInventory() {
        partyInfos = append(partyInfos, recfile.Field{Name: "item", Value: item.Encode()})

        if wearableItem, ok := item.(game.Wearable); ok && wearableItem.IsEquipped() {
            memberIndexOfWearer := party.GetMemberIndex(wearableItem.GetWearer())
            fieldValue := strconv.Itoa(memberIndexOfWearer)
            partyInfos = append(partyInfos, recfile.Field{Name: "prevItemEquippedBy", Value: fieldValue})
        }
    }

    partyRecord = append(partyRecord, partyInfos)
    for _, member := range party.GetMembers() {
        charRecords = append(charRecords, member.ToRecord())
    }
    return map[string][]recfile.Record{
        "party": partyRecord,
        "chars": charRecords,
    }
}
