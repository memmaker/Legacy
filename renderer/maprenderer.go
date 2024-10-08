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
    input              MapView
    gridRenderer       *DualGridRenderer
    disableFieldOfView func() bool
}

func NewRenderer(gridRenderer *DualGridRenderer, input MapView) *MapRenderer {
    return &MapRenderer{
        input:        input,
        gridRenderer: gridRenderer,
        disableFieldOfView: func() bool {
            return false
        },
    }
}

func (r *MapRenderer) Draw(fov *geometry.FOV, screen *ebiten.Image, tick uint64, isExplored func(pos geometry.Point) bool) {
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
            mapPos := geometry.Point{X: mapX, Y: mapY}
            if !fov.Visible(mapPos) && !r.disableFieldOfView() {
                if isExplored(mapPos) {
                    tintColor = color.RGBA{
                        R: 80,
                        G: 80,
                        B: 80,
                        A: 255,
                    }
                } else {
                    tintColor = color.Black
                }
            }

            r.gridRenderer.DrawBigOnScreenWithAtlasAndTint(screen, float64(x), float64(y), textureAtlas, textureIndex, tintColor)
        }
    }
}

func (r *MapRenderer) SetDisableFieldOfView(shouldDisableFoV func() bool) {
    r.disableFieldOfView = shouldDisableFoV
}
