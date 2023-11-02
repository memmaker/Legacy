package renderer

import (
    "Legacy/geometry"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
)

type MapView interface {
    GetScreenOffset() geometry.Point
    GetTextureIndexAt(x, y int, tick uint64) (*ebiten.Image, int32, color.Color)
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

func (r *MapRenderer) Draw(fov *geometry.FOV, screen *ebiten.Image, tick uint64) {
    screenOffset := r.input.GetScreenOffset()
    tileCountX, tileCountY := r.input.GetWindowSizeInCells()

    scrollOffset := r.input.GetScrollOffset()

    for yOff := 0; yOff < tileCountY; yOff++ {
        for xOff := 0; xOff < tileCountX; xOff++ {
            mapX := xOff + scrollOffset.X
            mapY := yOff + scrollOffset.Y
            x := (xOff)*r.gridRenderer.GetScaledBigGridSize() + int(float64(screenOffset.X)*r.gridRenderer.GetScale())
            y := (yOff)*r.gridRenderer.GetScaledBigGridSize() + int(float64(screenOffset.Y)*r.gridRenderer.GetScale())
            textureAtlas, textureIndex, tintColor := r.input.GetTextureIndexAt(mapX, mapY, tick)
            if textureIndex == -1 {
                continue
            }
            if !fov.Visible(geometry.Point{X: mapX, Y: mapY}) {
                tintColor = color.Black
            }
            r.gridRenderer.DrawBigOnScreenWithAtlasAndTint(screen, float64(x), float64(y), textureAtlas, textureIndex, tintColor)
        }
    }
}
