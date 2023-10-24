package game

import (
    "Legacy/geometry"
    "fmt"
)

type Book struct {
    pos     geometry.Point
    icon    int
    title   string
    useFunc func()
}

func (b *Book) Use() {
    b.useFunc()
}

func (b *Book) ShortDescription() string {
    return fmt.Sprintf("Book: %s", b.title)
}
func (b *Book) Pos() geometry.Point {
    return b.pos
}

func (b *Book) Icon() int {
    return b.icon
}

func (b *Book) SetPos(point geometry.Point) {
    b.pos = point
}

func NewBook(title string, onUse func()) *Book {
    return &Book{
        icon:    181,
        title:   title,
        useFunc: onUse,
    }
}
