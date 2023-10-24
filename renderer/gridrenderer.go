package renderer

import (
    "Legacy/geometry"
    "github.com/hajimehoshi/ebiten/v2"
    "image"
    "image/color"
)

type DualGridRenderer struct {
    smallAtlas          *ebiten.Image
    bigAtlas            *ebiten.Image
    smallGridSize       int
    bigGridSize         int
    scale               float64
    fontMap             map[rune]uint16
    smallGridScreenSize geometry.Point
    op                  *ebiten.DrawImageOptions
    borderDef           GridBorderDefinition
}

func NewDualGridRenderer(smallAtlas *ebiten.Image, bigAtlas *ebiten.Image, scale float64, fontMap map[rune]uint16) *DualGridRenderer {
    return &DualGridRenderer{
        smallAtlas:          smallAtlas,
        bigAtlas:            bigAtlas,
        scale:               scale,
        fontMap:             fontMap,
        smallGridSize:       8,
        bigGridSize:         16,
        smallGridScreenSize: geometry.Point{X: 40, Y: 25},
        op:                  &ebiten.DrawImageOptions{},
    }
}
func (g *DualGridRenderer) GetScaledBigGridSize() int {
    return int(float64(g.bigGridSize) * g.scale)
}

func (g *DualGridRenderer) GetScaledSmallGridSize() int {
    return int(float64(g.smallGridSize) * g.scale)
}
func (g *DualGridRenderer) SetBorderDefinition(borderDef GridBorderDefinition) {
    g.borderDef = borderDef
}

func (g *DualGridRenderer) BigCellToScreen(cellX, cellY int) (float64, float64) {
    return float64(cellX*g.bigGridSize) * g.scale, float64(cellY*g.bigGridSize) * g.scale
}

func (g *DualGridRenderer) SmallCellToScreen(cellX, cellY int) (float64, float64) {
    return float64(cellX*g.smallGridSize) * g.scale, float64(cellY*g.smallGridSize) * g.scale
}

func (g *DualGridRenderer) DrawOnSmallGrid(screen *ebiten.Image, cellX, cellY int, textureIndex int) {
    g.op.ColorScale.Reset()
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(g.scale, g.scale)
    g.op.GeoM.Translate(g.SmallCellToScreen(cellX, cellY))
    screen.DrawImage(ExtractSubImageFromAtlas(textureIndex, g.smallGridSize, g.smallGridSize, g.smallAtlas), g.op)
}

func (g *DualGridRenderer) DrawMappingOnSmallGrid(screen *ebiten.Image, gridMapping map[geometry.Point]int) {
    for point, textureIndex := range gridMapping {
        g.DrawOnSmallGrid(screen, point.X, point.Y, textureIndex)
    }
}

func (g *DualGridRenderer) DrawColoredChar(screen *ebiten.Image, cellX, cellY int, char rune, textColor color.Color) {
    textureIndex, ok := g.fontMap[char]
    if !ok {
        return
    }
    g.op.ColorScale.Reset()
    g.op.ColorScale.ScaleWithColor(textColor)
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(g.scale, g.scale)
    g.op.GeoM.Translate(g.SmallCellToScreen(cellX, cellY))
    screen.DrawImage(ExtractSubImageFromAtlas(int(textureIndex), g.smallGridSize, g.smallGridSize, g.smallAtlas), g.op)
}

func (g *DualGridRenderer) DrawColoredString(screen *ebiten.Image, cellX, cellY int, text string, textColor color.Color) {
    for i, char := range text {
        g.DrawColoredChar(screen, cellX+i, cellY, char, textColor)
    }
}

func (g *DualGridRenderer) DrawSmallOnScreen(screen *ebiten.Image, xPos, yPos float64, textureIndex int) {
    g.op.ColorScale.Reset()
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(g.scale, g.scale)
    g.op.GeoM.Translate(xPos, yPos)
    screen.DrawImage(ExtractSubImageFromAtlas(textureIndex, g.smallGridSize, g.smallGridSize, g.smallAtlas), g.op)
}

