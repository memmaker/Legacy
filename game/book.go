package game

import (
    "Legacy/renderer"
    "fmt"
    "image/color"
)

type Book struct {
    BaseItem
    icon     int
    filename string
}

func (b *Book) TintColor() color.Color {
    return color.White
}

func (b *Book) GetContextActions(engine Engine) []renderer.MenuItem {
    actions := inventoryItemActions(b, engine)
    actions = append(actions, renderer.MenuItem{
        Text:   fmt.Sprintf("Read \"%s\"", b.name),
        Action: func() { b.read(engine) },
    })
    return actions
}

func (b *Book) read(engine Engine) {
    text := engine.GetTextFile(b.filename)
    engine.ShowScrollableText(text, color.White)
}

func (b *Book) Icon() int {
    return b.icon
}
func NewBook(title, filename string) *Book {
    return &Book{
        BaseItem: BaseItem{
            name: title,
        },
        icon:     181,
        filename: filename,
    }
}
