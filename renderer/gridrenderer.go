package renderer

import (
    "Legacy/geometry"
    "Legacy/util"
    "github.com/hajimehoshi/ebiten/v2"
    "image"
    "image/color"
    "strings"
)

type AtlasName int

const (
    AtlasCharacters AtlasName = iota
    AtlasWorld
    AtlasEntities
    AtlasEntitiesGrayscale
)

type DualGridRenderer struct {
    atlasMap            map[AtlasName]*ebiten.Image
    smallGridSize       int
    bigGridSize         int
    scale               float64
    smallGridFontMap    map[rune]uint16
    bigGridFontMap      map[rune]uint16
    smallGridScreenSize geometry.Point
    op                  *ebiten.DrawImageOptions
    borderDef           GridBorderDefinition
}

func NewDualGridRenderer(scale float64, fontMapForSmallGrid map[rune]uint16) *DualGridRenderer {
    return &DualGridRenderer{
        atlasMap:            make(map[AtlasName]*ebiten.Image),
        scale:               scale,
        smallGridFontMap:    fontMapForSmallGrid,
        smallGridSize:       8,
        bigGridSize:         16,
        smallGridScreenSize: geometry.Point{X: 40, Y: 25},
        op:                  &ebiten.DrawImageOptions{},
    }
}
func (g *DualGridRenderer) SetAtlasMap(atlasMap map[AtlasName]*ebiten.Image) {
    g.atlasMap = atlasMap
}
func (g *DualGridRenderer) getCharAtlas() *ebiten.Image {
    return g.atlasMap[AtlasCharacters]
}
func (g *DualGridRenderer) GetScaledBigGridSize() int {
    return int(float64(g.bigGridSize) * g.scale)
}

