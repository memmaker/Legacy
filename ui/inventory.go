package ui

import (
    "Legacy/ega"
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/renderer"
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
)

type InventoryWindow struct {
    renderer.ButtonHolder
    engine       game.Engine
    gridRenderer *renderer.DualGridRenderer
    currentPage  InventoryPage
    items        [][]game.Item
    topLeft      geometry.Point
    bottomRight  geometry.Point

    downIndicator int32
    upIndicator   int32

    currentSelection int
    scrollOffset     int
    filterMap        map[InventoryPage]func(item game.Item) bool
}

func (i *InventoryWindow) canScrollUp() bool {
    if !i.needsScroll() {
        return false
    }
    return i.scrollOffset > 0
}
func (i *InventoryWindow) canScrollDown() bool {
    if !i.needsScroll() {
        return false
    }
    return i.scrollOffset < i.neededHeightForContent()-i.availableHeightForContent()
}
func (i *InventoryWindow) needsScroll() bool {
    return i.neededHeightForContent() > i.availableHeightForContent()
}
func (i *InventoryWindow) availableHeightForContent() int {
    return i.bottomRight.Y - i.topLeft.Y - 4
}
func (i *InventoryWindow) neededHeightForContent() int {
    return len(i.items)
}

func (i *InventoryWindow) Draw(screen *ebiten.Image) {
    i.gridRenderer.DrawFilledBorder(screen, i.topLeft, i.bottomRight, "")
    i.ButtonHolder.Draw(i.gridRenderer, screen)
    i.drawItems(screen)
    i.drawScrollIndicators(screen)
}

func (i *InventoryWindow) ActionUp() {
    i.currentSelection--
    if i.currentSelection < 0 {
        i.currentSelection = len(i.items) - 1
        i.scrollOffset = i.neededHeightForContent() - i.availableHeightForContent()
    } else if i.currentSelection < i.scrollOffset {
        i.scrollOffset--
    }
}

func (i *InventoryWindow) ActionDown() {
    i.currentSelection++
    if i.currentSelection >= len(i.items) {
        i.currentSelection = 0
        i.scrollOffset = 0
    } else if i.currentSelection >= i.scrollOffset+i.availableHeightForContent() {
        i.scrollOffset++
    }
}

func (i *InventoryWindow) ActionConfirm() {
    if !i.isValidItemSelected() {
        return
    }
    item := i.items[i.currentSelection]
    i.activateItemContextMenu(item[0])
}

func (i *InventoryWindow) ShouldClose() bool {
    return false
}

func (i *InventoryWindow) OnAvatarSwitched() {

}

func (i *InventoryWindow) ActionLeft() {

}

func (i *InventoryWindow) ActionRight() {

}

func (i *InventoryWindow) OnMouseMoved(x int, y int) renderer.Tooltip {
    if y == i.topLeft.Y {
        i.ButtonHolder.OnMouseMoved(x, y)
        return renderer.NoTooltip{}
    }

    if y < i.topLeft.Y+2 || y > i.bottomRight.Y-2 {
        return renderer.NoTooltip{}
    }

    if x < i.topLeft.X+2 || x > i.bottomRight.X-3 {
        return renderer.NoTooltip{}
    }

    relativeOffset := y - i.topLeft.Y - 2 + i.scrollOffset

    if relativeOffset >= len(i.items) || relativeOffset < 0 {
        return renderer.NoTooltip{}
    }

    i.currentSelection = relativeOffset

    item := i.items[i.currentSelection][0]
    return NewItemTooltip(i.gridRenderer, item, geometry.Point{X: x, Y: y})
}

func (i *InventoryWindow) OnMouseClicked(x int, y int) bool {
    if i.ButtonHolder.OnMouseClicked(x, y) {
        return true
    }
    if y < i.topLeft.Y+2 || y > i.bottomRight.Y-2 {
        return false
    }

    if x < i.topLeft.X+2 || x > i.bottomRight.X-3 {
        return false
    }

    i.OnMouseMoved(x, y)
    if i.isValidItemSelected() {
        i.ActionConfirm()
        return true
    }
    return false
}

func (i *InventoryWindow) activateItemContextMenu(item game.Item) {
    if item == nil {
        return
    }
    itemActions := item.GetContextActions(i.engine)
    for index, action := range itemActions {
        oldAction := action.Action
        newAction := func() {
            oldAction()
            i.engine.OpenPartyInventoryOnPage(int(i.currentPage))
        }
        action.Action = newAction
        itemActions[index] = action
    }
    i.engine.OpenMenu(itemActions)
}

