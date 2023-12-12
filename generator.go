package main

import (
    "Legacy/dungen"
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/gridmap"
    "fmt"
    "strconv"
    "strings"
)

func (g *GridEngine) generateMap(targetMap string) *gridmap.GridMap[*game.Actor, game.Item, game.Object] {
    // !gen_dungeon_level_1
    level := levelFromName(targetMap)
    // todo: get map meta data, we go with some defaults for now

    mapWidth := 32
    mapHeight := 32
    emptyMap := gridmap.NewEmptyMap[*game.Actor, game.Item, game.Object](mapWidth, mapHeight, 11)

    dunGen := dungen.NewAccretionGenerator()
    generatedLayout := dunGen.Generate(mapWidth, mapHeight)

    ladderUpTile := gridmap.Tile{
        DefinedIcon:        20,
        DefinedDescription: "a ladder leading up",
        IsWalkable:         true,
        IsTransparent:      true,
    }
    ladderDownTile := gridmap.Tile{
        DefinedIcon:        21,
        DefinedDescription: "a ladder leading down",
        IsWalkable:         true,
        IsTransparent:      true,
    }

    wallTile := gridmap.Tile{
        DefinedIcon:        80,
        DefinedDescription: "a wall",
        IsWalkable:         false,
        IsTransparent:      false,
    }

    floorTile := gridmap.Tile{
        DefinedIcon:        36,
        DefinedDescription: "a floor",
        IsWalkable:         true,
        IsTransparent:      true,
    }

    for x := 0; x < mapWidth; x++ {
        for y := 0; y < mapHeight; y++ {
            mapPos := geometry.Point{X: x, Y: y}
            switch generatedLayout.GetTile(x, y) {
            case dungen.Wall:
                emptyMap.SetTile(mapPos, wallTile)
            case dungen.Door:
                emptyMap.SetTile(mapPos, floorTile)
                emptyMap.AddObject(game.NewDoor(), mapPos)
            case dungen.Room:
                fallthrough
            case dungen.Corridor:
                fallthrough
            default:
                emptyMap.SetTile(mapPos, floorTile)
            }
        }
    }

    mapEntryPosition := generatedLayout.GetCenterOfRandomRoom()
    emptyMap.SetTile(mapEntryPosition, ladderUpTile)
    emptyMap.AddNamedLocation("ladder_up", mapEntryPosition)

    if level == 1 {
        emptyMap.AddTransitionAt(mapEntryPosition, gridmap.Transition{
            TargetMap:      "Edge_Town",
            TargetLocation: "dungeon_entrance",
        })
    } else {
        emptyMap.AddTransitionAt(mapEntryPosition, gridmap.Transition{
            TargetMap:      levelName(level - 1),
            TargetLocation: "ladder_down",
        })
    }

    mapExitPosition := generatedLayout.GetCenterOfRandomRoom()
    for mapExitPosition == mapEntryPosition {
        mapExitPosition = generatedLayout.GetCenterOfRandomRoom()
    }
    emptyMap.SetTile(mapExitPosition, ladderDownTile)
    emptyMap.AddNamedLocation("ladder_down", mapExitPosition)
    emptyMap.AddTransitionAt(mapExitPosition, gridmap.Transition{
        TargetMap:      levelName(level + 1),
        TargetLocation: "ladder_up",
    })

    emptyMap.SetDisplayName("Dungeon Level " + strconv.Itoa(level))

    return emptyMap
}

func levelFromName(targetMap string) int {
    nameParts := strings.Split(targetMap, "_")
    level, _ := strconv.Atoi(nameParts[len(nameParts)-1])
    return level
}

func levelName(level int) string {
    //!gen_dungeon_level_1
    return fmt.Sprintf("!gen_dungeon_level_%d", level)
}
