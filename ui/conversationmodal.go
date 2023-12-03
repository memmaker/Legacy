package ui

import (
    "Legacy/geometry"
    "Legacy/renderer"
    "Legacy/util"
    "github.com/hajimehoshi/ebiten/v2"
)

type Menu interface {
    OnCommand(command CommandType) bool
    Draw(screen *ebiten.Image)
    OnMouseMoved(x int, y int) (bool, Tooltip)
    OnMouseClicked(x int, y int) bool
    ShouldClose() bool
    OnMouseWheel(x int, y int, dy float64) bool
}

type ConversationModal struct {
    gridRenderer *renderer.DualGridRenderer

    textWindow      *MultiPageWindow
    responseInput   Menu
    onCloseFunc     func()
    reachedLastPage bool
    shouldClose     bool
    closeOnCancel   bool
}

func (c *ConversationModal) OnMouseWheel(x int, y int, dy float64) bool {
    if c.responseInput != nil {
        return c.responseInput.OnMouseWheel(x, y, dy)
    }
    return false
}

func (c *ConversationModal) OnCommand(command CommandType) bool {
    if c.isShowingResponseInput() {
        return c.responseInput.OnCommand(command)
    }
    if c.textWindow != nil {
        return c.textWindow.OnCommand(command)
    }
    return false
}

func NewConversationModal(gridRenderer *renderer.DualGridRenderer, journalFunc func(text []string)) *ConversationModal {
    c := &ConversationModal{
        gridRenderer:  gridRenderer,
        closeOnCancel: true,
    }
    c.openTextWindow(journalFunc)
    return c
}

func (c *ConversationModal) Draw(screen *ebiten.Image) {
    if c.textWindow != nil {
        c.textWindow.Draw(screen)
    }
    if c.isShowingResponseInput() {
        c.responseInput.Draw(screen)
    }
}

func (c *ConversationModal) isShowingResponseInput() bool {
    return c.responseInput != nil && !c.responseInput.ShouldClose() && c.reachedLastPage
}

func (c *ConversationModal) openTextWindow(addToJournal func(text []string)) {
    multipageWindow := NewMultiPageWindow(c.gridRenderer)
    multipageWindow.InitWithoutText()
    multipageWindow.AddTextActionButton(134, addToJournal)
    multipageWindow.SetAutoCloseOnConfirm()
    multipageWindow.SetOnLastPage(c.onLastPageOfText)
    multipageWindow.SetOnClose(c.onTextWindowWantsToClose)
    c.textWindow = multipageWindow
}

func (c *ConversationModal) onLastPageOfText() {
    c.reachedLastPage = true
}

func (c *ConversationModal) SetIcon(icon int32) {
    c.textWindow.SetIcon(icon)
}
func (c *ConversationModal) SetTitle(title string) {
    c.textWindow.SetTitle(title)
}

func (c *ConversationModal) SetText(text [][]string) {
    if text == nil || len(text) == 0 {
        c.textWindow = nil
        return
    }
    c.shouldClose = false
    if c.textWindow == nil {
        c.openTextWindow(nil)
    }
    c.reachedLastPage = false
    c.textWindow.SetFixedText(text)
    c.textWindow.PositionAtY(2)
}

func (c *ConversationModal) SetOptions(items []util.MenuItem) {
    if items == nil || len(items) == 0 {
        c.responseInput = nil
        return
    }
    c.shouldClose = false
    c.responseInput = NewGridDialogueMenu(c.gridRenderer, geometry.Point{X: 2, Y: 12}, items)
}
func (c *ConversationModal) SetVendorOptions(items []util.MenuItem) {
    c.closeOnCancel = true
    if items == nil || len(items) == 0 {
        c.responseInput = nil
        return
    }
    c.shouldClose = false
    vendorGridMenu := NewGridMenuAtY(c.gridRenderer, items, 10)
    c.responseInput = vendorGridMenu
}
func (c *ConversationModal) SetOnClose(onClassCallback func()) {
    c.onCloseFunc = onClassCallback
}

func (c *ConversationModal) onTextWindowWantsToClose() {
    if c.isShowingResponseInput() {
        return
    }
    if c.closeOnCancel {
        c.shouldClose = true
    }
    c.onClose()
}
func (c *ConversationModal) onClose() {
    if c.onCloseFunc != nil {
        c.onCloseFunc()
        c.onCloseFunc = nil
    }
}

func (c *ConversationModal) ShouldClose() bool {
    return c.shouldClose
}

func (c *ConversationModal) OnMouseMoved(x int, y int) (bool, Tooltip) {
    if c.isShowingResponseInput() {
        return c.responseInput.OnMouseMoved(x, y)

    }
    return false, NoTooltip{}
}

func (c *ConversationModal) OnMouseClicked(x int, y int) bool {
    if !c.isShowingResponseInput() {
        if c.textWindow != nil {
            textWindowClicked := c.textWindow.OnMouseClicked(x, y)
            if !textWindowClicked {
                c.textWindow.nextPage()
                return true
            }
            return textWindowClicked
        }
        return false
    }
    if c.responseInput.OnMouseClicked(x, y) {
        return true
    }
    if c.textWindow != nil {
        return c.textWindow.OnMouseClicked(x, y)
    }
    return false
}

func (c *ConversationModal) ActionCancel() {
    if c.closeOnCancel {
        c.shouldClose = true
    }
}
