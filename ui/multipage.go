package ui

import (
    "Legacy/renderer"
    "github.com/hajimehoshi/ebiten/v2"
)

type MultiPageWindow struct {
    window         *IconWindow
    onLastPage     func()
    lastPageCalled bool
    pages          [][]string
    currentPage    int
    closeOnConfirm bool

    shouldClose    bool
    gridRenderer   *renderer.DualGridRenderer
    onClose        func()
    cannotBeClosed bool
}

func (m *MultiPageWindow) OnMouseWheel(x int, y int, dy float64) bool {
    return false
}

func (m *MultiPageWindow) OnCommand(command CommandType) bool {
    switch command {
    case PlayerCommandCancel:
        m.ActionCancel()
    case PlayerCommandConfirm:
        m.ActionConfirm()
    case PlayerCommandUp:
        m.ActionUp()
    case PlayerCommandDown:
        m.ActionDown()
    case PlayerCommandLeft:
        m.ActionLeft()
    case PlayerCommandRight:
        m.ActionRight()
    }
    return true
}

func (m *MultiPageWindow) ActionLeft() {

}

func (m *MultiPageWindow) ActionRight() {

}

func (m *MultiPageWindow) OnMouseMoved(x int, y int) (bool, Tooltip) {
    return false, NoTooltip{}
}

func (m *MultiPageWindow) CanBeClosed() bool {
    return !m.cannotBeClosed
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

func NewMultiPageWindow(dualGrid *renderer.DualGridRenderer) *MultiPageWindow {
    return &MultiPageWindow{
        gridRenderer:   dualGrid,
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

func (m *MultiPageWindow) InitWithoutText() {
    m.currentPage = 0
    m.window = NewIconWindow(m.gridRenderer)
}

func (m *MultiPageWindow) PositionAtY(offset int) {
    m.window.SetYOffset(offset)
}
func (m *MultiPageWindow) ActionConfirm() {
    m.nextPage()
}

func (m *MultiPageWindow) ActionCancel() {
    m.nextPage()
}

func (m *MultiPageWindow) nextPage() {
    m.currentPage++
    if m.currentPage >= len(m.pages)-1 {
        m.currentPage = len(m.pages) - 1
        if !m.lastPageCalled {
            m.lastPage()
        } else if m.closeOnConfirm && !m.cannotBeClosed {
            m.shouldClose = true
            if m.onClose != nil {
                m.onClose()
            }
        }
    }
    m.window.SetFixedText(m.pages[m.currentPage])
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

func (m *MultiPageWindow) SetOnClose(onCloseCallback func()) {
    m.onClose = onCloseCallback
}

func (m *MultiPageWindow) SetCannotBeClosed() {
    m.cannotBeClosed = true
}

func (m *MultiPageWindow) SetOnLastPage(onLastPageFunc func()) {
    m.onLastPage = onLastPageFunc
}

func (m *MultiPageWindow) SetFixedText(text [][]string) {
    m.pages = text
    m.currentPage = 0
    m.lastPageCalled = false
    m.window.SetFixedText(m.pages[m.currentPage])
    if len(text) == 1 {
        m.lastPage()
    }
}
