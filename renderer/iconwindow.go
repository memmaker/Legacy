package renderer

import (
    "Legacy/geometry"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
    "slices"
)

func splitIntoPages(text []string) [][]string {
    result := make([][]string, 0)
    currentPage := make([]string, 0)
    for _, line := range text {
        if line == "" {
            result = append(result, currentPage)
            currentPage = make([]string, 0)
            continue
        }
        currentPage = append(currentPage, line)
    }
    if len(currentPage) > 0 {
        result = append(result, currentPage)
    }
    return result
}

type IconWindow struct {
    ButtonHolder
    topLeft     geometry.Point
    bottomRight geometry.Point

    iconTextureIndex int32
    text             []string

    textColor  color.Color
    iconOffset geometry.Point

    gridRenderer *DualGridRenderer
    title        string
    shouldClose  bool
    onClose      func()
    currentIndex int
    allowedIcons []int32
}

func (i *IconWindow) OnAvatarSwitched() {

}

func (i *IconWindow) ActionConfirm() {
    i.shouldClose = true
    if i.onClose != nil {
        i.onClose()
    }
}

func (i *IconWindow) ShouldClose() bool {
    return i.shouldClose
}

func (i *IconWindow) ActionUp() {
    i.currentIndex++
    if i.currentIndex >= len(i.allowedIcons) {
        i.currentIndex = len(i.allowedIcons) - 1
    }
}

func (i *IconWindow) ActionDown() {
    i.currentIndex--
    if i.currentIndex < 0 {
        i.currentIndex = 0
    }
}

func (i *IconWindow) Icon() int32 {
    if len(i.allowedIcons) == 0 {
        return i.iconTextureIndex
    }
    return i.allowedIcons[i.currentIndex]
}

func NewIconWindow(dualGrid *DualGridRenderer) *IconWindow {
    return &IconWindow{
        ButtonHolder: NewButtonHolder(),
        iconOffset:   geometry.Point{X: 2, Y: 2},
        gridRenderer: dualGrid,
        textColor:    color.White,
    }
}
func (i *IconWindow) SetAutoLayoutText(inputText string) {
    screenSize := i.gridRenderer.GetSmallGridScreenSize()
    maxLineLength := screenSize.X - 11
    text := AutoLayout(inputText, maxLineLength)
    i.SetFixedText(text)
}

// SetYOffset sets the position of the window on the screen.
// IMPORTANT: It assumes that a text has been set!
func (i *IconWindow) SetYOffset(yOffset int) {
    startX, endX, height := i.gridRenderer.GetXPosAndHeightForIconText(i.text)
    topLeft := geometry.Point{X: startX, Y: yOffset}
    bottomRight := geometry.Point{X: endX, Y: yOffset + height}
    i.topLeft = topLeft
    i.bottomRight = bottomRight
}

func (i *IconWindow) RePosition() {
    i.SetYOffset(i.topLeft.Y)
}

func (i *IconWindow) SetFixedText(text []string) {
    i.text = text
}
func (i *IconWindow) AddTextActionButton(icon int32, action func(currentText []string)) {
    startPosX := i.bottomRight.X - 3
    yPos := i.topLeft.Y
    iconPos := geometry.Point{X: startPosX, Y: yPos}
    i.AddIconButton(iconPos, icon, func() {
        action(i.text)
    })
}

func (i *IconWindow) Draw(screen *ebiten.Image) {

    iconSmallGridX := i.topLeft.X + i.iconOffset.X
    iconSmallGridY := i.topLeft.Y + i.iconOffset.Y

    iconScreenX, iconScreenY := i.gridRenderer.SmallCellToScreen(iconSmallGridX, iconSmallGridY)

    i.gridRenderer.DrawFilledBorder(screen, i.topLeft, i.bottomRight, i.title)

    i.ButtonHolder.Draw(i.gridRenderer, screen)

    i.gridRenderer.DrawEntityOnScreen(screen, iconScreenX, iconScreenY, i.Icon())

    for y, line := range i.text {
        i.gridRenderer.DrawColoredString(screen, i.topLeft.X+5, i.topLeft.Y+2+y, line, i.textColor)
    }
}

func (i *IconWindow) OnMouseClicked(x int, y int) bool {
    return i.ButtonHolder.OnMouseClicked(x, y)
}

func (i *IconWindow) maxLineLength() int {
    return i.bottomRight.X - i.topLeft.X - 7
}

func (i *IconWindow) SetOnClose(callback func()) {
    i.onClose = callback
}

func (i *IconWindow) SetAllowedIcons(icons []int32) {
    i.allowedIcons = icons
    i.currentIndex = 0
}

func (i *IconWindow) SetCurrentIcon(icon int32) {
    if len(i.allowedIcons) == 0 {
        i.iconTextureIndex = icon
        return
    }
    indexOf := slices.Index(i.allowedIcons, icon)
    if indexOf == -1 {
        return
    }
    i.currentIndex = indexOf
}
