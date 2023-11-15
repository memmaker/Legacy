package util

import "image/color"

type MenuItem struct {
    Text      string
    Action    func()
    TextColor color.Color
    CharIcon  int32
}
