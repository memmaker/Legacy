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
    pages          [][]string
    currentPage    int

    shouldClose bool
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

func NewMultiPageWindow(dualGrid *DualGridRenderer, yOffset int, icon int, text []string, onLastPageDisplayed func()) *MultiPageWindow {
    pages := splitIntoPages(text)
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
    m.window.text = m.pages[m.currentPage]
}
func (m *MultiPageWindow) Draw(screen *ebiten.Image) {
    m.window.Draw(screen)
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
    topLeft     geometry.Point
    bottomRight geometry.Point

    iconTextureIndex int
    text             []string

    textColor  color.Color
    iconOffset geometry.Point

    gridRenderer *DualGridRenderer
}

func (i *IconWindow) ActionUp() {

}

func (i *IconWindow) ActionDown() {

}

func NewIconWindow(dualGrid *DualGridRenderer, yOffset int, icon int, text []string) *IconWindow {
    screenSize := dualGrid.GetSmallGridScreenSize()
    topLeft := geometry.Point{X: 3, Y: yOffset}
    bottomRight := geometry.Point{X: screenSize.X - 3, Y: yOffset + 9}
    maxLineLength := bottomRight.X - topLeft.X - 7
    maxLines := bottomRight.Y - topLeft.Y - 4

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
        topLeft:          topLeft,
        bottomRight:      bottomRight,
        iconTextureIndex: icon,
        text:             text,
        iconOffset:       geometry.Point{X: 2, Y: 2},
        gridRenderer:     dualGrid,
        textColor:        color.White,
    }
}

func (i *IconWindow) Draw(screen *ebiten.Image) {

    iconSmallGridX := i.topLeft.X + i.iconOffset.X
    iconSmallGridY := i.topLeft.Y + i.iconOffset.Y

    iconScreenX, iconScreenY := i.gridRenderer.SmallCellToScreen(iconSmallGridX, iconSmallGridY)

    i.gridRenderer.DrawFilledBorder(screen, i.topLeft, i.bottomRight)

    i.gridRenderer.DrawBigOnScreen(screen, iconScreenX, iconScreenY, i.iconTextureIndex)

    for y, line := range i.text {
        i.gridRenderer.DrawColoredString(screen, i.topLeft.X+5, i.topLeft.Y+2+y, line, i.textColor)
    }
}
