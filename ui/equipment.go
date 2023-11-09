package ui

import (
    "Legacy/ega"
    "Legacy/game"
    "Legacy/geometry"
    "Legacy/renderer"
    "Legacy/util"
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
    "strconv"
)

type EquipmentWindow struct {
    actor          *game.Actor
    gridRenderer   *renderer.DualGridRenderer
    shouldClose    bool
    armorSlotLines map[int]game.ArmorSlot
    engine         game.Engine
    lastMouseLine  int
}

func (e *EquipmentWindow) OnAvatarSwitched() {
    e.actor = e.engine.GetAvatar()
}

func NewEquipmentWindow(engine game.Engine, partyMember *game.Actor, gridRenderer *renderer.DualGridRenderer) *EquipmentWindow {
    return &EquipmentWindow{
        engine:       engine,
        actor:        partyMember,
        gridRenderer: gridRenderer,
        armorSlotLines: map[int]game.ArmorSlot{
            // 7 -> right hand
            // 8 -> left hand
            // 9 -> ranged
            11: game.ArmorSlotHelmet,
            12: game.ArmorSlotBreastPlate,
            13: game.ArmorSlotShoes,
            15: game.AccessorySlotRobe,
            16: game.AccessorySlotRingLeft,
            17: game.AccessorySlotRingRight,
            18: game.AccessorySlotAmuletOne,
            19: game.AccessorySlotAmuletTwo,
        },
    }
}

func (e *EquipmentWindow) OnMouseClicked(x int, y int) bool {
    if y < 7 || y > 19 {
        return false
    }
    smallScreen := e.gridRenderer.GetSmallGridScreenSize()

    if x < 4 || x > smallScreen.X-4 {
        return false
    }

    if y == 7 {
        e.onRightHandClicked()
    } else if y == 8 {
        e.onLeftHandClicked()
    } else if y == 9 {
        e.onRangedClicked()
    }

    if armorSlot, clicked := e.armorSlotLines[y]; clicked {
        e.onArmorSlotClicked(armorSlot)
    }
    return false
}

