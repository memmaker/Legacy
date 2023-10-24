package renderer

import (
    "Legacy/geometry"
    "github.com/hajimehoshi/ebiten/v2"
)

type MapView interface {
    GetScreenOffset() geometry.Point
    GetTextureIndexAt(x, y int) (*ebiten.Image, int)
    GetScrollOffset() geometry.Point
    GetWindowSizeInCells() (int, int)
}

type MapRenderer struct {
    input        MapView
    gridRenderer *DualGridRenderer
}

func NewRenderer(gridRenderer *DualGridRenderer, input MapView) *MapRenderer {
    return &MapRenderer{
        input:        input,
        gridRenderer: gridRenderer,
    }
}

func (r *MapRenderer) Draw(screen *ebiten.Image) {
    screenOffset := r.input.GetScreenOffset()
    tileCountX, tileCountY := r.input.GetWindowSizeInCells()

    scrollOffset := r.input.GetScrollOffset()

    for yOff := 0; yOff < tileCountY; yOff++ {
        for xOff := 0; xOff < tileCountX; xOff++ {
            mapX := xOff + scrollOffset.X
            mapY := yOff + scrollOffset.Y
            x := (xOff)*r.gridRenderer.GetScaledBigGridSize() + int(float64(screenOffset.X)*r.gridRenderer.GetScale())
            y := (yOff)*r.gridRenderer.GetScaledBigGridSize() + int(float64(screenOffset.Y)*r.gridRenderer.GetScale())
            textureAtlas, textureIndex := r.input.GetTextureIndexAt(mapX, mapY)
            if textureIndex == -1 {
                continue
            }
            r.gridRenderer.DrawBigOnScreenWithAtlas(screen, float64(x), float64(y), textureAtlas, textureIndex)
        }
    }
}
