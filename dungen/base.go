package dungen

import (
    "Legacy/geometry"
    "math/rand"
)

type DungeonTile int

const (
    Wall DungeonTile = iota
    Door
    Room
    Corridor
)

type DungeonMap struct {
    width      int
    height     int
    tiles      []DungeonTile
    rooms      []*DungeonRoom
    pathfinder *geometry.PathRange
}

func (m *DungeonMap) GetJPSPath(start geometry.Point, end geometry.Point) []geometry.Point {
    if !m.IsWalkable(end) || !m.IsWalkable(start) {
        return []geometry.Point{}
    }
    //println(fmt.Sprintf("JPS from %v to %v", start, end))
    return m.pathfinder.JPSPath([]geometry.Point{}, start, end, m.IsWalkable, false)
}
func (m *DungeonMap) AddRoomAndSetTiles(room *DungeonRoom) {
    m.rooms = append(m.rooms, room)
    for _, tile := range room.GetAbsoluteFloorTiles() {
        m.SetRoom(tile.X, tile.Y)
    }
}

func (m *DungeonMap) SetDoorsFromRooms() {
    for _, room := range m.rooms {
        for _, door := range room.GetUsedAbsoluteDoorTiles() {
            m.SetDoor(door.X, door.Y)
        }
    }
}

func (m *DungeonMap) SetWall(x, y int) {
    m.tiles[x+y*m.width] = Wall
}

func (m *DungeonMap) SetCorridor(x, y int) {
    m.tiles[x+y*m.width] = Corridor
}

func (m *DungeonMap) SetRoom(x, y int) {
    m.tiles[x+y*m.width] = Room
}

func (m *DungeonMap) GetTile(x int, y int) DungeonTile {
    return m.tiles[x+y*m.width]
}

func (m *DungeonMap) GetTileAt(pos geometry.Point) DungeonTile {
    return m.tiles[pos.X+pos.Y*m.width]
}

func (m *DungeonMap) GetCenterOfRandomRoom() geometry.Point {
    return m.GetRandomRoom().Center()
}

func (m *DungeonMap) GetRandomRoom() *DungeonRoom {
    randomIndex := rand.Intn(len(m.rooms))
    return m.rooms[randomIndex]
}

func (m *DungeonMap) AllRooms() []*DungeonRoom {
    return m.rooms
}

func (m *DungeonMap) Print() {
    for y := 0; y < m.height; y++ {
        for x := 0; x < m.width; x++ {
            switch m.GetTile(x, y) {
            case Wall:
                print("#")
            case Corridor:
                print(".")
            case Room:
                print(".")
            case Door:
                print("+")
            }
        }
        println()
    }
}

func (m *DungeonMap) SetDoor(x int, y int) {
    m.tiles[x+y*m.width] = Door
}

func (m *DungeonMap) CanPlaceRoom(room *DungeonRoom) bool {
    nb := geometry.Neighbors{}
    for _, tile := range room.GetAbsoluteFloorTiles() {
        if !m.Contains(tile) || m.GetTileAt(tile) != Wall {
            return false
        }
        unusableNeighborTiles := nb.All(tile, func(pos geometry.Point) bool {
            return !m.Contains(pos) || m.GetTileAt(pos) != Wall
        })
        if len(unusableNeighborTiles) > 0 {
            return false
        }
    }
    return true
}

func (m *DungeonMap) Contains(pos geometry.Point) bool {
    return pos.X >= 0 && pos.X < m.width && pos.Y >= 0 && pos.Y < m.height
}

func (m *DungeonMap) CouldBeADoor(pos geometry.Point) (geometry.CompassDirection, bool) {
    // needs to be a walltile, that has empty tiles on both sides in cardinal directions

    if m.GetTileAt(pos) != Wall {
        return geometry.East, false
    }

    // north/south
    if m.IsEmptySpace(pos.Add(geometry.North.ToPoint())) && m.IsEmptySpace(pos.Add(geometry.South.ToPoint())) {
        return geometry.North, true
    }

    // east/west
    if m.IsEmptySpace(pos.Add(geometry.East.ToPoint())) && m.IsEmptySpace(pos.Add(geometry.West.ToPoint())) {
        return geometry.East, true
    }

    return geometry.East, false
}

