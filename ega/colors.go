package ega

import "image/color"

// all EGA colors

/*
Default EGA 16-Color Palette
Color	Name	RGB	Binary	Decimal
0	black	#000000	000000	0
1	blue	#0000AA	000001	1
2	green	#00AA00	000010	2
3	cyan	#00AAAA	000011	3
4	red	#AA0000	000100	4
5	magenta	#AA00AA	000101	5
6	yellow / brown	#AA5500	010100	20
7	white / light gray	#AAAAAA	000111	7
8	dark gray / bright black	#555555	111000	56
9	bright blue	#5555FF	111001	57
10	bright green	#55FF55	111010	58
11	bright cyan	#55FFFF	111011	59
12	bright red	#FF5555	111100	60
13	bright magenta	#FF55FF	111101	61
14	bright yellow	#FFFF55	111110	62
15	bright white	#FFFFFF	111111	63
*/
var Black = color.RGBA{R: 0x00, G: 0x00, B: 0x00, A: 0xff}
var Blue = color.RGBA{R: 0x00, G: 0x00, B: 0xaa, A: 0xff}
var Green = color.RGBA{R: 0x00, G: 0xaa, B: 0x00, A: 0xff}
var Cyan = color.RGBA{R: 0x00, G: 0xaa, B: 0xaa, A: 0xff}
var Red = color.RGBA{R: 0xaa, G: 0x00, B: 0x00, A: 0xff}
var Magenta = color.RGBA{R: 0xaa, G: 0x00, B: 0xaa, A: 0xff}
var Yellow = color.RGBA{R: 0xaa, G: 0x55, B: 0x00, A: 0xff}
var White = color.RGBA{R: 0xaa, G: 0xaa, B: 0xaa, A: 0xff}
var BrightBlack = color.RGBA{R: 0x55, G: 0x55, B: 0x55, A: 0xff}
var BrightBlue = color.RGBA{R: 0x55, G: 0x55, B: 0xff, A: 0xff}
var BrightGreen = color.RGBA{R: 0x55, G: 0xff, B: 0x55, A: 0xff}
var BrightCyan = color.RGBA{R: 0x55, G: 0xff, B: 0xff, A: 0xff}
var BrightRed = color.RGBA{R: 0xff, G: 0x55, B: 0x55, A: 0xff}
var BrightMagenta = color.RGBA{R: 0xff, G: 0x55, B: 0xff, A: 0xff}
var BrightYellow = color.RGBA{R: 0xff, G: 0xff, B: 0x55, A: 0xff}
var BrightWhite = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