func (i *InventoryWindow) createButtons() {
    buttonPositionStart := geometry.Point{X: i.topLeft.X + 1, Y: i.topLeft.Y}
    buttonWidth := 4
    // +1 padding between buttons

    allButtonRect := geometry.NewRect(buttonPositionStart.X, buttonPositionStart.Y, buttonPositionStart.X+buttonWidth, buttonPositionStart.Y+1)
    allButton := i.ButtonHolder.AddIconAndTextButton(allButtonRect, func() {
        i.SwitchPageTo(InventoryPageAll)
    })
    allButton.SetText("All")
    allButton.SetIcon(8)
    buttonPositionStart.X += buttonWidth + 1

    weaponButtonRect := geometry.NewRect(buttonPositionStart.X, buttonPositionStart.Y, buttonPositionStart.X+buttonWidth, buttonPositionStart.Y+1)
    weaponButton := i.ButtonHolder.AddIconAndTextButton(weaponButtonRect, func() {
        i.SwitchPageTo(InventoryPageWeapons)
    })
    weaponButton.SetText("Wep")
    weaponButton.SetIcon(164)
    buttonPositionStart.X += buttonWidth + 1

    armorButtonRect := geometry.NewRect(buttonPositionStart.X, buttonPositionStart.Y, buttonPositionStart.X+buttonWidth, buttonPositionStart.Y+1)
    armorButton := i.ButtonHolder.AddIconAndTextButton(armorButtonRect, func() {
        i.SwitchPageTo(InventoryPageArmor)
    })

    armorButton.SetText("Arm")
    armorButton.SetIcon(161)
    buttonPositionStart.X += buttonWidth + 1

    accessoryButtonRect := geometry.NewRect(buttonPositionStart.X, buttonPositionStart.Y, buttonPositionStart.X+buttonWidth, buttonPositionStart.Y+1)
    accessoryButton := i.ButtonHolder.AddIconAndTextButton(accessoryButtonRect, func() {
        i.SwitchPageTo(InventoryPageAccessories)
    })
    accessoryButton.SetText("Acc")
    accessoryButton.SetIcon(168)
    buttonPositionStart.X += buttonWidth + 1

    consumableButtonRect := geometry.NewRect(buttonPositionStart.X, buttonPositionStart.Y, buttonPositionStart.X+buttonWidth, buttonPositionStart.Y+1)
    consumableButton := i.ButtonHolder.AddIconAndTextButton(consumableButtonRect, func() {
        i.SwitchPageTo(InventoryPageConsumables)
    })
    consumableButton.SetText("Con")
    consumableButton.SetIcon(173)
    buttonPositionStart.X += buttonWidth + 1

    scrollButtonRect := geometry.NewRect(buttonPositionStart.X, buttonPositionStart.Y, buttonPositionStart.X+buttonWidth, buttonPositionStart.Y+1)
    scrollButton := i.ButtonHolder.AddIconAndTextButton(scrollButtonRect, func() {
        i.SwitchPageTo(InventoryPageScrolls)
    })
    scrollButton.SetText("Scr")
    scrollButton.SetIcon(175)
    buttonPositionStart.X += buttonWidth + 1
    /*
       keyButtonRect := geometry.NewRect(buttonPositionStart.X, buttonPositionStart.Y, buttonPositionStart.X+buttonWidth, buttonPositionStart.Y+1)
       keyButton := i.ButtonHolder.AddIconAndTextButton(keyButtonRect, func() {
           i.showListOfKeys()
       })
       keyButton.SetText("Key")
       keyButton.SetIcon(172)
       buttonPositionStart.X += buttonWidth + 1
    */
    miscButtonRect := geometry.NewRect(buttonPositionStart.X, buttonPositionStart.Y, buttonPositionStart.X+buttonWidth, buttonPositionStart.Y+1)
    miscButton := i.ButtonHolder.AddIconAndTextButton(miscButtonRect, func() {
        i.SwitchPageTo(InventoryPageMisc)
    })
    miscButton.SetText("Msc")
    miscButton.SetIcon(169)

    // start with all selected.
    i.ButtonHolder.SetSelectedButton(allButton)
}

func (i *InventoryWindow) SwitchPageTo(page InventoryPage) {
    i.currentPage = page
    i.updateItemList()
}

func (i *InventoryWindow) getCurrentFilter() func(item game.Item) bool {
    return i.filterMap[i.currentPage]
}
func (i *InventoryWindow) updateItemList() {
    party := i.engine.GetParty()
    i.items = party.GetFilteredStackedInventory(i.getCurrentFilter())
}

