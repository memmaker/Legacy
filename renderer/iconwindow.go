package renderer

import (
    "Legacy/geometry"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
)

type MultiPageWindow struct {
    window         *IconWindow
    onLastPage     func()
    lastPageCalled bool
    pages          []string
    currentPage    int

    shouldClose bool
}

func (m *MultiPageWindow) OnMouseClicked(x int, y int) bool {
    if m.window.OnMouseClicked(x, y) {
        return true
    }
    if x < m.window.topLeft.X || x >= m.window.bottomRight.X {
        return false
    }
    if y < m.window.topLeft.Y || y >= m.window.bottomRight.Y {
        return false
    }
    m.ActionConfirm()
    return true
}

func (m *MultiPageWindow) ShouldClose() bool {
    return m.shouldClose
}

func (m *MultiPageWindow) ActionUp() {
    m.currentPage--
    if m.currentPage < 0 {
        m.currentPage = 0
    }
}

func (m *MultiPageWindow) ActionDown() {
    m.ActionConfirm()
}

func NewMultiPageWindow(dualGrid *DualGridRenderer, yOffset int, icon int32, pages []string, onLastPageDisplayed func()) *MultiPageWindow {
    lastPageCalled := false
    if len(pages) < 2 {
        onLastPageDisplayed()
        lastPageCalled = true
    }
    return &MultiPageWindow{
        window:         NewIconWindow(dualGrid, yOffset, icon, pages[0]),
        onLastPage:     onLastPageDisplayed,
        pages:          pages,
        lastPageCalled: lastPageCalled,
    }
}
func (m *MultiPageWindow) ActionConfirm() {
    m.currentPage++
    if m.currentPage >= len(m.pages)-1 {
        m.currentPage = len(m.pages) - 1
        if !m.lastPageCalled {
            m.onLastPage()
            m.lastPageCalled = true
        } else {
            m.shouldClose = true
        }
    }
    m.window.SetText(m.pages[m.currentPage])
}
func (m *MultiPageWindow) Draw(screen *ebiten.Image) {
    m.window.Draw(screen)
}

func (m *MultiPageWindow) SetTitle(name string) {
    m.window.title = name
}

func (m *MultiPageWindow) AddTextActionButton(icon int32, callback func(text []string)) {
    m.window.AddTextActionButton(icon, callback)
}

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
}

func (i *IconWindow) ActionUp() {

}

func (i *IconWindow) ActionDown() {

}

func NewIconWindow(dualGrid *DualGridRenderer, yOffset int, icon int32, inputText string) *IconWindow {
    screenSize := dualGrid.GetSmallGridScreenSize()
    topLeft := geometry.Point{X: 3, Y: yOffset}
    bottomRight := geometry.Point{X: screenSize.X - 3, Y: yOffset + 9}
    maxLineLength := bottomRight.X - topLeft.X - 7
    maxLines := bottomRight.Y - topLeft.Y - 4
    text := AutoLayout(inputText, maxLineLength)

    for i, line := range text {
        if i >= maxLines {
            println("WARNING: too many lines in icon window\n" + line)
            break
        }
        if len(line) > maxLineLength {
            text[i] = line[:maxLineLength]
            println("WARNING: line too long in icon window\n" + line)
        }
    }
    return &IconWindow{
        ButtonHolder:     NewButtonHolder(),
        topLeft:          topLeft,
        bottomRight:      bottomRight,
        iconTextureIndex: icon,
        text:             text,
        iconOffset:       geometry.Point{X: 2, Y: 2},
        gridRenderer:     dualGrid,
        textColor:        color.White,
    }
}

func (i *IconWindow) AddTextActionButton(icon int32, action func(currentText []string)) {
    startPosX := i.bottomRight.X - 3
    yPos := i.topLeft.Y
    iconPos := geometry.Point{X: startPosX, Y: yPos}
    i.AddButton(iconPos, icon, func() {
        action(i.text)
    })
}

func (i *IconWindow) Draw(screen *ebiten.Image) {

    iconSmallGridX := i.topLeft.X + i.iconOffset.X
    iconSmallGridY := i.topLeft.Y + i.iconOffset.Y

    iconScreenX, iconScreenY := i.gridRenderer.SmallCellToScreen(iconSmallGridX, iconSmallGridY)

    i.gridRenderer.DrawFilledBorder(screen, i.topLeft, i.bottomRight, i.title)

    i.ButtonHolder.Draw(i.gridRenderer, screen)

    i.gridRenderer.DrawEntityOnScreen(screen, iconScreenX, iconScreenY, i.iconTextureIndex)

    for y, line := range i.text {
        i.gridRenderer.DrawColoredString(screen, i.topLeft.X+5, i.topLeft.Y+2+y, line, i.textColor)
    }
}

func (i *IconWindow) OnMouseClicked(x int, y int) bool {
    return i.ButtonHolder.OnMouseClicked(x, y)
}

func (i *IconWindow) SetText(inputText string) {
    i.text = AutoLayout(inputText, i.maxLineLength())
}

func (i *IconWindow) maxLineLength() int {
    return i.bottomRight.X - i.topLeft.X - 7
}
