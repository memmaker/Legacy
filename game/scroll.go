package game

import (
    "Legacy/renderer"
    "fmt"
    "image/color"
)

type Scroll struct {
    BaseItem
    icon     int
    filename string
    spell    *Spell
}

func (b *Scroll) CanStackWith(other Item) bool {
    if otherScroll, ok := other.(*Scroll); ok {
        return b.spell == otherScroll.spell && b.filename == otherScroll.filename && b.name == otherScroll.name && b.icon == otherScroll.icon
    } else {
        return false
    }
}

func (b *Scroll) TintColor() color.Color {
    return color.White
}

func (b *Scroll) GetContextActions(engine Engine) []renderer.MenuItem {
    actions := inventoryItemActions(b, engine)
    actions = append(actions, renderer.MenuItem{
        Text:   fmt.Sprintf("Read \"%s\"", b.name),
        Action: func() { b.read(engine) },
    })
    if b.spell != nil {
        actions = append(actions, renderer.MenuItem{
            Text:   fmt.Sprintf("Cast \"%s\"", b.spell.name),
            Action: func() { b.spell.Cast(engine, engine.GetAvatar()) },
        })
    }
    return actions
}

func (b *Scroll) read(engine Engine) {
    text := engine.GetTextFile(b.filename)
    engine.ShowScrollableText(text, color.White)
}

func (b *Scroll) SetSpell(spell *Spell) {
    b.spell = spell
}

func (b *Scroll) Icon(uint64) int {
    return b.icon
}
func NewScroll(title, filename string) *Scroll {
    return &Scroll{
        BaseItem: BaseItem{
            name: title,
        },
        icon:     181,
        filename: filename,
    }
}
