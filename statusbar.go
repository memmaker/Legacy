package main

import (
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
    "strconv"
)

func (g *GridEngine) drawLowerStatusBar(screen *ebiten.Image) {
    status := g.playerParty.Status()
    screenSize := g.gridRenderer.GetSmallGridScreenSize()
    y := screenSize.Y - 1
    x := 0
    //divider := '█'
    for i, charStatus := range status {
        x = i * 10
        //g.gridRenderer.DrawOnSmallGrid(screen, x, y, int(g.fontIndex[charStatus.HealthIcon]))
        g.gridRenderer.DrawColoredString(screen, x+1, y, charStatus.Name, charStatus.StatusColor)
        //g.gridRenderer.DrawOnSmallGrid(screen, x+9, y, int(g.fontIndex[divider]))
    }
}

func (g *GridEngine) drawUpperStatusBar(screen *ebiten.Image) {
    screenSize := g.gridRenderer.GetSmallGridScreenSize()
    foodCount := g.playerParty.GetFood()
    goldCount := g.playerParty.GetGold()
    lockpickCount := g.playerParty.GetLockpicks()

    foodString := strconv.Itoa(foodCount)
    goldString := strconv.Itoa(goldCount)
    lockpickString := strconv.Itoa(lockpickCount)

    foodIcon := 131
    goldIcon := 132
    lockpickIcon := 133

    yPos := screenSize.Y - 2
    xPosFood := 2
    xPosLockpick := xPosFood + len(foodString) + 2
    xPosGold := screenSize.X - 2 - len(goldString)

    g.gridRenderer.DrawColoredString(screen, xPosFood, yPos, foodString, color.White)
    g.gridRenderer.DrawColoredString(screen, xPosLockpick, yPos, lockpickString, color.White)
    g.gridRenderer.DrawColoredString(screen, xPosGold, yPos, goldString, color.White)

    g.gridRenderer.DrawOnSmallGrid(screen, xPosFood-1, yPos, foodIcon)
    g.gridRenderer.DrawOnSmallGrid(screen, xPosLockpick-1, yPos, lockpickIcon)
    g.gridRenderer.DrawOnSmallGrid(screen, xPosGold+len(goldString), yPos, goldIcon)
}