func (g *DualGridRenderer) DrawBorder(screen *ebiten.Image, topLeft, bottomRight geometry.Point) {
    BorderTraversal(topLeft, bottomRight, func(p geometry.Point, borderType BorderCase) {
        textureIndex := borderType.GetIndex(g.borderDef)
        g.DrawOnSmallGrid(screen, p.X, p.Y, textureIndex)
    })
}

func (g *DualGridRenderer) DrawFilledBorder(screen *ebiten.Image, topLeft, bottomRight geometry.Point) {
    BorderTraversal(topLeft, bottomRight, func(p geometry.Point, borderType BorderCase) {
        textureIndex := borderType.GetIndex(g.borderDef)
        g.DrawOnSmallGrid(screen, p.X, p.Y, textureIndex)
    })

    width := bottomRight.X - topLeft.X - 2
    height := bottomRight.Y - topLeft.Y - 2

    backgroundIndex := g.borderDef.BackgroundTextureIndex

    subImage := ExtractSubImageFromAtlas(backgroundIndex, g.smallGridSize, g.smallGridSize, g.smallAtlas)

    g.op.ColorScale.Reset()
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(g.scale*float64(width), g.scale*float64(height))
    g.op.GeoM.Translate(g.SmallCellToScreen(topLeft.X+1, topLeft.Y+1))
    screen.DrawImage(subImage, g.op)
}

func (g *DualGridRenderer) DrawOnBigGrid(screen *ebiten.Image, cellX, cellY int, textureIndex int) {
    g.op.ColorScale.Reset()
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(g.scale, g.scale)
    g.op.GeoM.Translate(g.BigCellToScreen(cellX, cellY))
    screen.DrawImage(ExtractSubImageFromAtlas(textureIndex, g.bigGridSize, g.bigGridSize, g.bigAtlas), g.op)
}

func (g *DualGridRenderer) DrawBigOnScreen(screen *ebiten.Image, xPos, yPos float64, textureIndex int) {
    g.op.ColorScale.Reset()
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(g.scale, g.scale)
    g.op.GeoM.Translate(xPos, yPos)
    screen.DrawImage(ExtractSubImageFromAtlas(textureIndex, g.bigGridSize, g.bigGridSize, g.bigAtlas), g.op)
}

func (g *DualGridRenderer) DrawBigOnScreenWithAtlas(screen *ebiten.Image, xPos, yPos float64, atlas *ebiten.Image, textureIndex int) {
    g.op.ColorScale.Reset()
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(g.scale, g.scale)
    g.op.GeoM.Translate(xPos, yPos)
    screen.DrawImage(ExtractSubImageFromAtlas(textureIndex, g.bigGridSize, g.bigGridSize, atlas), g.op)
}

func (g *DualGridRenderer) GetSmallGridScreenSize() geometry.Point {
    return g.smallGridScreenSize
}

func (g *DualGridRenderer) GetScale() float64 {
    return g.scale
}

func ExtractSubImageFromAtlas(textureIndex int, tileSizeX int, tileSizeY int, textureAtlas *ebiten.Image) *ebiten.Image {
    atlasItemCountX := textureAtlas.Bounds().Size().X / tileSizeX
    textureRectTopLeft := image.Point{
        X: (textureIndex % atlasItemCountX) * tileSizeX,
        Y: (textureIndex / atlasItemCountX) * tileSizeY,
    }
    textureRect := image.Rectangle{
        Min: textureRectTopLeft,
        Max: image.Point{
            X: textureRectTopLeft.X + tileSizeX,
            Y: textureRectTopLeft.Y + tileSizeY,
        },
    }

    tileImage := textureAtlas.SubImage(textureRect).(*ebiten.Image)
    return tileImage
}
