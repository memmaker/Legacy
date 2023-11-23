package ui

import (
    "Legacy/ega"
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/renderer"
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "slices"
    "sort"
)

type InventoryWindow struct {
    ButtonHolder
    ScrollingContent
    engine       game.Engine
    gridRenderer *renderer.DualGridRenderer
    currentPage  InventoryPage
    items        [][]game.Item
    topLeft      geometry.Point
    bottomRight  geometry.Point

    currentSelection int

    filterMap   map[InventoryPage]func(item game.Item) bool
    sortMap     map[InventoryPage]func(one game.Item, two game.Item) bool
    shouldClose bool
}

func (i *InventoryWindow) OnMouseWheel(x int, y int, dy float64) bool {
    if i.ScrollingContent.OnMouseWheel(x, y, dy) {
        i.OnMouseMoved(x, y)
    }
    return true
}

func (i *InventoryWindow) OnCommand(command CommandType) bool {
    switch command {
    case PlayerCommandUp:
        i.ActionUp()
    case PlayerCommandDown:
        i.ActionDown()
    case PlayerCommandLeft:
        i.ActionLeft()
    case PlayerCommandRight:
        i.ActionRight()
    case PlayerCommandConfirm:
        i.ActionConfirm()
    case PlayerCommandCancel:
        i.ActionCancel()
    }
    return true
}

func (i *InventoryWindow) CanBeClosed() bool {
    return true
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
    i.ScrollingContent.drawScrollIndicators(i.gridRenderer, screen)
}

func (i *InventoryWindow) ActionUp() {
    i.currentSelection--
    if i.currentSelection < 0 {
        i.currentSelection = len(i.items) - 1
        i.ScrollingContent.ScrollToBottom()
    } else {
        i.ScrollingContent.MakeIndexVisible(i.currentSelection)
    }
}

