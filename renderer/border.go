package renderer

import "Legacy/geometry"

type GridBorderDefinition struct {
    HorizontalLineTextureIndex int
    VerticalLineTextureIndex   int
    CornerTextureIndex         int
    BackgroundTextureIndex     int
}

func NewBorderMap(topLeft, bottomRight geometry.Point, borderDef GridBorderDefinition) map[geometry.Point]int {
    result := make(map[geometry.Point]int)
    for y := topLeft.Y; y < bottomRight.Y; y++ {
        for x := topLeft.X; x < bottomRight.X; x++ {
            // corner
            if x == topLeft.X && y == topLeft.Y || x == bottomRight.X-1 && y == topLeft.Y || x == topLeft.X && y == bottomRight.Y-1 || x == bottomRight.X-1 && y == bottomRight.Y-1 {
                result[geometry.Point{X: x, Y: y}] = borderDef.CornerTextureIndex
                continue
            } else if x == topLeft.X || x == bottomRight.X-1 {
                result[geometry.Point{X: x, Y: y}] = borderDef.VerticalLineTextureIndex
                continue
            } else if y == topLeft.Y || y == bottomRight.Y-1 {
                result[geometry.Point{X: x, Y: y}] = borderDef.HorizontalLineTextureIndex
                continue
            }
        }
    }
    return result
}

type BorderCase int

func (c BorderCase) GetIndex(def GridBorderDefinition) int {
    switch c {
    case Corner:
        return def.CornerTextureIndex
    case HorizontalLine:
        return def.HorizontalLineTextureIndex
    case VerticalLine:
        return def.VerticalLineTextureIndex
    case Background:
        return def.BackgroundTextureIndex
    }
    return -1
}

const (
    Corner BorderCase = iota
    HorizontalLine
    VerticalLine
    Background
)

func BorderTraversal(topLeft, bottomRight geometry.Point, traverse func(point geometry.Point, elem BorderCase)) {
    for y := topLeft.Y; y < bottomRight.Y; y++ {
        for x := topLeft.X; x < bottomRight.X; x++ {
            // corner
            if x == topLeft.X && y == topLeft.Y || x == bottomRight.X-1 && y == topLeft.Y || x == topLeft.X && y == bottomRight.Y-1 || x == bottomRight.X-1 && y == bottomRight.Y-1 {
                traverse(geometry.Point{X: x, Y: y}, Corner)
                continue
            } else if x == topLeft.X || x == bottomRight.X-1 {
                traverse(geometry.Point{X: x, Y: y}, VerticalLine)
                continue
            } else if y == topLeft.Y || y == bottomRight.Y-1 {
                traverse(geometry.Point{X: x, Y: y}, HorizontalLine)
                continue
            }
        }
    }
}
