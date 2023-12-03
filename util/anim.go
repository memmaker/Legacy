package util

import "github.com/hajimehoshi/ebiten/v2"

func GetLoopingFrameFromTick(tick uint64, delayInSeconds float64, frameCount int) int32 {
    oneSecondInTicks := float64(max(60, uint64(ebiten.ActualTPS())))
    ticksPerInterval := delayInSeconds * oneSecondInTicks
    intervallCount := float64(tick) / ticksPerInterval
    return int32(intervallCount) % int32(frameCount)
}
