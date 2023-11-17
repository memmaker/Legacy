package util

import "image/color"

type MenuItem struct {
    Text        string
    Action      func()
    TextColor   color.Color
    CharIcon    int32
    TooltipText []string
}

func IndexToXY(index int, width int) (int, int) {
    return index % width, index / width
}

func XYToIndex(x int, y int, width int) int {
    return y*width + x
}