func (m *DungeonMap) IsEmptySpace(pos geometry.Point) bool {
    if !m.Contains(pos) {
        return false
    }
    tileAt := m.GetTileAt(pos)
    return tileAt == Room || tileAt == Corridor
}

func (m *DungeonMap) IsWalkable(pos geometry.Point) bool {
    if !m.Contains(pos) {
        return false
    }
    tileAt := m.GetTileAt(pos)
    return tileAt != Wall
}

func (m *DungeonMap) AddDoorAndSetTiles(doorPos geometry.Point, direction geometry.CompassDirection) {
    posOne := doorPos.Add(direction.ToPoint())
    posTwo := doorPos.Add(direction.Opposite().ToPoint())

    roomOne := m.GetRoomAt(posOne)
    roomTwo := m.GetRoomAt(posTwo)

    if roomOne == nil || roomTwo == nil {
        return
    }

    roomOne.AddConnectedRoom(doorPos, roomTwo)
    roomTwo.AddConnectedRoom(doorPos, roomOne)

    m.SetDoor(doorPos.X, doorPos.Y)
}

func (m *DungeonMap) GetRoomAt(posOne geometry.Point) *DungeonRoom {
    for _, room := range m.rooms {
        if room.Contains(posOne) {
            return room
        }
    }
    return nil
}

func (m *DungeonMap) TraverseTilesRandomly(random *rand.Rand, traversalFunc func(pos geometry.Point)) {
    randomIndices := random.Perm(m.width * m.height)
    for _, index := range randomIndices {
        x := index % m.width
        y := index / m.width
        traversalFunc(geometry.Point{X: x, Y: y})
    }
}

func (m *DungeonMap) FillDeadEnds(random *rand.Rand) {
    m.TraverseTilesRandomly(random, func(pos geometry.Point) {
        direction, isDeadEnd := m.IsDeadEnd(pos)
        for isDeadEnd {
            m.SetWall(pos.X, pos.Y)
            pos = pos.Add(direction.ToPoint())
            direction, isDeadEnd = m.IsDeadEnd(pos)
        }
    })
}

func (m *DungeonMap) IsDeadEnd(pos geometry.Point) (geometry.CompassDirection, bool) {
    if !m.IsEmptySpace(pos) {
        return 0, false
    }

    nb := geometry.Neighbors{}
    neighoringWalls := nb.All(pos, func(pos geometry.Point) bool {
        return m.Contains(pos) && m.GetTileAt(pos) == Wall
    })

    var openDirection geometry.CompassDirection
    cardinalDirs := []geometry.CompassDirection{geometry.North, geometry.South, geometry.East, geometry.West}

    for _, dir := range cardinalDirs {
        if m.IsEmptySpace(pos.Add(dir.ToPoint())) {
            openDirection = dir
        }
    }

    return openDirection, len(neighoringWalls) == 7
}

func NewDungeonMap(width, height int) *DungeonMap {
    return &DungeonMap{
        width:      width,
        height:     height,
        tiles:      make([]DungeonTile, width*height),
        rooms:      make([]*DungeonRoom, 0),
        pathfinder: geometry.NewPathRange(geometry.NewRect(0, 0, width, height)),
    }
}

type DungeonRoom struct {
    roomPositionOffset geometry.Point

    center geometry.Point

    floorTiles map[geometry.Point]bool

    availableDoorTiles map[geometry.Point]geometry.CompassDirection
    connectedRooms     map[geometry.Point]*DungeonRoom
}

func (r *DungeonRoom) Center() geometry.Point {
    return r.roomPositionOffset.Add(r.center)
}

func (r *DungeonRoom) GetFreeDoorsAsRelative() map[geometry.Point]geometry.CompassDirection {
    result := make(map[geometry.Point]geometry.CompassDirection)
    for point, direction := range r.availableDoorTiles {
        if _, ok := r.connectedRooms[point]; !ok {
            result[point] = direction
        }
    }
    return result
}

func (r *DungeonRoom) HasFreeRelativeDoorInDirection(direction geometry.CompassDirection) (geometry.Point, bool) {
    for pos, dir := range r.availableDoorTiles {
        if dir == direction {
            if _, ok := r.connectedRooms[pos]; !ok {
                return pos, true
            } else {
                return pos, false
            }
        }
    }
    return geometry.Point{}, false
}

