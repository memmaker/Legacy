package main

import (
    "Legacy/geometry"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
)

func (g *GridEngine) mapLookup(x, y int, tick uint64) (*ebiten.Image, int32, color.Color) {
    location := geometry.Point{X: x, Y: y}
    if !g.currentMap.Contains(location) {
        return g.worldTiles, 0, color.Black
    }

    if g.playerParty.CanSee(location) {
        // can see this tile, so we lookup dynamic entities
        if g.currentMap.IsActorAt(location) {
            actorAt := g.currentMap.GetActor(location)
            if !actorAt.IsHidden() {
                if actorAt.IsTinted() {
                    return g.grayScaleEntityTiles, actorAt.Icon(tick), actorAt.TintColor()
                }
                return g.entityTiles, actorAt.Icon(tick), color.White
            }
        }

        if g.currentMap.IsDownedActorAt(location) {
            actorAt, _ := g.currentMap.TryGetDownedActorAt(location)
            if !actorAt.IsHidden() {
                return g.entityTiles, actorAt.Icon(tick), color.White
            }
        }

        if g.currentMap.IsItemAt(location) {
            itemAt := g.currentMap.ItemAt(location)
            if !itemAt.IsHidden() {
                return g.entityTiles, itemAt.Icon(tick), itemAt.TintColor()
            }
        }
    }

    if g.currentMap.IsObjectAt(location) {
        objectAt := g.currentMap.ObjectAt(location)
        if !objectAt.IsHidden() {
            return g.entityTiles, objectAt.Icon(tick), objectAt.TintColor()
        }
    }
    tile := g.currentMap.GetCell(location)

    return g.worldTiles, tile.TileType.DefinedIcon, color.White
}

func (g *GridEngine) TotalScale() float64 {
    return g.tileScale * g.deviceDPIScale
}
func (g *GridEngine) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
    panic("should use layoutf")
}

func (g *GridEngine) LayoutF(outsideWidth, outsideHeight float64) (screenWidth, screenHeight float64) {
    g.deviceDPIScale = ebiten.DeviceScaleFactor()
    totalScale := g.tileScale * g.deviceDPIScale
    return float64(g.internalWidth) * totalScale, float64(g.internalHeight) * totalScale
}

func (g *GridEngine) drawUIOverlay(screen *ebiten.Image) {
    screenW := g.gridRenderer.GetSmallGridScreenSize().X // in 8x8 cells
    for cellIndex, tileIndex := range g.uiOverlay {
        cellX := cellIndex % screenW
        cellY := cellIndex / screenW
        g.gridRenderer.DrawOnSmallGrid(screen, cellX, cellY, tileIndex)
    }
}