func (g *DualGridRenderer) GetScaledSmallGridSize() int {
    return int(float64(g.smallGridSize) * g.scale)
}
func (g *DualGridRenderer) GetSmallGridSize() int {
    return int(float64(g.smallGridSize))
}
func (g *DualGridRenderer) ScreenToSmallCell(x, y int) (int, int) {
    return x / g.GetScaledSmallGridSize(), y / g.GetScaledSmallGridSize()
}
func (g *DualGridRenderer) ScreenToBigCell(x int, y int) geometry.Point {
    return geometry.Point{X: x / g.GetScaledBigGridSize(), Y: y / g.GetScaledBigGridSize()}
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

func (g *DualGridRenderer) DrawOnSmallGrid(screen *ebiten.Image, cellX, cellY int, textureIndex int32) {
    g.op.ColorScale.Reset()
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(g.scale, g.scale)
    g.op.GeoM.Translate(g.SmallCellToScreen(cellX, cellY))
    screen.DrawImage(ExtractSubImageFromAtlas(textureIndex, g.smallGridSize, g.smallGridSize, g.getCharAtlas()), g.op)
}

func (g *DualGridRenderer) DrawMappingOnSmallGrid(screen *ebiten.Image, gridMapping map[geometry.Point]int32) {
    for point, textureIndex := range gridMapping {
        g.DrawOnSmallGrid(screen, point.X, point.Y, textureIndex)
    }
}

func (g *DualGridRenderer) DrawColoredChar(screen *ebiten.Image, cellX, cellY int, char rune, textColor color.Color) {
    textureIndex, ok := g.smallGridFontMap[char]
    if !ok {
        return
    }
    g.op.ColorScale.Reset()
    g.op.ColorScale.ScaleWithColor(textColor)
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(g.scale, g.scale)
    g.op.GeoM.Translate(g.SmallCellToScreen(cellX, cellY))
    screen.DrawImage(ExtractSubImageFromAtlas(int32(textureIndex), g.smallGridSize, g.smallGridSize, g.getCharAtlas()), g.op)
}

func (g *DualGridRenderer) DrawColoredString(screen *ebiten.Image, cellX, cellY int, text string, textColor color.Color) {
    asRunes := []rune(text)
    for i, char := range asRunes {
        g.DrawColoredChar(screen, cellX+i, cellY, char, textColor)
    }
}

func (g *DualGridRenderer) DrawFilledBorder(screen *ebiten.Image, topLeft, bottomRight geometry.Point, title string) {
    centeredTitleXStart := (bottomRight.X - topLeft.X - len(title)) / 2
    centeredTitleXEnd := centeredTitleXStart + len(title)

    borderFunc := func(p geometry.Point, borderType BorderCase) {
        textureIndex := borderType.GetIndex(g.borderDef)
        relativeX := p.X - topLeft.X
        if len(title) > 0 && relativeX >= centeredTitleXStart && relativeX < centeredTitleXEnd && p.Y == topLeft.Y {
            g.DrawColoredChar(screen, p.X, p.Y, rune(title[relativeX-centeredTitleXStart]), color.White)
        } else {
            g.DrawOnSmallGrid(screen, p.X, p.Y, textureIndex)
        }
    }
    BorderTraversal(topLeft, bottomRight, borderFunc)

    width := bottomRight.X - topLeft.X - 2
    height := bottomRight.Y - topLeft.Y - 2

    backgroundIndex := g.borderDef.BackgroundTextureIndex

    subImage := ExtractSubImageFromAtlas(backgroundIndex, g.smallGridSize, g.smallGridSize, g.getCharAtlas())

    g.op.ColorScale.Reset()
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(g.scale*float64(width), g.scale*float64(height))
    g.op.GeoM.Translate(g.SmallCellToScreen(topLeft.X+1, topLeft.Y+1))
    screen.DrawImage(subImage, g.op)
}

func (g *DualGridRenderer) DrawSmallOnScreen(screen *ebiten.Image, xPos, yPos float64, textureIndex int32) {
    g.op.ColorScale.Reset()
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(g.scale, g.scale)
    g.op.GeoM.Translate(xPos, yPos)
    screen.DrawImage(ExtractSubImageFromAtlas(textureIndex, g.smallGridSize, g.smallGridSize, g.getCharAtlas()), g.op)
}
func (g *DualGridRenderer) DrawEntityOnScreen(screen *ebiten.Image, xPos, yPos float64, textureIndex int32) {
    g.op.ColorScale.Reset()
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(g.scale, g.scale)
    g.op.GeoM.Translate(xPos, yPos)
    screen.DrawImage(ExtractSubImageFromAtlas(textureIndex, g.bigGridSize, g.bigGridSize, g.atlasMap[AtlasEntities]), g.op)
}

func (g *DualGridRenderer) DrawOnBigGrid(screen *ebiten.Image, cellPos, offset geometry.Point, atlasName AtlasName, textureIndex int32) {
    g.op.ColorScale.Reset()
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(g.scale, g.scale)
    g.op.GeoM.Translate(float64(offset.X), float64(offset.Y))
    g.op.GeoM.Translate(g.BigCellToScreen(cellPos.X, cellPos.Y))
    screen.DrawImage(ExtractSubImageFromAtlas(textureIndex, g.bigGridSize, g.bigGridSize, g.atlasMap[atlasName]), g.op)
}

func (g *DualGridRenderer) DrawOnBigGridWithColor(screen *ebiten.Image, cellPos, offset geometry.Point, atlasName AtlasName, textureIndex int32, color color.Color) {
    g.op.ColorScale.Reset()
    g.op.ColorScale.ScaleWithColor(color)
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(g.scale, g.scale)
    g.op.GeoM.Translate(float64(offset.X), float64(offset.Y))
    g.op.GeoM.Translate(g.BigCellToScreen(cellPos.X, cellPos.Y))
    screen.DrawImage(ExtractSubImageFromAtlas(textureIndex, g.bigGridSize, g.bigGridSize, g.atlasMap[atlasName]), g.op)
}
func (g *DualGridRenderer) DrawBigOnScreenWithAtlasNameAndTint(screen *ebiten.Image, xPos, yPos float64, atlasName AtlasName, textureIndex int32, tintColor color.Color) {
    g.op.ColorScale.Reset()
    g.op.ColorScale.ScaleWithColor(tintColor)
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(g.scale, g.scale)
    g.op.GeoM.Translate(xPos, yPos)

    subImageFromAtlas := ExtractSubImageFromAtlas(textureIndex, g.bigGridSize, g.bigGridSize, g.atlasMap[atlasName])
    screen.DrawImage(subImageFromAtlas, g.op)
}
func (g *DualGridRenderer) DrawBigOnScreenWithAtlasAndTint(screen *ebiten.Image, xPos, yPos float64, atlas *ebiten.Image, textureIndex int32, tintColor color.Color) {
    g.op.ColorScale.Reset()
    g.op.ColorScale.ScaleWithColor(tintColor)
    g.op.GeoM.Reset()
    g.op.GeoM.Scale(g.scale, g.scale)
    g.op.GeoM.Translate(xPos, yPos)

    subImageFromAtlas := ExtractSubImageFromAtlas(textureIndex, g.bigGridSize, g.bigGridSize, atlas)
    screen.DrawImage(subImageFromAtlas, g.op)
}

func (g *DualGridRenderer) GetSmallGridScreenSize() geometry.Point {
    return g.smallGridScreenSize
}

func (g *DualGridRenderer) GetScale() float64 {
    return g.scale
}

func (g *DualGridRenderer) GetAutoFitRect(text []string) (geometry.Point, geometry.Point) {
    return g.GetAutoFitRectWithExtraSpace(text, geometry.Point{})
}

func (g *DualGridRenderer) GetAutoFitRectWithExtraSpace(text []string, padding geometry.Point) (geometry.Point, geometry.Point) {
    width := min(util.MaxLen(text)+4+padding.X, 36)
    height := min(len(text)+4+padding.Y, 18)
    screenSize := g.GetSmallGridScreenSize()

    startX := (screenSize.X - width) / 2
    endX := startX + width

    startY := (screenSize.Y - height) / 2
    endY := startY + height

    topLeft := geometry.Point{X: startX, Y: startY}
    bottomRight := geometry.Point{X: endX, Y: endY}
    return topLeft, bottomRight
}

func (g *DualGridRenderer) GetXPosAndHeightForIconText(text []string) (int, int, int) {
    screenSize := g.GetSmallGridScreenSize()

    maxWidth := screenSize.X - 4      // 2*4 for the borders(incl. 1 padding each) + 2 for icon + 1 for padding
    minWidth := util.MaxLen(text) + 7 // 2*2 for the borders(incl. 1 padding each) + 2 for icon + 1 for padding
    width := min(minWidth, maxWidth)

    minHeight := max(len(text)+4, 6)    // 2*2 for the borders(incl. 1 padding each) but we need at least 6 for the icon
    maxHeight := screenSize.Y - 4       // 2*2 for the borders(incl. 1 padding each)
    height := min(minHeight, maxHeight) // max = 18

    startX := (screenSize.X - width) / 2
    endX := startX + width

    return startX, endX, height
}

func (g *DualGridRenderer) AutolayoutArrayToPages(height int, inputText []string) [][]string {
    screenSize := g.GetSmallGridScreenSize()
    width := screenSize.X - 8
    return util.AutoLayoutPages(strings.Join(inputText, " "), width, height)
}

func (g *DualGridRenderer) AutolayoutArrayToIconPages(height int, inputText []string) [][]string {
    screenSize := g.GetSmallGridScreenSize()
    width := screenSize.X - 11
    return util.AutoLayoutPages(strings.Join(inputText, " "), width, height)
}

func (g *DualGridRenderer) SetFontIndexForBigGrid(index map[rune]uint16) {
    g.bigGridFontMap = index
}

func (g *DualGridRenderer) GetBigGridSize() int {
    return g.bigGridSize
}

func ExtractSubImageFromAtlas(textureIndex int32, tileSizeX int, tileSizeY int, textureAtlas *ebiten.Image) *ebiten.Image {
    atlasItemCountX := int32(textureAtlas.Bounds().Size().X / tileSizeX)
    textureRectTopLeft := image.Point{
        X: int((textureIndex % atlasItemCountX) * int32(tileSizeX)),
        Y: int((textureIndex / atlasItemCountX) * int32(tileSizeY)),
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
