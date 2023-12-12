package dungen

import (
    "Legacy/geometry"
    "math/rand"
)

type AccretionGenerator struct {
    random *rand.Rand
}

func NewAccretionGenerator() *AccretionGenerator {
    return &AccretionGenerator{
        random: rand.New(rand.NewSource(42)),
    }
}
func (g *AccretionGenerator) Generate(width, height int) *DungeonMap {
    g.random.Seed(23)

    dMap := NewDungeonMap(width, height)
    rect := g.randomRect(width/2, height/2)

    firstRoom := NewDungeonRoomFromRect(g.random, rect)
    firstRoom.SetPositionOffset(geometry.Point{X: 1 + g.random.Intn(width-rect.Size().X-2), Y: 1 + g.random.Intn(height-rect.Size().Y-2)})

    for !dMap.CanPlaceRoom(firstRoom) {
        firstRoom.SetPositionOffset(geometry.Point{X: 1 + g.random.Intn(width-rect.Size().X-2), Y: 1 + g.random.Intn(height-rect.Size().Y-2)})
    }

    dMap.AddRoomAndSetTiles(firstRoom)

    maxRooms := 100

    for i := 0; i < maxRooms; i++ {
        nextRoom := NewDungeonRoomFromRect(g.random, g.randomRect(width/3, height/3))
        for _, existingRoom := range dMap.rooms {
            if g.tryConnectTo(dMap, nextRoom, existingRoom) {
                dMap.AddRoomAndSetTiles(nextRoom)
                break
            }
        }
    }
    dMap.SetDoorsFromRooms()

    g.addMoreDoors(dMap)

    dMap.FillDeadEnds(g.random)

    return dMap
}

func (g *AccretionGenerator) randomRect(width int, height int) geometry.Rect {
    randWidth := max(3, g.random.Intn(width))
    randHeight := max(3, g.random.Intn(height))
    rect := geometry.NewRect(0, 0, randWidth, randHeight)
    return rect
}

func (g *AccretionGenerator) tryConnectTo(dMap *DungeonMap, newRoom *DungeonRoom, existingRoom *DungeonRoom) bool {
    freeDoors := existingRoom.GetFreeDoorsAsRelative()

    for relativeDoorPos, direction := range freeDoors {
        if relativeNewDoorPos, ok := newRoom.HasFreeRelativeDoorInDirection(direction.Opposite()); ok {
            // both have free doors in opposite directions
            absoluteDoorPos := existingRoom.GetAbsoluteDoorPosition(relativeDoorPos)

            // the new room should be placed so that the new door is at the same position as the existing door
            newRoom.SetPositionOffset(absoluteDoorPos.Sub(relativeNewDoorPos))

            if dMap.CanPlaceRoom(newRoom) {
                newRoom.AddConnectedRoom(relativeNewDoorPos, existingRoom)
                existingRoom.AddConnectedRoom(relativeDoorPos, newRoom)
                return true
            }
        }
    }
    return false
}

func (g *AccretionGenerator) addMoreDoors(dMap *DungeonMap) {
    pathingDistanceThreshold := 10
    dMap.TraverseTilesRandomly(g.random, func(pos geometry.Point) {
        if direction, ok := dMap.CouldBeADoor(pos); ok {
            posOne := pos.Add(direction.ToPoint())
            posTwo := pos.Add(direction.Opposite().ToPoint())
            path := dMap.GetJPSPath(posOne, posTwo)
            if len(path) == 0 || len(path) > pathingDistanceThreshold {
                dMap.AddDoorAndSetTiles(pos, direction)
            }
        }
    })
}
