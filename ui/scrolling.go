package ui

import (
    "Legacy/geometry"
    "Legacy/renderer"
    "github.com/hajimehoshi/ebiten/v2"
)

type ScrollingContent struct {
    downIndicator          int32
    upIndicator            int32
    scrollOffset           float64
    neededHeightForContent func() int
    availableSpace         func() geometry.Rect
}

func NewScrollingContentWithFunctions(neededHeight func() int, availableSpace func() geometry.Rect) ScrollingContent {
    return ScrollingContent{
        downIndicator:          5,
        upIndicator:            6,
        neededHeightForContent: neededHeight,
        availableSpace:         availableSpace,
    }
}

func (s *ScrollingContent) ScrollDown(amount int) {
    if !s.canScrollDown() {
        return
    }
    s.scrollOffset += float64(amount)
    s.keepOffsetInRange()
}

func (s *ScrollingContent) ScrollUp(amount int) {
    if !s.canScrollUp() {
        return
    }
    s.scrollOffset -= float64(amount)
    s.keepOffsetInRange()
}

func (s *ScrollingContent) availableHeightForContent() int {
    size := s.availableSpace().Size()
    return size.Y
}

func (s *ScrollingContent) gridScrollOffset() int {
    return int(s.scrollOffset)
}
func (s *ScrollingContent) getLineFromScreenLine(y int) int {
    topLeft := s.availableSpace().Min
    return y - topLeft.Y + s.gridScrollOffset()
}
func (s *ScrollingContent) maxScrollOffset() float64 {
    return float64(s.neededHeightForContent() - s.availableHeightForContent())
}
func (s *ScrollingContent) OnMouseWheel(x int, y int, dy float64) bool {
    if !s.needsScroll() {
        return false
    }
    sensitity := 0.2
    s.scrollOffset += -dy * sensitity
    s.keepOffsetInRange()
    return true
}

func (s *ScrollingContent) keepOffsetInRange() {
    if s.scrollOffset < 0 {
        s.ScrollToTop()
    } else if s.scrollOffset > s.maxScrollOffset() {
        s.ScrollToBottom()
    }
}

func (s *ScrollingContent) canScrollUp() bool {
    if !s.needsScroll() {
        return false
    }
    return s.scrollOffset > 0
}
func (s *ScrollingContent) canScrollDown() bool {
    if !s.needsScroll() {
        return false
    }
    return s.gridScrollOffset() < s.neededHeightForContent()-s.availableHeightForContent()
}
func (s *ScrollingContent) needsScroll() bool {
    neededHeightForContent := s.neededHeightForContent()
    availableHeightForContent := s.availableHeightForContent()
    return neededHeightForContent > availableHeightForContent
}

func (s *ScrollingContent) ScrollToTop() {
    s.scrollOffset = 0
}

func (s *ScrollingContent) ScrollToBottom() {
    if !s.needsScroll() {
        return
    }
    s.scrollOffset = s.maxScrollOffset()
}

func (s *ScrollingContent) MakeIndexVisible(selectionIndex int) {
    if !s.needsScroll() {
        return
    }
    if selectionIndex < s.gridScrollOffset() {
        s.scrollOffset = float64(selectionIndex)
    } else if selectionIndex >= s.gridScrollOffset()+s.availableHeightForContent() {
        s.scrollOffset = float64(selectionIndex - s.availableHeightForContent() + 1)
    }
}

func (s *ScrollingContent) drawScrollIndicators(gridRenderer *renderer.DualGridRenderer, screen *ebiten.Image) {
    if s.needsScroll() {
        xPos := s.availableSpace().Max.X
        if s.canScrollUp() {
            yPos := s.availableSpace().Min.Y
            gridRenderer.DrawOnSmallGrid(screen, xPos, yPos, s.upIndicator)
        }
        if s.canScrollDown() {
            yPos := s.availableSpace().Max.Y - 1
            gridRenderer.DrawOnSmallGrid(screen, xPos, yPos, s.downIndicator)
        }
    }
}
