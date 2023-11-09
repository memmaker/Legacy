package renderer

import "github.com/hajimehoshi/ebiten/v2"

type MultiPageWindow struct {
    window         *IconWindow
    onLastPage     func()
    lastPageCalled bool
    pages          [][]string
    currentPage    int
    closeOnConfirm bool

    shouldClose  bool
    gridRenderer *DualGridRenderer
}

func (m *MultiPageWindow) OnAvatarSwitched() {

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

func (m *MultiPageWindow) SetAutoCloseOnConfirm() {
    m.closeOnConfirm = true
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

func NewMultiPageWindow(dualGrid *DualGridRenderer, onLastPageDisplayed func()) *MultiPageWindow {
    return &MultiPageWindow{
        gridRenderer:   dualGrid,
        onLastPage:     onLastPageDisplayed,
        lastPageCalled: false,
    }
}

func (m *MultiPageWindow) InitWithFixedText(pages [][]string) {
    m.pages = pages
    m.currentPage = 0
    m.window = NewIconWindow(m.gridRenderer)
    m.window.SetFixedText(m.pages[m.currentPage])

    if len(pages) == 1 {
        m.lastPage()
    }
}

func (m *MultiPageWindow) PositionAtY(offset int) {
    m.window.SetYOffset(offset)
}
func (m *MultiPageWindow) ActionConfirm() {
    m.currentPage++
    if m.currentPage >= len(m.pages)-1 {
        m.currentPage = len(m.pages) - 1
        if !m.lastPageCalled {
            m.lastPage()
        } else if m.closeOnConfirm {
            m.shouldClose = true
        }
    }
    m.window.SetFixedText(m.pages[m.currentPage])
    m.window.RePosition()
}

func (m *MultiPageWindow) lastPage() {
    m.lastPageCalled = true
    if m.onLastPage == nil {
        return
    }
    m.onLastPage()
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

func (m *MultiPageWindow) SetIcon(icon int32) {
    m.window.SetCurrentIcon(icon)
}