func (r *DungeonRoom) SetPositionOffset(offset geometry.Point) {
    r.roomPositionOffset = offset
}

func (r *DungeonRoom) GetAbsoluteDoorPosition(doorPos geometry.Point) geometry.Point {
    return r.roomPositionOffset.Add(doorPos)
}

func (r *DungeonRoom) GetAbsoluteFloorTiles() []geometry.Point {
    result := make([]geometry.Point, 0)
    for point, _ := range r.floorTiles {
        result = append(result, r.roomPositionOffset.Add(point))
    }
    return result
}

func (r *DungeonRoom) AddConnectedRoom(doorPos geometry.Point, room *DungeonRoom) {
    r.connectedRooms[doorPos] = room
}

func (r *DungeonRoom) GetUsedAbsoluteDoorTiles() []geometry.Point {
    result := make([]geometry.Point, 0)
    for point, _ := range r.availableDoorTiles {
        if _, ok := r.connectedRooms[point]; !ok {
            continue
        }
        result = append(result, r.roomPositionOffset.Add(point))
    }
    return result
}

func (r *DungeonRoom) AddCorridor(direction geometry.CompassDirection, length int) {
    // remove door tile from available door tiles
    doorPos, ok := r.HasFreeRelativeDoorInDirection(direction)
    if !ok {
        return
    }
    delete(r.availableDoorTiles, doorPos)

    // add corridor tiles
    for i := 0; i < length; i++ {
        r.floorTiles[doorPos.Add(direction.ToPoint().Mul(i))] = true
    }

    // add door tile at the end of the corridor
    newDoorPos := doorPos.Add(direction.ToPoint().Mul(length))
    r.availableDoorTiles[newDoorPos] = direction

    if length > 1 {
        // add another set of doors at 90 degrees to the corridor
        startPos := doorPos.Add(direction.ToPoint().Mul(length - 1))

        secondDoor := startPos.Add(direction.TurnRightBy90().ToPoint())
        r.availableDoorTiles[secondDoor] = direction.TurnRightBy90()

        thirdDoor := startPos.Add(direction.TurnLeftBy90().ToPoint())
        r.availableDoorTiles[thirdDoor] = direction.TurnLeftBy90()
    }
}

func (r *DungeonRoom) Contains(absolutePosition geometry.Point) bool {
    relative := absolutePosition.Sub(r.roomPositionOffset)
    if _, ok := r.floorTiles[relative]; ok {
        return true
    }
    return false
}

func NewDungeonRoomFromRect(random *rand.Rand, bounds geometry.Rect) *DungeonRoom {
    rectRoom := &DungeonRoom{
        center:         bounds.Center(),
        floorTiles:     roomTilesFromRect(bounds),
        connectedRooms: make(map[geometry.Point]*DungeonRoom),
        availableDoorTiles: map[geometry.Point]geometry.CompassDirection{
            bounds.GetRandomPointOnEdge(random, geometry.North).Add(geometry.North.ToPoint()): geometry.North,
            bounds.GetRandomPointOnEdge(random, geometry.South).Add(geometry.South.ToPoint()): geometry.South,
            bounds.GetRandomPointOnEdge(random, geometry.East).Add(geometry.East.ToPoint()):   geometry.East,
            bounds.GetRandomPointOnEdge(random, geometry.West).Add(geometry.West.ToPoint()):   geometry.West,
        },
    }
    // choose a random direction or none

    switch random.Intn(5) {
    case 0:
        rectRoom.AddCorridor(geometry.North, random.Intn(4)+1)
    case 1:
        rectRoom.AddCorridor(geometry.South, random.Intn(4)+1)
    case 2:
        rectRoom.AddCorridor(geometry.East, random.Intn(4)+1)
    case 3:
        rectRoom.AddCorridor(geometry.West, random.Intn(4)+1)
    }

    return rectRoom
}

func roomTilesFromRect(bounds geometry.Rect) map[geometry.Point]bool {
    result := make(map[geometry.Point]bool)
    for x := bounds.Min.X; x < bounds.Max.X; x++ {
        for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
            result[geometry.Point{X: x, Y: y}] = true
        }
    }
    return result
}
