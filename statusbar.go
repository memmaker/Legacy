package main

import (
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
    "strconv"
)

func (g *GridEngine) drawLowerStatusBar(screen *ebiten.Image) {
    status := g.playerParty.Status(g)
    screenSize := g.gridRenderer.GetSmallGridScreenSize()
    y := screenSize.Y - 1
    x := 0
    //divider := '█'
    for i, charStatus := range status {
        x = i * 10
        g.gridRenderer.DrawOnSmallGrid(screen, x, y, charStatus.HealthIcon)
        g.gridRenderer.DrawColoredString(screen, x+1, y, charStatus.Name, charStatus.StatusColor)
        //g.gridRenderer.DrawOnSmallGrid(screen, x+9, y, int(g.fontIndex[divider]))
    }
}
func (g *GridEngine) drawCombatActionBar(screen *ebiten.Image) {
    //TOO MUCH?
}
func (g *GridEngine) drawUpperStatusBar(screen *ebiten.Image) {
    screenSize := g.gridRenderer.GetSmallGridScreenSize()
    foodCount := g.playerParty.GetFood()
    goldCount := g.playerParty.GetGold()
    lockpickCount := g.playerParty.GetLockpicks()

    foodString := strconv.Itoa(foodCount)
    goldString := strconv.Itoa(goldCount)
    lockpickString := strconv.Itoa(lockpickCount)

    foodIcon := int32(131)
    goldIcon := int32(132)
    lockpickIcon := int32(133)
    bagIcon := int32(169)

    yPos := screenSize.Y - 2
    xPosFood := 2
    xPosLockpick := xPosFood + len(foodString) + 2
    xPosGold := screenSize.X - 4 - len(goldString)
    xPosBag := screenSize.X - 2

    g.gridRenderer.DrawOnSmallGrid(screen, xPosBag, yPos, bagIcon)

    if foodCount > 0 {
        g.gridRenderer.DrawOnSmallGrid(screen, xPosFood-1, yPos, foodIcon)
        g.gridRenderer.DrawColoredString(screen, xPosFood, yPos, foodString, color.White)
    }

    if lockpickCount > 0 {
        g.gridRenderer.DrawOnSmallGrid(screen, xPosLockpick-1, yPos, lockpickIcon)
        g.gridRenderer.DrawColoredString(screen, xPosLockpick, yPos, lockpickString, color.White)
    }

    if goldCount > 0 {
        g.gridRenderer.DrawColoredString(screen, xPosGold, yPos, goldString, color.White)
        g.gridRenderer.DrawOnSmallGrid(screen, xPosGold+len(goldString), yPos, goldIcon)
    }

    if g.playerParty.HasDefenseBuffs() {
        shieldIcon := int32(151)
        g.gridRenderer.DrawOnSmallGrid(screen, g.defenseBuffsButton.X, g.defenseBuffsButton.Y, shieldIcon)
    }

    if g.playerParty.HasOffenseBuffs() {
        swordIcon := int32(152)
        g.gridRenderer.DrawOnSmallGrid(screen, g.offenseBuffsButton.X, g.offenseBuffsButton.Y, swordIcon)
    }

}