func (e *EquipmentWindow) Draw(screen *ebiten.Image) {
    smallScreen := e.gridRenderer.GetSmallGridScreenSize()

    // border
    topLeft := geometry.Point{X: 2, Y: 2}
    bottomRight := geometry.Point{X: smallScreen.X - 2, Y: smallScreen.Y - 3}
    e.gridRenderer.DrawFilledBorder(screen, topLeft, bottomRight, e.actor.Name())

    // icon
    icon := e.actor.Icon(0)
    iconScreenPosX, iconScreenPosY := e.gridRenderer.SmallCellToScreen(4, 4)
    e.gridRenderer.DrawBigOnScreenWithAtlasNameAndTint(screen, iconScreenPosX, iconScreenPosY, renderer.AtlasEntities, icon, color.White)

    mainStatsLeft := []util.TableRow{
        {Label: "Health", Columns: []string{strconv.Itoa(e.actor.GetHealth())}},
        {Label: "Level", Columns: []string{strconv.Itoa(e.actor.GetLevel())}},
    }
    mainStatsRight := []util.TableRow{
        {Label: "Damage", Columns: []string{strconv.Itoa(e.actor.GetMeleeDamage())}},
        {Label: "Armor", Columns: []string{strconv.Itoa(e.actor.GetTotalArmor())}},
    }
    leftStats := util.TableLayout(mainStatsLeft)
    rightStats := util.TableLayout(mainStatsRight)

    rightWidth := len(rightStats[0])

    for i, row := range leftStats {
        e.gridRenderer.DrawColoredString(screen, topLeft.X+5, topLeft.Y+2+i, row, color.White)
    }
    for i, row := range rightStats {
        e.gridRenderer.DrawColoredString(screen, bottomRight.X-rightWidth-2, topLeft.Y+2+i, row, color.White)
    }

    // weapons
    currentY := topLeft.Y + 5
    drawColor := color.Color(color.White)

    rightHandItem, rhExists := e.actor.GetRightHandItem()
    rightHandIcon := int32(163)
    rightHandLabel := "Right Hand"
    if rhExists {
        rightHandIcon = rightHandItem.InventoryIcon()
        rightHandLabel = rightHandItem.Name()
        drawColor = rightHandItem.TintColor()
    }
    e.drawItemWithIcon(screen, rightHandIcon, rightHandLabel, currentY, drawColor)
    currentY += 1
    drawColor = color.White

    leftHandItem, lhExists := e.actor.GetLeftHandItem()
    leftHandIcon := int32(162)
    leftHandLabel := "Left Hand"
    if lhExists {
        leftHandIcon = leftHandItem.InventoryIcon()
        leftHandLabel = leftHandItem.Name()
        drawColor = leftHandItem.TintColor()
    }
    e.drawItemWithIcon(screen, leftHandIcon, leftHandLabel, currentY, drawColor)
    currentY += 1
    drawColor = color.White

    rangedItem, rngExists := e.actor.GetRangedItem()
    rangedIcon := int32(32)
    rangedLabel := "Ranged"
    if rngExists {
        rangedIcon = rangedItem.InventoryIcon()
        rangedLabel = rangedItem.Name()
        drawColor = rangedItem.TintColor()
    }
    e.drawItemWithIcon(screen, rangedIcon, rangedLabel, currentY, drawColor)
    currentY += 2
    drawColor = color.White

    // armor
    helmetItem, helmExists := e.actor.GetHelmet()
    helmetIcon := int32(32)
    helmetLabel := "Helmet"
    if helmExists {
        helmetIcon = helmetItem.InventoryIcon()
        helmetLabel = helmetItem.Name()
        drawColor = helmetItem.TintColor()
    }
    e.drawItemWithIcon(screen, helmetIcon, helmetLabel, currentY, drawColor)
    currentY += 1
    drawColor = color.White

    armorItem, armorExists := e.actor.GetArmorBreastPlate()
    armorIcon := int32(32)
    armorLabel := "Armor"
    if armorExists {
        armorIcon = armorItem.InventoryIcon()
        armorLabel = armorItem.Name()
        drawColor = armorItem.TintColor()
    }
    e.drawItemWithIcon(screen, armorIcon, armorLabel, currentY, drawColor)
    currentY += 1
    drawColor = color.White

    bootsItem, shExists := e.actor.GetShoes()
    bootsIcon := int32(32)
    bootsLabel := "Boots"
    if shExists {
        bootsIcon = bootsItem.InventoryIcon()
        bootsLabel = bootsItem.Name()
        drawColor = bootsItem.TintColor()
    }
    e.drawItemWithIcon(screen, bootsIcon, bootsLabel, currentY, drawColor)
    currentY += 2
    drawColor = color.White

    // accessories
    robeItem, rbExists := e.actor.GetRobe()
    robeIcon := int32(32)
    robeLabel := "Robe"
    if rbExists {
        robeIcon = robeItem.InventoryIcon()
        robeLabel = robeItem.Name()
        drawColor = robeItem.TintColor()
    }
    e.drawItemWithIcon(screen, robeIcon, robeLabel, currentY, drawColor)
    currentY += 1
    drawColor = color.White

    ringLeftItem, rlExists := e.actor.GetRingLeft()
    ringLeftIcon := int32(32)
    ringLeftLabel := "Ring Left"
    if rlExists {
        ringLeftIcon = ringLeftItem.InventoryIcon()
        ringLeftLabel = ringLeftItem.Name()
        drawColor = ringLeftItem.TintColor()
    }
    e.drawItemWithIcon(screen, ringLeftIcon, ringLeftLabel, currentY, drawColor)
    currentY += 1
    drawColor = color.White

    ringRightItem, rrExists := e.actor.GetRingRight()
    ringRightIcon := int32(32)
    ringRightLabel := "Ring Right"
    if rrExists {
        ringRightIcon = ringRightItem.InventoryIcon()
        ringRightLabel = ringRightItem.Name()
        drawColor = ringRightItem.TintColor()
    }
    e.drawItemWithIcon(screen, ringRightIcon, ringRightLabel, currentY, drawColor)
    currentY += 1
    drawColor = color.White

    amuletOneItem, amoExists := e.actor.GetAmuletOne()
    amuletOneIcon := int32(32)
    amuletOneLabel := "Amulet One"
    if amoExists {
        amuletOneIcon = amuletOneItem.InventoryIcon()
        amuletOneLabel = amuletOneItem.Name()
        drawColor = amuletOneItem.TintColor()
    }
    e.drawItemWithIcon(screen, amuletOneIcon, amuletOneLabel, currentY, drawColor)
    currentY += 1
    drawColor = color.White

    amuletTwoItem, amtExists := e.actor.GetAmuletTwo()
    amuletTwoIcon := int32(32)
    amuletTwoLabel := "Amulet Two"
    if amtExists {
        amuletTwoIcon = amuletTwoItem.InventoryIcon()
        amuletTwoLabel = amuletTwoItem.Name()
        drawColor = amuletTwoItem.TintColor()
    }
    e.drawItemWithIcon(screen, amuletTwoIcon, amuletTwoLabel, currentY, drawColor)
}

