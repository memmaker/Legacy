package gridmap

import (
    "Legacy/geometry"
    "bytes"
    "encoding/gob"
    "fmt"
    "io"
    "math"
    "math/rand"
    "os"
    "sort"
    "strconv"
    "time"
)

type Transition struct {
    TargetMap      string
    TargetLocation string
}
type MapObject interface {
    Pos() geometry.Point
    Icon(tick uint64) int32
    SetPos(geometry.Point)
}

type MapActor interface {
    MapObject
}
type MapObjectWithProperties[ActorType interface {
    comparable
    MapActor
}] interface {
    MapObject
    IsWalkable(person ActorType) bool
    IsTransparent() bool
    IsPassableForProjectile() bool
}
type ZoneType int

func (t ZoneType) ToString() string {
    switch t {
    case ZoneTypePublic:
        return "Public"
    case ZoneTypePrivate:
        return "Private"
    case ZoneTypeHighSecurity:
        return "High Security"
    case ZoneTypeDropOff:
        return "Drop Off"
    }
    return "Unknown"
}

func NewZoneTypeFromString(str string) ZoneType {
    switch str {
    case "Public":
        return ZoneTypePublic
    case "Private":
        return ZoneTypePrivate
    case "High Security":
        return ZoneTypeHighSecurity
    case "Drop Off":
        return ZoneTypeDropOff
    }
    return ZoneTypePublic
}

const (
    ZoneTypePublic ZoneType = iota
    ZoneTypePrivate
    ZoneTypeHighSecurity
    ZoneTypeDropOff
)

type ZoneInfo struct {
    Name        string
    Type        ZoneType
    AmbienceCue string
}

const PublicZoneName = "Public Space"

func (i ZoneInfo) IsDropOff() bool {
    return i.Type == ZoneTypeDropOff
}

func (i ZoneInfo) IsHighSecurity() bool {
    return i.Type == ZoneTypeHighSecurity || i.Type == ZoneTypeDropOff
}

func (i ZoneInfo) IsPublic() bool {
    return i.Type == ZoneTypePublic
}

func (i ZoneInfo) IsPrivate() bool {
    return i.Type == ZoneTypePrivate
}

func (i ZoneInfo) ToString() string {
    return fmt.Sprintf("%s (%s)", i.Name, i.Type.ToString())
}

func NewZone(name string) *ZoneInfo {
    return &ZoneInfo{
        Name: name,
    }
}
func NewPublicZone(name string) *ZoneInfo {
    return &ZoneInfo{
        Name: name,
        Type: ZoneTypePublic,
    }
}

type GlobalMapDataOnDisk struct {
    MissionTitle      string
    Width             int
    Height            int
    PlayerSpawn       geometry.Point
    MaxLightIntensity float64
    MaxVisionRange    int
    TimeOfDay         time.Time
    AmbienceSoundCue  string
}

func (d GlobalMapDataOnDisk) ToString() string {
    return fmt.Sprintf("Mission Title: %s\nWidth: %d\nHeight: %d\nPlayer Spawn: %s\nMax Light Intensity: %f\nMax Vision Range: %d\nTime of Day: %s\nAmbience Sound Cue: %s",
        d.MissionTitle, d.Width, d.Height, d.PlayerSpawn.String(), d.MaxLightIntensity, d.MaxVisionRange, d.TimeOfDay.String(), d.AmbienceSoundCue)
}

