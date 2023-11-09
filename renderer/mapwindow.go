package renderer

import (
    "Legacy/geometry"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
)

type MapWindow struct {
    lookup       func(x, y int, tick uint64) (*ebiten.Image, int32, color.Color)
    screenOffset geometry.Point
    windowSize   geometry.Point
    mapSize      geometry.Point
    scrollOffset geometry.Point
}

func NewMapWindow(
    screenOffset geometry.Point,
    windowSize geometry.Point,
    mapSize geometry.Point,
    lookup func(x, y int, tick uint64) (*ebiten.Image, int32, color.Color),
) *MapWindow {
    return &MapWindow{
        lookup:       lookup,
        screenOffset: screenOffset,
        windowSize:   windowSize,
        mapSize:      mapSize,
    }
}

func (m *MapWindow) GetWindowSizeInCells() (int, int) {
    return m.windowSize.X, m.windowSize.Y
}

func (m *MapWindow) GetScreenOffset() geometry.Point {
    return m.screenOffset
}

func (m *MapWindow) GetTextureIndexAt(cellX, cellY int, tick uint64) (*ebiten.Image, int32, color.Color) {
    return m.lookup(cellX, cellY, tick)
}

func (m *MapWindow) GetScrollOffset() geometry.Point {
    return m.scrollOffset
}

func (m *MapWindow) ScrollBy(point geometry.Point) {
    newScrollX := m.scrollOffset.X + point.X
    newScrollY := m.scrollOffset.Y + point.Y
    m.setScrollOffset(newScrollX, newScrollY)
}

func (m *MapWindow) CenterOn(pos geometry.Point) {
    newScrollX := pos.X - m.windowSize.X/2
    newScrollY := pos.Y - m.windowSize.Y/2
    m.setScrollOffset(newScrollX, newScrollY)
}

func (m *MapWindow) EnsurePositionIsInview(pos geometry.Point, borderSize int) {
    screenMapWidth := m.windowSize.X
    screenMapHeight := m.windowSize.Y

    newPositionOnScreen := m.GetScreenGridPositionFromMapGridPosition(pos)

    moveDelta := geometry.Point{X: 0, Y: 0}
    if newPositionOnScreen.X < borderSize {
        moveDelta.X = newPositionOnScreen.X - borderSize
    } else if newPositionOnScreen.X >= screenMapWidth-borderSize {
        moveDelta.X = newPositionOnScreen.X - (screenMapWidth - borderSize)
    }
    if newPositionOnScreen.Y < borderSize {
        moveDelta.Y = newPositionOnScreen.Y - borderSize
    } else if newPositionOnScreen.Y >= screenMapHeight-borderSize {
        moveDelta.Y = newPositionOnScreen.Y - (screenMapHeight - borderSize)
    }

    if moveDelta.X != 0 || moveDelta.Y != 0 {
        m.ScrollBy(moveDelta)
    }
}

func (m *MapWindow) setScrollOffset(newScrollX, newScrollY int) {
    if newScrollX > m.mapSize.X-m.windowSize.X {
        newScrollX = m.mapSize.X - m.windowSize.X
    } else if newScrollX < 0 {
        newScrollX = 0
    }

    if newScrollY > m.mapSize.Y-m.windowSize.Y {
        newScrollY = m.mapSize.Y - m.windowSize.Y
    } else if newScrollY < 0 {
        newScrollY = 0
    }

    m.scrollOffset = geometry.Point{X: newScrollX, Y: newScrollY}
}

func (m *MapWindow) GetScreenGridPositionFromMapGridPosition(mapGridPosition geometry.Point) geometry.Point {
    return geometry.Point{
        X: mapGridPosition.X - m.scrollOffset.X,
        Y: mapGridPosition.Y - m.scrollOffset.Y,
    }
}

func (m *MapWindow) GetMapGridPositionFromScreenGridPosition(screenGridPosition geometry.Point) geometry.Point {
    return geometry.Point{
        X: screenGridPosition.X + m.scrollOffset.X,
        Y: screenGridPosition.Y + m.scrollOffset.Y,
    }
}

func (m *MapWindow) GetVisibleMap() geometry.Rect {
    return geometry.Rect{
        Min: m.scrollOffset,
        Max: geometry.Point{
            X: m.scrollOffset.X + m.windowSize.X,
            Y: m.scrollOffset.Y + m.windowSize.Y,
        },
    }
}