func (i *InventoryWindow) drawItems(screen *ebiten.Image) {
    for y := i.topLeft.Y + 2; y < i.bottomRight.Y-2; y++ {
        relativeItemOffset := y - i.topLeft.Y - 2 + i.scrollOffset
        if relativeItemOffset >= len(i.items) {
            break
        }
        itemStack := i.items[relativeItemOffset]
        firstItemInStack := itemStack[0]
        i.gridRenderer.DrawOnSmallGrid(screen, i.topLeft.X+2, y, firstItemInStack.InventoryIcon())
        drawColor := firstItemInStack.TintColor()
        if i.currentSelection == relativeItemOffset {
            drawColor = ega.BrightGreen
        }
        stackLabel := fmt.Sprintf("%s (x%d)", firstItemInStack.Name(), len(itemStack))
        if len(itemStack) == 1 {
            stackLabel = firstItemInStack.Name()
        }
        if wornItem, ok := firstItemInStack.(game.Equippable); ok && wornItem.IsEquipped() {
            wearer := wornItem.GetWearer()
            wearerIcon := i.engine.GetParty().GetMemberIcon(wearer)
            stackLabel = fmt.Sprintf("%s (%s)", stackLabel, string(wearerIcon))
        }
        i.gridRenderer.DrawColoredString(screen, i.topLeft.X+3, y, stackLabel, drawColor)
    }
}

func (i *InventoryWindow) drawScrollIndicators(screen *ebiten.Image) {
    if i.needsScroll() {
        if i.canScrollUp() {
            i.gridRenderer.DrawOnSmallGrid(screen, i.bottomRight.X-1, i.topLeft.Y+2, i.upIndicator)
        }
        if i.canScrollDown() {
            i.gridRenderer.DrawOnSmallGrid(screen, i.bottomRight.X-1, i.bottomRight.Y-3, i.downIndicator)
        }
    }
}

func (i *InventoryWindow) isValidItemSelected() bool {
    return i.currentSelection < len(i.items)
}

type InventoryPage int

const (
    InventoryPageAll InventoryPage = iota
    InventoryPageWeapons
    InventoryPageArmor
    InventoryPageAccessories
    InventoryPageConsumables
    InventoryPageScrolls
    InventoryPageMisc
)

func (i *InventoryWindow) loadFilter() {
    filtermap := make(map[InventoryPage]func(item game.Item) bool)
    filtermap[InventoryPageAll] = func(item game.Item) bool {
        return true
    }

    filtermap[InventoryPageWeapons] = func(item game.Item) bool {
        if _, ok := item.(*game.Weapon); ok {
            return true
        }
        return false
    }

    filtermap[InventoryPageArmor] = func(item game.Item) bool {
        if armorItem, ok := item.(*game.Armor); ok {
            return !armorItem.IsAccessory()
        }
        return false
    }

    filtermap[InventoryPageAccessories] = func(item game.Item) bool {
        if armorItem, ok := item.(*game.Armor); ok {
            return armorItem.IsAccessory()
        }
        return false
    }

    filtermap[InventoryPageConsumables] = func(item game.Item) bool {
        if _, ok := item.(*game.Potion); ok {
            return true
        }
        return false
    }

    filtermap[InventoryPageScrolls] = func(item game.Item) bool {
        if _, ok := item.(*game.Scroll); ok {
            return true
        }
        return false
    }

    filtermap[InventoryPageMisc] = func(item game.Item) bool {
        // not armor, not weapon, not scroll, not consumable, not key
        if _, ok := item.(*game.Armor); ok {
            return false
        }
        if _, ok := item.(*game.Weapon); ok {
            return false
        }
        if _, ok := item.(*game.Scroll); ok {
            return false
        }
        if _, ok := item.(*game.Potion); ok {
            return false
        }
        if _, ok := item.(*game.Key); ok {
            return false
        }
        return true
    }
    i.filterMap = filtermap
}

func NewInventoryWindow(engine game.Engine, gridRenderer *renderer.DualGridRenderer) *InventoryWindow {
    smallScreenSize := gridRenderer.GetSmallGridScreenSize()
    i := &InventoryWindow{
        ButtonHolder:  renderer.NewButtonHolder(),
        engine:        engine,
        gridRenderer:  gridRenderer,
        topLeft:       geometry.Point{X: 2, Y: 2},
        bottomRight:   geometry.Point{X: smallScreenSize.X - 2, Y: smallScreenSize.Y - 3},
        downIndicator: 5,
        upIndicator:   6,
    }
    i.loadFilter()
    i.createButtons()
    i.SwitchPageTo(InventoryPageAll)
    return i
}