func (i *InventoryWindow) ActionDown() {
    i.currentSelection++
    if i.currentSelection >= len(i.items) {
        i.currentSelection = 0
        i.ScrollingContent.ScrollToTop()
    } else {
        i.ScrollingContent.MakeIndexVisible(i.currentSelection)
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
    return i.shouldClose
}

func (i *InventoryWindow) OnAvatarSwitched() {

}

func (i *InventoryWindow) ActionLeft() {

}

func (i *InventoryWindow) ActionRight() {

}

func (i *InventoryWindow) OnMouseMoved(x int, y int) (bool, Tooltip) {
    if y == i.topLeft.Y {
        return i.ButtonHolder.OnMouseMoved(x, y)
    }

    if y < i.topLeft.Y+2 || y > i.bottomRight.Y-2 {
        return false, NoTooltip{}
    }

    if x < i.topLeft.X+2 || x > i.bottomRight.X-3 {
        return false, NoTooltip{}
    }

    lineOfItem := i.getLineFromScreenLine(y)

    if lineOfItem >= len(i.items) || lineOfItem < 0 {
        return false, NoTooltip{}
    }

    i.currentSelection = lineOfItem

    item := i.items[i.currentSelection][0]
    return true, NewItemTooltip(i.gridRenderer, item, geometry.Point{X: x, Y: y})
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
        item := i.items[i.currentSelection]
        i.activateItemContextMenu(item[0])
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
            i.updateItemList()
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
       keyButton.SetFixedText("Key")
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
    i.currentSelection = 0
    i.ScrollingContent.ScrollToTop()
    i.updateItemList()
}

func (i *InventoryWindow) getCurrentFilter() func(item game.Item) bool {
    return i.filterMap[i.currentPage]
}
func (i *InventoryWindow) updateItemList() {
    party := i.engine.GetParty()
    if i.currentPage == InventoryPageAll {
        items := party.GetFilteredStackedInventory(i.getCurrentFilter())
        slices.Reverse(items)
        i.items = items
    } else {
        i.items = i.sortInventory(party.GetFilteredStackedInventory(i.getCurrentFilter()))
    }
}

func (i *InventoryWindow) drawItems(screen *ebiten.Image) {
    for y := i.topLeft.Y + 2; y < i.bottomRight.Y-2; y++ {
        relativeItemOffset := y - i.topLeft.Y - 2 + i.gridScrollOffset()
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

    sortMap := make(map[InventoryPage]func(one game.Item, two game.Item) bool)
    sortMap[InventoryPageAll] = func(one game.Item, two game.Item) bool {
        return one.Name() < two.Name()
    }
    sortMap[InventoryPageWeapons] = func(one game.Item, two game.Item) bool {
        weaponOne := one.(*game.Weapon)
        weaponTwo := two.(*game.Weapon)

        if weaponOne.IsEquipped() && !weaponTwo.IsEquipped() {
            return true
        } else if !weaponOne.IsEquipped() && weaponTwo.IsEquipped() {
            return false
        }

        if weaponOne.GetBaseDamage() > weaponTwo.GetBaseDamage() {
            return true
        } else if weaponOne.GetBaseDamage() < weaponTwo.GetBaseDamage() {
            return false
        }

        return one.Name() < two.Name()
    }

    sortMap[InventoryPageArmor] = func(one game.Item, two game.Item) bool {
        armorOne := one.(*game.Armor)
        armorTwo := two.(*game.Armor)

        if armorOne.IsEquipped() && !armorTwo.IsEquipped() {
            return true
        } else if !armorOne.IsEquipped() && armorTwo.IsEquipped() {
            return false
        }

        if armorOne.GetProtection() > armorTwo.GetProtection() {
            return true
        } else if armorOne.GetProtection() < armorTwo.GetProtection() {
            return false
        }

        return one.Name() < two.Name()
    }

    sortMap[InventoryPageAccessories] = func(one game.Item, two game.Item) bool {
        armorOne := one.(*game.Armor)
        armorTwo := two.(*game.Armor)

        if armorOne.IsEquipped() && !armorTwo.IsEquipped() {
            return true
        } else if !armorOne.IsEquipped() && armorTwo.IsEquipped() {
            return false
        }

        if armorOne.GetProtection() > armorTwo.GetProtection() {
            return true
        } else if armorOne.GetProtection() < armorTwo.GetProtection() {
            return false
        }

        return one.Name() < two.Name()
    }

    sortMap[InventoryPageConsumables] = func(one game.Item, two game.Item) bool {
        return one.Name() < two.Name()
    }

    sortMap[InventoryPageScrolls] = func(one game.Item, two game.Item) bool {
        return one.Name() < two.Name()
    }

    sortMap[InventoryPageMisc] = func(one game.Item, two game.Item) bool {
        return one.Name() < two.Name()
    }

    i.sortMap = sortMap
}

func (i *InventoryWindow) ActionCancel() {
    i.shouldClose = true
}

func (i *InventoryWindow) sortInventory(inventory [][]game.Item) [][]game.Item {
    // sort by name
    sort.SliceStable(inventory, func(one, two int) bool {
        firstItem := inventory[one][0]
        secondItem := inventory[two][0]
        return i.sortMap[i.currentPage](firstItem, secondItem)
    })
    return inventory
}

func NewInventoryWindow(engine game.Engine, gridRenderer *renderer.DualGridRenderer) *InventoryWindow {
    smallScreenSize := gridRenderer.GetSmallGridScreenSize()
    i := &InventoryWindow{
        ButtonHolder: NewButtonHolder(),
        engine:       engine,
        gridRenderer: gridRenderer,
        topLeft:      geometry.Point{X: 2, Y: 2},
        bottomRight:  geometry.Point{X: smallScreenSize.X - 2, Y: smallScreenSize.Y - 3},
    }
    neededHeight := func() int { return len(i.items) }
    availableSpace := func() geometry.Rect { return geometry.NewRect(i.topLeft.X+2, i.topLeft.Y+2, i.bottomRight.X-2, i.bottomRight.Y-2) }
    i.ScrollingContent = NewScrollingContentWithFunctions(neededHeight, availableSpace)

    i.loadFilter()
    i.createButtons()
    i.SwitchPageTo(InventoryPageAll)
    return i
}