type GridMap[ActorType interface {
    comparable
    MapActor
}, ItemType interface {
    comparable
    MapObject
}, ObjectType interface {
    comparable
    MapObjectWithProperties[ActorType]
}] struct {
    name            string
    Cells           []MapCell[ActorType, ItemType, ObjectType]
    AllActors       []ActorType
    AllDownedActors []ActorType
    removedActors   []ActorType
    AllItems        []ItemType
    AllObjects      []ObjectType

    PlayerSpawn geometry.Point

    MapWidth       int
    MapHeight      int
    MaxVisionRange int

    pathfinder  *geometry.PathRange
    ListOfZones []*ZoneInfo
    ZoneMap     []*ZoneInfo
    Player      ActorType

    maxLOSRange geometry.Rect
    TimeOfDay   time.Time

    NamedLocations   map[string]geometry.Point
    AmbienceSoundCue string
    noClip           bool

    transitionMap map[geometry.Point]Transition
    secretDoors   map[geometry.Point]bool

    namedRects   map[string]geometry.Rect
    namedTrigger map[string]Trigger
    displayName  string
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddZone(zone *ZoneInfo) {
    m.ListOfZones = append(m.ListOfZones, zone)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetTile(position geometry.Point, mapTile Tile) {
    if !m.Contains(position) {
        return
    }
    index := position.Y*m.MapWidth + position.X
    m.Cells[index].TileType = mapTile
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetZone(position geometry.Point, zone *ZoneInfo) {
    if !m.Contains(position) {
        return
    }
    m.ZoneMap[position.Y*m.MapWidth+position.X] = zone
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveItemAt(position geometry.Point) {
    m.RemoveItem(m.ItemAt(position))
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveObjectAt(position geometry.Point) {
    m.RemoveObject(m.ObjectAt(position))
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveObject(obj ObjectType) {
    m.Cells[obj.Pos().Y*m.MapWidth+obj.Pos().X] = m.Cells[obj.Pos().Y*m.MapWidth+obj.Pos().X].WithObjectHereRemoved(obj)
    for i := len(m.AllObjects) - 1; i >= 0; i-- {
        if m.AllObjects[i] == obj {
            m.AllObjects = append(m.AllObjects[:i], m.AllObjects[i+1:]...)
            return
        }
    }
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IterAll(f func(p geometry.Point, c MapCell[ActorType, ItemType, ObjectType])) {
    for y := 0; y < m.MapHeight; y++ {
        for x := 0; x < m.MapWidth; x++ {
            f(geometry.Point{X: x, Y: y}, m.Cells[y*m.MapWidth+x])
        }
    }
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IterWindow(window geometry.Rect, f func(p geometry.Point, c MapCell[ActorType, ItemType, ObjectType])) {
    for y := window.Min.Y; y < window.Max.Y; y++ {
        for x := window.Min.X; x < window.Max.X; x++ {
            mapPos := geometry.Point{X: x, Y: y}
            if !m.Contains(mapPos) {
                continue
            }
            f(mapPos, m.Cells[y*m.MapWidth+x])
        }
    }
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetPlayerSpawn(position geometry.Point) {
    m.PlayerSpawn = position
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SaveToDisk(path string) error {
    file, _ := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)
    encodeErr := enc.Encode(m)
    encodedMap := buf.Bytes()
    writeCount, writeErr := file.Write(encodedMap)
    if encodeErr == nil && writeErr == nil {
        println("Map saved to file: " + path + ", length: " + strconv.Itoa(writeCount))
        return nil
    } else {
        println("Error saving map to file: " + path)
        if encodeErr != nil {
            println(encodeErr.Error())
            return encodeErr
        } else if writeErr != nil {
            println(writeErr.Error())
            return writeErr
        }
    }
    return nil
}

func (m *GridMap[ActorType, ItemType, ObjectType]) CellAt(location geometry.Point) MapCell[ActorType, ItemType, ObjectType] {
    return m.Cells[m.MapWidth*location.Y+location.X]
}

func (m *GridMap[ActorType, ItemType, ObjectType]) ItemAt(location geometry.Point) ItemType {
    return *m.Cells[m.MapWidth*location.Y+location.X].Item
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsItemAt(location geometry.Point) bool {
    return m.Cells[m.MapWidth*location.Y+location.X].Item != nil
}
func (m *GridMap[ActorType, ItemType, ObjectType]) SetActorToDowned(a ActorType) {
    if !m.RemoveActor(a) {
        println("Could not remove actor from map")
        return
    }
    m.AllDownedActors = append(m.AllDownedActors, a)
    if m.IsDownedActorAt(a.Pos()) && m.DownedActorAt(a.Pos()) != a {
        m.displaceDownedActor(a)
        return
    }
    m.Cells[a.Pos().Y*m.MapWidth+a.Pos().X] = m.Cells[a.Pos().Y*m.MapWidth+a.Pos().X].WithDownedActor(a)
}
func (m *GridMap[ActorType, ItemType, ObjectType]) SetActorToRemoved(person ActorType) {
    m.RemoveActor(person)
    m.RemoveDownedActor(person)
    m.removedActors = append(m.removedActors, person)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) MoveItem(item ItemType, to geometry.Point) {
    m.Cells[item.Pos().Y*m.MapWidth+item.Pos().X] = m.Cells[item.Pos().Y*m.MapWidth+item.Pos().X].WithItemHereRemoved(item)
    item.SetPos(to)
    m.Cells[to.Y*m.MapWidth+to.X] = m.Cells[to.Y*m.MapWidth+to.X].WithItem(item)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetRandomFreeNeighbor(location geometry.Point) geometry.Point {
    freeNearbyPositions := m.GetFreeCellsForDistribution(location, 1, m.IsCurrentlyPassable)
    if len(freeNearbyPositions) == 0 {
        return location
    }
    return freeNearbyPositions[0]
}

type SetOfPoints map[geometry.Point]bool

func (s *SetOfPoints) Pop() geometry.Point {
    for k := range *s {
        delete(*s, k)
        return k
    }
    return geometry.Point{}
}
func (s *SetOfPoints) Contains(p geometry.Point) bool {
    _, ok := (*s)[p]
    return ok
}

func (s *SetOfPoints) ToSlice() []geometry.Point {
    result := make([]geometry.Point, 0)
    for k := range *s {
        result = append(result, k)
    }
    return result
}
func (m *GridMap[ActorType, ItemType, ObjectType]) GetFreeCellsForDistribution(position geometry.Point, neededCellCount int, freePredicate func(p geometry.Point) bool) []geometry.Point {
    foundFreeCells := make(SetOfPoints)
    currentPosition := position
    openList := make(SetOfPoints)
    closedList := make(SetOfPoints)
    closedList[currentPosition] = true

    for _, neighbor := range m.GetFilteredNeighbors(currentPosition, m.IsTileWalkable) {
        openList[neighbor] = true
    }
    for len(foundFreeCells) < neededCellCount && len(openList) > 0 {
        freeNeighbors := m.GetFilteredNeighbors(currentPosition, freePredicate)
        for _, neighbor := range freeNeighbors {
            foundFreeCells[neighbor] = true
        }
        // pop from open list
        pop := openList.Pop()
        currentPosition = pop
        for _, neighbor := range m.GetFilteredNeighbors(currentPosition, m.IsTileWalkable) {
            if !closedList.Contains(neighbor) {
                openList[neighbor] = true
            }
        }
        closedList[currentPosition] = true
    }

    freeCells := foundFreeCells.ToSlice()
    sort.Slice(freeCells, func(i, j int) bool {
        return geometry.DistanceSquared(freeCells[i], position) < geometry.DistanceSquared(freeCells[j], position)
    })
    return freeCells
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsNextToTileWithSpecial(pos geometry.Point, specialType SpecialTileType) bool {
    for _, neighbor := range m.GetAllCardinalNeighbors(pos) {
        if m.CellAt(neighbor).TileType.Special == specialType {
            return true
        }
    }
    return false
}
func (m *GridMap[ActorType, ItemType, ObjectType]) Neighbors(point geometry.Point) []geometry.Point {
    return m.GetFilteredCardinalNeighbors(point, func(p geometry.Point) bool {
        return m.Contains(p) && m.IsTileWalkable(p)
    })
}

func (m *GridMap[ActorType, ItemType, ObjectType]) Cost(point geometry.Point, point2 geometry.Point) int {
    return geometry.DistanceManhattan(point, point2)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveItem(item ItemType) {
    m.Cells[item.Pos().Y*m.MapWidth+item.Pos().X] = m.Cells[item.Pos().Y*m.MapWidth+item.Pos().X].WithItemHereRemoved(item)
    for i := len(m.AllItems) - 1; i >= 0; i-- {
        if m.AllItems[i] == item {
            m.AllItems = append(m.AllItems[:i], m.AllItems[i+1:]...)
            return
        }
    }
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetAllCardinalNeighbors(pos geometry.Point) []geometry.Point {
    neighbors := geometry.Neighbors{}
    allCardinalNeighbors := neighbors.Cardinal(pos, func(p geometry.Point) bool {
        return m.Contains(p)
    })
    return allCardinalNeighbors
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsSpecialAt(pos geometry.Point, specialValue SpecialTileType) bool {
    special := m.CellAt(pos).TileType.Special
    return special == specialValue
}

func (m *GridMap[ActorType, ItemType, ObjectType]) WavePropagationFrom(pos geometry.Point, size int, pressure int) map[int][]geometry.Point {

    soundAnimationMap := make(map[int][]geometry.Point)
    m.pathfinder.DijkstraMap(m, []geometry.Point{pos}, size)
    for _, v := range m.pathfinder.DijkstraIterNodes {
        cost := v.Cost
        point := v.P
        if soundAnimationMap[cost] == nil {
            soundAnimationMap[cost] = make([]geometry.Point, 0)
        }
        soundAnimationMap[cost] = append(soundAnimationMap[cost], point)
    }
    return soundAnimationMap
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetConnected(startLocation geometry.Point, traverse func(p geometry.Point) bool) []geometry.Point {
    results := make([]geometry.Point, 0)
    for _, node := range m.pathfinder.BreadthFirstMap(MapPather{neighborPredicate: traverse, allNeighbors: m.GetAllCardinalNeighbors}, []geometry.Point{startLocation}, 100) {
        results = append(results, node.P)
    }
    return results
}

// update for entities:
// call update for every updatable entity (genMap, AllItems, AllObjects, tiles)
// default: just return
// entities have an internal schedule, waiting for ticks to happen

func NewEmptyMap[ActorType interface {
    comparable
    MapActor
}, ItemType interface {
    comparable
    MapObject
}, ObjectType interface {
    comparable
    MapObjectWithProperties[ActorType]
}](width, height, maxVisionRange int) *GridMap[ActorType, ItemType, ObjectType] {
    pathRange := geometry.NewPathRange(geometry.NewRect(0, 0, width, height))
    publicSpaceZone := NewPublicZone(PublicZoneName)
    m := &GridMap[ActorType, ItemType, ObjectType]{
        Cells:           make([]MapCell[ActorType, ItemType, ObjectType], width*height),
        AllActors:       make([]ActorType, 0),
        AllDownedActors: make([]ActorType, 0),
        AllItems:        make([]ItemType, 0),
        AllObjects:      make([]ObjectType, 0),
        NamedLocations:  map[string]geometry.Point{},
        ListOfZones:     []*ZoneInfo{publicSpaceZone},
        ZoneMap:         NewZoneMap(publicSpaceZone, width, height),
        MapWidth:        width,
        MapHeight:       height,
        TimeOfDay:       time.Now(),
        pathfinder:      pathRange,
        maxLOSRange:     geometry.NewRect(-maxVisionRange, -maxVisionRange, maxVisionRange+1, maxVisionRange+1),
        MaxVisionRange:  maxVisionRange,
        secretDoors:     make(map[geometry.Point]bool),
        transitionMap:   make(map[geometry.Point]Transition),
        namedRects:      make(map[string]geometry.Rect),
        namedTrigger:    make(map[string]Trigger),
    }
    m.Fill(MapCell[ActorType, ItemType, ObjectType]{
        TileType: Tile{
            DefinedIcon:        ' ',
            DefinedDescription: "empty space",
            IsWalkable:         true,
            IsTransparent:      true,
            Special:            SpecialTileNone,
        },
        IsExplored: false,
    })
    return m
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetCell(p geometry.Point) MapCell[ActorType, ItemType, ObjectType] {
    return m.Cells[p.X+p.Y*m.MapWidth]
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetCell(p geometry.Point, cell MapCell[ActorType, ItemType, ObjectType]) {
    m.Cells[p.X+p.Y*m.MapWidth] = cell
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetActor(p geometry.Point) ActorType {
    return *m.Cells[p.X+p.Y*m.MapWidth].Actor
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveActor(actor ActorType) bool {
    m.Cells[actor.Pos().X+actor.Pos().Y*m.MapWidth] = m.Cells[actor.Pos().X+actor.Pos().Y*m.MapWidth].WithActorHereRemoved(actor)
    for i := len(m.AllActors) - 1; i >= 0; i-- {
        if m.AllActors[i] == actor {
            m.AllActors = append(m.AllActors[:i], m.AllActors[i+1:]...)
            return true
        }
    }
    return false
}

// MoveActor Should only be called my the model, so we can ensure that a HUD IsFinished will follow
func (m *GridMap[ActorType, ItemType, ObjectType]) MoveActor(actor ActorType, newPos geometry.Point) {
    if !m.Contains(newPos) {
        return
    }
    if !m.IsWalkableFor(newPos, actor) {
        return
    }
    if m.Contains(actor.Pos()) {
        m.Cells[actor.Pos().X+actor.Pos().Y*m.MapWidth] = m.Cells[actor.Pos().X+actor.Pos().Y*m.MapWidth].WithActorHereRemoved(actor)
    }
    actor.SetPos(newPos)
    m.Cells[newPos.X+newPos.Y*m.MapWidth] = m.Cells[newPos.X+newPos.Y*m.MapWidth].WithActor(actor)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) MoveObject(obj ObjectType, newPos geometry.Point) {
    m.Cells[obj.Pos().X+obj.Pos().Y*m.MapWidth] = m.Cells[obj.Pos().X+obj.Pos().Y*m.MapWidth].WithObjectHereRemoved(obj)
    obj.SetPos(newPos)
    m.Cells[newPos.X+newPos.Y*m.MapWidth] = m.Cells[newPos.X+newPos.Y*m.MapWidth].WithObject(obj)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) Fill(mapCell MapCell[ActorType, ItemType, ObjectType]) {
    for i := range m.Cells {
        m.Cells[i] = mapCell
    }
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsTransparent(p geometry.Point) bool {
    if !m.Contains(p) {
        return false
    }

    if objectAt, ok := m.TryGetObjectAt(p); ok && !objectAt.IsTransparent() {
        return false
    }

    return m.GetCell(p).TileType.IsTransparent
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsTileWalkable(point geometry.Point) bool {
    if !m.Contains(point) {
        return false
    }
    return m.GetCell(point).TileType.IsWalkable
}

func (m *GridMap[ActorType, ItemType, ObjectType]) Contains(dest geometry.Point) bool {
    return dest.X >= 0 && dest.X < m.MapWidth && dest.Y >= 0 && dest.Y < m.MapHeight
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsActorAt(location geometry.Point) bool {
    if !m.Contains(location) {
        return false
    }
    return m.Cells[location.X+location.Y*m.MapWidth].Actor != nil
}

func (m *GridMap[ActorType, ItemType, ObjectType]) ActorAt(location geometry.Point) ActorType {
    return *m.Cells[location.X+location.Y*m.MapWidth].Actor
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsDownedActorAt(location geometry.Point) bool {
    if !m.Contains(location) {
        return false
    }
    return m.Cells[location.X+location.Y*m.MapWidth].DownedActor != nil
}

func (m *GridMap[ActorType, ItemType, ObjectType]) DownedActorAt(location geometry.Point) ActorType {
    return *m.Cells[location.X+location.Y*m.MapWidth].DownedActor
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsObjectAt(location geometry.Point) bool {
    return m.Cells[location.X+location.Y*m.MapWidth].Object != nil
}

func (m *GridMap[ActorType, ItemType, ObjectType]) ObjectAt(location geometry.Point) ObjectType {
    return *m.Cells[location.X+location.Y*m.MapWidth].Object
}
func (m *GridMap[ActorType, ItemType, ObjectType]) ZoneAt(p geometry.Point) *ZoneInfo {
    return m.ZoneMap[m.MapWidth*p.Y+p.X]
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetFilteredCardinalNeighbors(pos geometry.Point, filter func(geometry.Point) bool) []geometry.Point {
    neighbors := geometry.Neighbors{}
    filtered := neighbors.Cardinal(pos, filter)
    return filtered
}

func (m *GridMap[ActorType, ItemType, ObjectType]) Actors() []ActorType {
    return m.AllActors
}

func (m *GridMap[ActorType, ItemType, ObjectType]) DownedActors() []ActorType {
    return m.AllDownedActors
}

func (m *GridMap[ActorType, ItemType, ObjectType]) Items() []ItemType {
    return m.AllItems
}

func (m *GridMap[ActorType, ItemType, ObjectType]) Objects() []ObjectType {
    return m.AllObjects
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetFilteredNeighbors(pos geometry.Point, filter func(geometry.Point) bool) []geometry.Point {
    neighbors := geometry.Neighbors{}
    filtered := neighbors.All(pos, filter)
    return filtered
}

func (m *GridMap[ActorType, ItemType, ObjectType]) displaceDownedActor(a ActorType) {
    free := m.GetFreeCellsForDistribution(a.Pos(), 1, func(p geometry.Point) bool {
        return !m.IsDownedActorAt(p) && m.IsWalkable(p)
    })
    freePos := free[0]
    m.MoveDownedActor(a, freePos)
}
func (m *GridMap[ActorType, ItemType, ObjectType]) GetJPSPath(start geometry.Point, end geometry.Point, isWalkable func(geometry.Point) bool) []geometry.Point {
    if !isWalkable(end) {
        end = m.getNearestFreeNeighbor(start, end, isWalkable)
    }
    //println(fmt.Sprintf("JPS from %v to %v", start, end))
    return m.pathfinder.JPSPath([]geometry.Point{}, start, end, isWalkable, false)
}
func (m *GridMap[ActorType, ItemType, ObjectType]) getNearestFreeNeighbor(origin, pos geometry.Point, isFree func(geometry.Point) bool) geometry.Point {
    dist := math.MaxInt32
    nearest := pos
    for _, neighbor := range m.NeighborsCardinal(pos, isFree) {
        d := geometry.DistanceManhattan(origin, neighbor)
        if d < dist {
            dist = d
            nearest = neighbor
        }
    }
    return nearest
}

func (m *GridMap[ActorType, ItemType, ObjectType]) getCurrentlyPassableNeighbors(pos geometry.Point) []geometry.Point {
    neighbors := geometry.Neighbors{}
    freeNeighbors := neighbors.All(pos, func(p geometry.Point) bool {
        return m.Contains(p) && m.IsCurrentlyPassable(p)
    })
    return freeNeighbors
}
func (m *GridMap[ActorType, ItemType, ObjectType]) IsCurrentlyPassable(p geometry.Point) bool {
    if !m.Contains(p) {
        return false
    }
    return m.IsWalkable(p) && (!m.IsActorAt(p)) //&& !knownAsBlocked
}
func (m *GridMap[ActorType, ItemType, ObjectType]) CurrentlyPassableAndSafeForActor(person ActorType) func(p geometry.Point) bool {
    return func(p geometry.Point) bool {
        if !m.Contains(p) ||
            (m.IsActorAt(p) && m.ActorAt(p) != person) {
            return false
        }
        return m.IsWalkableFor(p, person) && !m.IsObviousHazardAt(p)
    }
}
func (m *GridMap[ActorType, ItemType, ObjectType]) IsWalkable(p geometry.Point) bool {
    if !m.Contains(p) {
        return false
    }
    var noActor ActorType
    if m.IsObjectAt(p) && (!m.ObjectAt(p).IsWalkable(noActor)) {
        return false
    }
    cellAt := m.GetCell(p)
    return cellAt.TileType.IsWalkable
}
func (m *GridMap[ActorType, ItemType, ObjectType]) IsObviousHazardAt(p geometry.Point) bool {
    return m.IsLethalTileAt(p)
}
func (m *GridMap[ActorType, ItemType, ObjectType]) IsWalkableFor(p geometry.Point, person ActorType) bool {
    if !m.Contains(p) {
        return false
    }

    if m.IsActorAt(p) && m.ActorAt(p) != person {
        return false
    }

    if m.noClip {
        return true
    }

    if m.IsObjectAt(p) && (!m.ObjectAt(p).IsWalkable(person)) {
        return false
    }

    cellAt := m.GetCell(p)
    return cellAt.TileType.IsWalkable

}
func (m *GridMap[ActorType, ItemType, ObjectType]) IsInHostileZone(person ActorType) bool {
    ourPos := person.Pos()
    zoneAt := m.ZoneAt(ourPos)
    if zoneAt == nil || zoneAt.IsPublic() {
        return false
    }
    return zoneAt.IsHighSecurity()
}

func (m *GridMap[ActorType, ItemType, ObjectType]) CurrentlyPassableForActor(person ActorType) func(p geometry.Point) bool {
    return func(p geometry.Point) bool {
        if !m.Contains(p) ||
            (m.IsActorAt(p) && m.ActorAt(p) != person) {
            return false
        }
        return m.IsWalkableFor(p, person)
    }
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsExplored(pos geometry.Point) bool {
    return m.GetCell(pos).IsExplored
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetExplored(pos geometry.Point) {
    m.Cells[pos.X+pos.Y*m.MapWidth].IsExplored = true
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetAllSpecialTilePositions(tile SpecialTileType) []geometry.Point {
    result := make([]geometry.Point, 0)
    for index, c := range m.Cells {
        if c.TileType.Special == tile {
            x := index % m.MapWidth
            y := index / m.MapWidth
            result = append(result, geometry.Point{X: x, Y: y})
        }
    }
    return result
}
func (m *GridMap[ActorType, ItemType, ObjectType]) GetNeighborWithSpecial(pos geometry.Point, specialType SpecialTileType) geometry.Point {
    neighbors := m.GetAllCardinalNeighbors(pos)
    for _, n := range neighbors {
        if m.GetCell(n).TileType.Special == specialType {
            return n
        }
    }
    return pos
}
func (m *GridMap[ActorType, ItemType, ObjectType]) GetNearestSpecialTile(pos geometry.Point, tile SpecialTileType) geometry.Point {
    result := pos
    allPositions := m.GetAllSpecialTilePositions(tile)

    m.pathfinder.DijkstraMap(m.currentlyPassablePather(), []geometry.Point{pos}, 100)

    sort.Slice(allPositions, func(i, j int) bool {
        return m.pathfinder.DijkstraMapAt(allPositions[i]) < m.pathfinder.DijkstraMapAt(allPositions[j])
    })
    if len(allPositions) > 0 {
        result = allPositions[0]
    }
    return result
}

func (m *GridMap[ActorType, ItemType, ObjectType]) currentlyPassablePather() MapPather {
    return MapPather{
        allNeighbors:      m.getCurrentlyPassableNeighbors,
        neighborPredicate: func(pos geometry.Point) bool { return true },
        pathCostFunc:      func(from geometry.Point, to geometry.Point) int { return 1 },
    }
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SwapDownedPositions(downedActorOne ActorType, downedActorTwo ActorType) {
    posTwo := downedActorTwo.Pos()
    posOne := downedActorOne.Pos()
    downedActorOne.SetPos(posTwo)
    downedActorTwo.SetPos(posOne)
    m.Cells[posOne.X+posOne.Y*m.MapWidth] = m.Cells[posOne.X+posOne.Y*m.MapWidth].WithDownedActor(downedActorTwo)
    m.Cells[posTwo.X+posTwo.Y*m.MapWidth] = m.Cells[posTwo.X+posTwo.Y*m.MapWidth].WithDownedActor(downedActorOne)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SwapPositions(actorOne ActorType, actorTwo ActorType) {
    posTwo := actorTwo.Pos()
    posOne := actorOne.Pos()
    actorOne.SetPos(posTwo)
    actorTwo.SetPos(posOne)
    m.Cells[posOne.X+posOne.Y*m.MapWidth] = m.Cells[posOne.X+posOne.Y*m.MapWidth].WithActor(actorTwo)
    m.Cells[posTwo.X+posTwo.Y*m.MapWidth] = m.Cells[posTwo.X+posTwo.Y*m.MapWidth].WithActor(actorOne)
}

func NewZoneMap(zone *ZoneInfo, width int, height int) []*ZoneInfo {
    zoneMap := make([]*ZoneInfo, width*height)
    for i := 0; i < width*height; i++ {
        zoneMap[i] = zone
    }
    return zoneMap
}

type MapPather struct {
    neighborPredicate func(pos geometry.Point) bool
    allNeighbors      func(pos geometry.Point) []geometry.Point
    pathCostFunc      func(from geometry.Point, to geometry.Point) int
}

func (m MapPather) Neighbors(point geometry.Point) []geometry.Point {
    neighbors := make([]geometry.Point, 0)
    for _, p := range m.allNeighbors(point) {
        if m.neighborPredicate(p) {
            neighbors = append(neighbors, p)
        }
    }
    return neighbors
}
func (m MapPather) Cost(from geometry.Point, to geometry.Point) int {
    return m.pathCostFunc(from, to)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsTileWithSpecialAt(pos geometry.Point, special SpecialTileType) bool {
    return m.GetCell(pos).TileType.Special == special
}

func (m *GridMap[ActorType, ItemType, ObjectType]) MoveDownedActor(actor ActorType, newPos geometry.Point) {
    if m.Cells[newPos.Y*m.MapWidth+newPos.X].DownedActor != nil {
        return
    }
    m.Cells[actor.Pos().Y*m.MapWidth+actor.Pos().X] = m.Cells[actor.Pos().Y*m.MapWidth+actor.Pos().X].WithDownedActorHereRemoved(actor)
    actor.SetPos(newPos)
    m.Cells[newPos.Y*m.MapWidth+newPos.X] = m.Cells[newPos.Y*m.MapWidth+newPos.X].WithDownedActor(actor)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveDownedActor(actor ActorType) bool {
    m.Cells[actor.Pos().Y*m.MapWidth+actor.Pos().X] = m.Cells[actor.Pos().Y*m.MapWidth+actor.Pos().X].WithDownedActorHereRemoved(actor)
    for i := len(m.AllDownedActors) - 1; i >= 0; i-- {
        if m.AllDownedActors[i] == actor {
            m.AllDownedActors = append(m.AllDownedActors[:i], m.AllDownedActors[i+1:]...)
            return true
        }
    }
    return false
}

func (m *GridMap[ActorType, ItemType, ObjectType]) Apply(f func(cell MapCell[ActorType, ItemType, ObjectType]) MapCell[ActorType, ItemType, ObjectType]) {
    for i, cell := range m.Cells {
        m.Cells[i] = f(cell)
    }
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetNearestDropOffPosition(pos geometry.Point) geometry.Point {
    nearestLocation := geometry.Point{X: 1, Y: 1}
    shortestDistance := math.MaxInt
    for index, c := range m.ZoneMap {
        xPos := index % m.MapWidth
        yPos := index / m.MapWidth
        curPos := geometry.Point{X: xPos, Y: yPos}
        curDist := geometry.DistanceManhattan(curPos, pos)
        isItemHere := m.IsItemAt(curPos)
        isValidZone := c.IsDropOff()
        if !isItemHere && isValidZone && curDist < shortestDistance {
            nearestLocation = curPos
            shortestDistance = curDist
        }
    }
    return nearestLocation
}

func (m *GridMap[ActorType, ItemType, ObjectType]) FindNearestItem(pos geometry.Point, predicate func(item ItemType) bool) ItemType {
    var nearestItem ItemType
    nearestDistance := math.MaxInt
    for _, item := range m.AllItems {
        if predicate(item) {
            curDist := geometry.DistanceManhattan(item.Pos(), pos)
            if curDist < nearestDistance {
                nearestItem = item
                nearestDistance = curDist
            }
        }
    }
    return nearestItem
}

func (m *GridMap[ActorType, ItemType, ObjectType]) FindAllNearbyActors(source geometry.Point, maxDist int, keep func(actor ActorType) bool) []ActorType {
    results := make([]ActorType, 0)
    for _, actor := range m.AllActors {
        if keep(actor) && geometry.DistanceManhattan(actor.Pos(), source) <= maxDist {
            results = append(results, actor)
        }
    }
    return results
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsNamedLocationAt(positionInWorld geometry.Point) bool {
    for _, loc := range m.NamedLocations {
        if loc == positionInWorld {
            return true
        }
    }
    return false
}
func (m *GridMap[ActorType, ItemType, ObjectType]) ZoneNames() []string {
    result := make([]string, 0)
    for _, zone := range m.ListOfZones {
        result = append(result, zone.Name)
    }
    return result
}

func (m *GridMap[ActorType, ItemType, ObjectType]) Resize(width int, height int, emptyTile Tile) {
    oldWidth := m.MapWidth
    oldHeight := m.MapHeight

    newCells := make([]MapCell[ActorType, ItemType, ObjectType], width*height)
    newZoneMap := make([]*ZoneInfo, width*height)
    for i := 0; i < width*height; i++ {
        newZoneMap[i] = m.ListOfZones[0]
    }
    for i := 0; i < width*height; i++ {
        newCells[i] = MapCell[ActorType, ItemType, ObjectType]{
            TileType:   emptyTile,
            IsExplored: true,
        }
    }

    // copy over the old cells into the center of the new map
    for y := 0; y < oldHeight; y++ {
        if y >= height {
            break
        }
        for x := 0; x < oldWidth; x++ {
            if x >= width {
                break
            }
            destIndex := y*width + x
            srcIndex := y*oldWidth + x
            newCells[destIndex] = m.Cells[srcIndex]
            newZoneMap[destIndex] = m.ZoneMap[srcIndex]
        }
    }

    m.Cells = newCells
    m.ZoneMap = newZoneMap
    m.MapWidth = width
    m.MapHeight = height
}
func (m *GridMap[ActorType, ItemType, ObjectType]) MapSize() geometry.Point {
    return geometry.Point{X: m.MapWidth, Y: m.MapHeight}
}

func (m *GridMap[ActorType, ItemType, ObjectType]) NeighborsAll(pos geometry.Point, filter func(p geometry.Point) bool) []geometry.Point {
    neighbors := geometry.Neighbors{}
    return neighbors.All(pos, filter)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) NeighborsCardinal(pos geometry.Point, filter func(p geometry.Point) bool) []geometry.Point {
    neighbors := geometry.Neighbors{}
    return neighbors.Cardinal(pos, filter)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetNearestWalkableNeighbor(start geometry.Point, dest geometry.Point) geometry.Point {
    minDist := math.MaxInt
    var minPos geometry.Point
    for _, neighbor := range m.NeighborsCardinal(dest, m.IsWalkable) {
        dist := geometry.DistanceManhattan(neighbor, start)
        if dist < minDist {
            minDist = dist
            minPos = neighbor
        }
    }
    return minPos
}
func (m *GridMap[ActorType, ItemType, ObjectType]) IsPassableForProjectile(p geometry.Point) bool {
    isTileWalkable := m.IsTileWalkable(p)
    isActorOnTile := m.IsActorAt(p)
    isObjectOnTile := m.IsObjectAt(p)
    isObjectBlocking := false
    if isObjectOnTile {
        objectOnTile := m.ObjectAt(p)
        isObjectBlocking = !objectOnTile.IsPassableForProjectile()
    }
    return isTileWalkable && !isActorOnTile && !isObjectBlocking
}
func (m *GridMap[ActorType, ItemType, ObjectType]) LineOfSight(source geometry.Point, target geometry.Point) []geometry.Point {
    var los []geometry.Point
    m.Bresenheim(source, target, func(x, y int) bool {
        visited := geometry.Point{X: x, Y: y}
        if visited == source {
            return true
        }
        los = append(los, visited)
        return m.IsPassableForProjectile(visited)
    })
    return los
}
func (m *GridMap[ActorType, ItemType, ObjectType]) Bresenheim(source geometry.Point, target geometry.Point, visitor func(x, y int) bool) {
    var dx, dy, e, slope int
    x1, y1 := target.X, target.Y
    x2, y2 := source.X, source.Y
    // Because drawing p1 -> p2 is equivalent to draw p2 -> p1,
    // I sort points in x-axis order to handle only half of possible cases.
    if x1 > x2 {
        x1, y1, x2, y2 = x2, y2, x1, y1
    }

    dx, dy = x2-x1, y2-y1
    // Because point is x-axis ordered, dx cannot be negative
    if dy < 0 {
        dy = -dy
    }

    switch {

    // Is line a point ?
    case x1 == x2 && y1 == y2:
        visitor(x1, y1)

    // Is line an horizontal ?
    case y1 == y2:
        for ; dx != 0; dx-- {
            if !visitor(x1, y1) {
                return
            }
            x1++
        }
        if !visitor(x1, y1) {
            return
        }

    // Is line a vertical ?
    case x1 == x2:
        if y1 > y2 {
            y1, y2 = y2, y1
        }
        for ; dy != 0; dy-- {
            if !visitor(x1, y1) {
                return
            }
            y1++
        }
        if !visitor(x1, y1) {
            return
        }

    // Is line a diagonal ?
    case dx == dy:
        if y1 < y2 {
            for ; dx != 0; dx-- {
                if !visitor(x1, y1) {
                    return
                }
                x1++
                y1++
            }
        } else {
            for ; dx != 0; dx-- {
                if !visitor(x1, y1) {
                    return
                }
                x1++
                y1--
            }
        }
        if !visitor(x1, y1) {
            return
        }

    // wider than high ?
    case dx > dy:
        if y1 < y2 {
            // BresenhamDxXRYD(img, x1, y1, x2, y2, col)
            dy, e, slope = 2*dy, dx, 2*dx
            for ; dx != 0; dx-- {
                if !visitor(x1, y1) {
                    return
                }
                x1++
                e -= dy
                if e < 0 {
                    y1++
                    e += slope
                }
            }
        } else {
            // BresenhamDxXRYU(img, x1, y1, x2, y2, col)
            dy, e, slope = 2*dy, dx, 2*dx
            for ; dx != 0; dx-- {
                if !visitor(x1, y1) {
                    return
                }
                x1++
                e -= dy
                if e < 0 {
                    y1--
                    e += slope
                }
            }
        }
        if !visitor(x2, y2) {
            return
        }

    // higher than wide.
    default:
        if y1 < y2 {
            // BresenhamDyXRYD(img, x1, y1, x2, y2, col)
            dx, e, slope = 2*dx, dy, 2*dy
            for ; dy != 0; dy-- {
                if !visitor(x1, y1) {
                    return
                }
                y1++
                e -= dx
                if e < 0 {
                    x1++
                    e += slope
                }
            }
        } else {
            // BresenhamDyXRYU(img, x1, y1, x2, y2, col)
            dx, e, slope = 2*dx, dy, 2*dy
            for ; dy != 0; dy-- {
                if !visitor(x1, y1) {
                    return
                }
                y1--
                e -= dx
                if e < 0 {
                    x1++
                    e += slope
                }
            }
        }
        if !visitor(x2, y2) {
            return
        }
    }
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetAllExplored() {
    for y := 0; y < m.MapHeight; y++ {
        for x := 0; x < m.MapWidth; x++ {
            m.Cells[y*m.MapWidth+x].IsExplored = true
        }
    }
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RandomSpawnPosition() geometry.Point {
    for {
        x := rand.Intn(m.MapWidth)
        y := rand.Intn(m.MapHeight)
        pos := geometry.Point{X: x, Y: y}
        if m.IsCurrentlyPassable(pos) {
            return pos
        }
    }
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetNamedLocation(name string, point geometry.Point) {
    m.NamedLocations[name] = point
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetNamedLocation(name string) geometry.Point {
    return m.NamedLocations[name]
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetNamedLocationByPos(pos geometry.Point) string {
    for name, location := range m.NamedLocations {
        if location == pos {
            return name
        }
    }
    return ""
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RenameLocation(oldName string, newName string) {
    if pos, ok := m.NamedLocations[oldName]; ok {
        delete(m.NamedLocations, oldName)
        m.NamedLocations[newName] = pos
    }
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveNamedLocation(namedLocation string) {
    delete(m.NamedLocations, namedLocation)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsLethalTileAt(p geometry.Point) bool {
    return m.CellAt(p).TileType.Special == SpecialTileLethal
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RandomPosAround(pos geometry.Point) geometry.Point {
    neighbors := m.NeighborsAll(pos, func(p geometry.Point) bool {
        return m.Contains(p)
    })
    if len(neighbors) == 0 {
        return pos
    }
    neighbors = append(neighbors, pos)
    return neighbors[rand.Intn(len(neighbors))]
}

func (m *GridMap[ActorType, ItemType, ObjectType]) TryGetActorAt(pos geometry.Point) (ActorType, bool) {
    var noActor ActorType
    isActorAt := m.IsActorAt(pos)
    if !isActorAt {
        return noActor, false
    }
    return m.ActorAt(pos), isActorAt
}

func (m *GridMap[ActorType, ItemType, ObjectType]) TryGetDownedActorAt(pos geometry.Point) (ActorType, bool) {
    var noActor ActorType
    isDownedActorAt := m.IsDownedActorAt(pos)
    if !isDownedActorAt {
        return noActor, false
    }
    return m.DownedActorAt(pos), isDownedActorAt
}

func (m *GridMap[ActorType, ItemType, ObjectType]) TryGetObjectAt(pos geometry.Point) (ObjectType, bool) {
    var noObject ObjectType
    isObjectAt := m.IsObjectAt(pos)
    if !isObjectAt {
        return noObject, false
    }
    return m.ObjectAt(pos), isObjectAt
}

func (m *GridMap[ActorType, ItemType, ObjectType]) TryGetItemAt(pos geometry.Point) (ItemType, bool) {
    var noItem ItemType
    isItemAt := m.IsItemAt(pos)
    if !isItemAt {
        return noItem, false
    }
    return m.ItemAt(pos), isItemAt
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddActor(actor ActorType, spawnPos geometry.Point) {
    m.AllActors = append(m.AllActors, actor)
    m.MoveActor(actor, spawnPos)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddDownedActor(actor ActorType, spawnPos geometry.Point) {
    m.AllDownedActors = append(m.AllDownedActors, actor)
    m.MoveDownedActor(actor, spawnPos)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddObject(object ObjectType, spawnPos geometry.Point) {
    m.AllObjects = append(m.AllObjects, object)
    m.MoveObject(object, spawnPos)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddItem(item ItemType, spawnPos geometry.Point) {
    m.AllItems = append(m.AllItems, item)
    m.MoveItem(item, spawnPos)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) UpdateFieldOfView(fov *geometry.FOV, fovPosition geometry.Point) {
    visionRange := 19 // 10
    visionRangeSquared := visionRange * visionRange

    var fovRange = geometry.NewRect(-visionRange, -visionRange, visionRange+1, visionRange+1)
    fov.SetRange(fovRange.Add(fovPosition).Intersect(geometry.NewRect(0, 0, m.MapWidth, m.MapHeight)))

    visionMap := fov.SSCVisionMap(fovPosition, visionRange, func(p geometry.Point) bool {
        return m.IsTransparent(p) && geometry.DistanceSquared(p, fovPosition) <= visionRangeSquared
    }, false)

    for _, p := range visionMap {
        if !m.IsExplored(p) {
            m.SetExplored(p)
        }
    }
}

func (m *GridMap[ActorType, ItemType, ObjectType]) ToggleNoClip() bool {
    m.noClip = !m.noClip
    return m.noClip
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetSecretDoorAt(pos geometry.Point) {
    m.secretDoors[pos] = true
}

func (m *GridMap[ActorType, ItemType, ObjectType]) IsSecretDoorAt(neighbor geometry.Point) bool {
    if _, ok := m.secretDoors[neighbor]; ok {
        return true
    }
    return false
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetName(name string) {
    m.name = name
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetTransitionAt(pos geometry.Point, transition Transition) {
    m.transitionMap[pos] = transition
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetTransitionAt(pos geometry.Point) (Transition, bool) {
    transition, ok := m.transitionMap[pos]
    return transition, ok

}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetName() string {
    return m.name
}

func (m *GridMap[ActorType, ItemType, ObjectType]) WriteTiles(out io.Writer) {
    for _, cell := range m.Cells {
        cell.TileType.ToBinary(out)
    }
}

func (m *GridMap[ActorType, ItemType, ObjectType]) ReadTiles(in io.Reader) {
    for i, _ := range m.Cells {
        m.Cells[i].TileType = NewTileFromBinary(in)
    }
}

func (m *GridMap[ActorType, ItemType, ObjectType]) Transitions() map[geometry.Point]Transition {
    return m.transitionMap
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SecretDoors() map[geometry.Point]bool {
    return m.secretDoors
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddNamedRegion(name string, region geometry.Rect) {
    m.namedRects[name] = region
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetNamedRegion(name string) geometry.Rect {
    return m.namedRects[name]
}

type Trigger struct {
    Name    string
    Bounds  geometry.Rect
    OneShot bool
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetNamedTriggerAt(pos geometry.Point) (Trigger, bool) {
    for _, trigger := range m.namedTrigger {
        if trigger.Bounds.Contains(pos) {
            return trigger, true
        }
    }
    return Trigger{}, false
}
func (m *GridMap[ActorType, ItemType, ObjectType]) SetTileIcon(pos geometry.Point, index int32) {
    m.Cells[pos.Y*m.MapWidth+pos.X].TileType.DefinedIcon = index
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetTileIconAt(pos geometry.Point) int32 {
    return m.Cells[pos.Y*m.MapWidth+pos.X].TileType.DefinedIcon
}

func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveNamedRegion(regionName string) {
    delete(m.namedRects, regionName)
}
func (m *GridMap[ActorType, ItemType, ObjectType]) RemoveNamedTrigger(triggerName string) {
    delete(m.namedTrigger, triggerName)
}

func (m *GridMap[ActorType, ItemType, ObjectType]) AddNamedTrigger(name string, rect Trigger) {
    m.namedTrigger[name] = rect
}

func (m *GridMap[ActorType, ItemType, ObjectType]) SetDisplayName(name string) {
    m.displayName = name
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetDisplayName() string {
    return m.displayName
}

func (m *GridMap[ActorType, ItemType, ObjectType]) GetFilteredActorsInRadius(location geometry.Point, radius int, filter func(actor ActorType) bool) []ActorType {
    if (radius * radius) < len(m.AllActors) {
        return m.iterateMapForActors(location, radius, filter)
    } else {
        return m.iterateActors(location, radius, filter)
    }
}
func (m *GridMap[ActorType, ItemType, ObjectType]) iterateActors(location geometry.Point, radius int, filter func(actor ActorType) bool) []ActorType {
    result := make([]ActorType, 0)
    for _, actor := range m.AllActors {
        if geometry.DistanceManhattan(location, actor.Pos()) <= radius && filter(actor) {
            result = append(result, actor)
        }
    }
    return result
}

func (m *GridMap[ActorType, ItemType, ObjectType]) iterateMapForActors(location geometry.Point, radius int, filter func(actor ActorType) bool) []ActorType {
    result := make([]ActorType, 0)
    for y := location.Y - radius; y <= location.Y+radius; y++ {
        for x := location.X - radius; x <= location.X+radius; x++ {
            pos := geometry.Point{X: x, Y: y}
            if m.IsActorAt(pos) {
                actorAt := m.ActorAt(pos)
                if filter(actorAt) {
                    result = append(result, actorAt)
                }
            }
        }
    }
    return result
}
