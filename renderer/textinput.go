package renderer

import (
    "Legacy/geometry"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
    "strings"
)

type EndAction int

const (
    EndActionCancel EndAction = iota
    EndActionConfirm
)

type TextInput struct {
    prompt       string
    currentText  string
    maxLength    int
    gridPos      geometry.Point
    cursorIcon   int32
    cursorFrames int
    shouldClose  bool
    onClose      func(endedWith EndAction, text string)
    gridRenderer *DualGridRenderer
    drawBorder   bool
}

func NewTextInput(gridRenderer *DualGridRenderer, gridPos geometry.Point, maxLength int, cursorIcon int32, cursorFrameCount int, onClose func(endedWith EndAction, text string)) *TextInput {
    return &TextInput{
        gridPos:      gridPos,
        maxLength:    maxLength,
        cursorIcon:   cursorIcon,
        cursorFrames: cursorFrameCount,
        onClose:      onClose,
        gridRenderer: gridRenderer,
    }
}
func (t *TextInput) CenterHorizontallyAtY(y int) {
    neededWidth := t.neededWidth()
    smallScreenSize := t.gridRenderer.GetSmallGridScreenSize()
    t.gridPos.X = (smallScreenSize.X - neededWidth) / 2
    t.gridPos.Y = y
}

func (t *TextInput) neededWidth() int {
    return len(t.prompt) + t.maxLength + 1 // for the cursor Icon
}
func (t *TextInput) SetPrompt(prompt string) {
    t.prompt = prompt
}

func (t *TextInput) SetDrawBorder(drawBorder bool) {
    t.drawBorder = drawBorder
}
func (t *TextInput) ShouldClose() bool {
    return t.shouldClose
}
func (t *TextInput) Draw(gridRenderer *DualGridRenderer, screen *ebiten.Image, tick uint64) {
    if t.shouldClose {
        return
    }

    if t.drawBorder {
        borderTopLeft := t.gridPos.Add(geometry.Point{X: -1, Y: -1})
        borderBottomRight := t.gridPos.Add(geometry.Point{X: t.neededWidth() + 1, Y: 2})
        gridRenderer.DrawFilledBorder(screen, borderTopLeft, borderBottomRight, "")
    }

    yPos := t.gridPos.Y
    xInputStart := t.gridPos.X
    if t.prompt != "" {
        gridRenderer.DrawColoredString(screen, xInputStart, yPos, t.prompt, color.White)
        xInputStart += len(t.prompt)
    }
    gridRenderer.DrawColoredString(screen, xInputStart, yPos, t.currentText, color.White)
    xCursor := xInputStart + len(t.currentText)
    gridRenderer.DrawOnSmallGrid(screen, xCursor, yPos, t.cursorFromTick(tick))

}

func (t *TextInput) OnKeyPressed(key ebiten.Key) {
    if t.shouldClose {
        return
    }
    if key == ebiten.KeyBackspace && len(t.currentText) > 0 {
        t.currentText = t.currentText[:len(t.currentText)-1]
        return
    }

    if key == ebiten.KeyEnter {
        t.onClose(EndActionConfirm, t.currentText)
        t.shouldClose = true
        return
    }

    if key == ebiten.KeyEscape {
        t.onClose(EndActionCancel, "")
        t.shouldClose = true
        return
    }
    if key == ebiten.KeySpace {
        t.currentText += " "
        return
    }
    if len(t.currentText) < t.maxLength {
        printable := key.String()
        if strings.HasPrefix(printable, "Digit") {
            printable = printable[5:]
        }
        // only accept standard printable ascii characters
        if len(printable) != 1 || printable[0] < 32 || printable[0] > 126 {
            return
        }
        if ebiten.IsKeyPressed(ebiten.KeyShift) {
            printable = strings.ToUpper(printable)
        } else {
            printable = strings.ToLower(printable)
        }
        t.currentText += printable
    }
}

func (t *TextInput) cursorFromTick(tick uint64) int32 {
    if t.cursorFrames == 1 {
        return t.cursorIcon
    }
    delays := tick / 20
    return t.cursorIcon + int32(delays%uint64(t.cursorFrames))
}

func (t *TextInput) SetMaxLength(length int) {
    t.maxLength = length
}