func (e *EquipmentWindow) drawItemWithIcon(screen *ebiten.Image, icon int32, label string, yOffset int, tintColor color.Color) {
    currentX := 4
    drawColor := tintColor
    if icon == 32 || icon == 163 || icon == 162 {
        drawColor = ega.BrightBlack
    }
    if yOffset == e.lastMouseLine {
        drawColor = ega.BrightGreen
    }
    e.gridRenderer.DrawOnSmallGrid(screen, currentX, yOffset, icon)
    e.gridRenderer.DrawColoredString(screen, currentX+2, yOffset, label, drawColor)
}

func (e *EquipmentWindow) ActionUp() {

}

func (e *EquipmentWindow) ActionDown() {

}

func (e *EquipmentWindow) ActionLeft() {

}

func (e *EquipmentWindow) ActionRight() {

}

func (e *EquipmentWindow) ActionConfirm() {

}

func (e *EquipmentWindow) OnMouseMoved(x int, y int) renderer.Tooltip {
    e.lastMouseLine = y
    return renderer.NoTooltip{}
}

func (e *EquipmentWindow) ShouldClose() bool {
    return e.shouldClose
}

func (e *EquipmentWindow) onRightHandClicked() {
    equipAction := func(item game.Item) {
        e.actor.EquipWeapon(item.(*game.Weapon))
    }
    e.chooseEquipment(equipAction, func(item game.Item) bool {
        weapon, isWeapon := item.(*game.Weapon)
        return isWeapon && e.actor.CanEquip(weapon) && !weapon.IsRanged()
    })
}

func (e *EquipmentWindow) onLeftHandClicked() {
    equipAction := func(item game.Item) {
        e.actor.EquipWeapon(item.(*game.Weapon))
    }
    e.chooseEquipment(equipAction, func(item game.Item) bool {
        weapon, isWeapon := item.(*game.Weapon)
        return isWeapon && e.actor.CanEquip(weapon) && !weapon.IsRanged()
    })
}

func (e *EquipmentWindow) onRangedClicked() {
    equipAction := func(item game.Item) {
        e.actor.EquipWeapon(item.(*game.Weapon))
    }
    e.chooseEquipment(equipAction, func(item game.Item) bool {
        weapon, isWeapon := item.(*game.Weapon)
        return isWeapon && e.actor.CanEquip(weapon) && weapon.IsRanged()
    })
}

func (e *EquipmentWindow) onArmorSlotClicked(slot game.ArmorSlot) {
    equipAction := func(item game.Item) {
        e.actor.EquipArmor(item.(*game.Armor), slot)
    }
    e.chooseEquipment(equipAction, func(item game.Item) bool {
        armor, isArmor := item.(*game.Armor)
        return isArmor && armor.GetSlot().IsEqualTo(slot) && e.actor.CanEquip(armor)
    })
}
func (e *EquipmentWindow) chooseEquipment(equipAction func(item game.Item), filter func(item game.Item) bool) {
    party := e.engine.GetParty()

    equippableItems := party.GetFilteredInventory(filter)
    if len(equippableItems) == 0 {
        e.engine.Print("No items to equip")
        return
    }

    menuItems := make([]renderer.MenuItem, 0)

    for _, item := range equippableItems {
        equippable := item.(game.Equippable)
        armorLabel := equippable.Name()
        if equippable.IsEquipped() {
            wearerIcon := string(party.GetMemberIcon(equippable.GetWearer()))
            armorLabel = fmt.Sprintf("%s (%s)", armorLabel, wearerIcon)
        }
        menuItem := renderer.MenuItem{
            CharIcon:  equippable.InventoryIcon(),
            TextColor: equippable.TintColor(),
            Text:      armorLabel,
            Action: func() {
                equipAction(equippable)
                e.engine.OpenEquipmentDetails(e.actor)
            },
        }
        menuItems = append(menuItems, menuItem)
    }

    e.engine.OpenMenu(menuItems)
}
