package game

import (
    "Legacy/recfile"
    "Legacy/util"
    "fmt"
    "image/color"
)

type Scroll struct {
    BaseItem
    icon           int32
    filename       string
    spell          *Spell
    wearer         ItemWearer
    autoLayoutText bool
}

func (b *Scroll) SetAutoLayoutText(value bool) {
    b.autoLayoutText = value
}
func (b *Scroll) GetTooltipLines() []string {
    if b.spell != nil {
        return []string{fmt.Sprintf("Scroll of %s", b.spell.name)}
    } else {
        return []string{fmt.Sprintf("Scroll of %s", b.name)}
    }
}

func (b *Scroll) GetSlot() ItemSlot {
    return ItemSlotScroll
}

func (b *Scroll) IsBetterThan(other Wearable) bool {
    return false
}

func (b *Scroll) InventoryIcon() int32 {
    return 175
}

func (b *Scroll) SetWearer(wearer ItemWearer) {
    b.wearer = wearer
}

func (b *Scroll) IsEquipped() bool {
    return b.wearer != nil
}

func (b *Scroll) GetValue() int {
    if b.spell != nil {
        return b.spell.GetValue()
    }
    return 1
}

func (b *Scroll) GetWearer() ItemWearer {
    return b.wearer
}

func (b *Scroll) CanStackWith(other Item) bool {
    if otherScroll, ok := other.(*Scroll); ok {
        return b.spell.Name() == otherScroll.spell.Name() && b.filename == otherScroll.filename && b.name == otherScroll.name && b.icon == otherScroll.icon && b.wearer == otherScroll.wearer
    } else {
        return false
    }
}

func (b *Scroll) TintColor() color.Color {
    return color.White
}

func (b *Scroll) GetContextActions(engine Engine) []util.MenuItem {
    actions := inventoryItemActions(b, engine)
    actions = append(actions, util.MenuItem{
        Text:   fmt.Sprintf("Read \"%s\"", b.name),
        Action: func() { b.read(engine) },
    })
    if b.spell != nil {
        if b.spell.IsTargeted() {
            actions = append(actions, util.MenuItem{
                Text:   fmt.Sprintf("Execute \"%s\"", b.spell.name),
                Action: func() { engine.PlayerStartsOffensiveSpell(engine.GetAvatar(), b.spell) },
            })
        } else {
            actions = append(actions, util.MenuItem{
                Text:   fmt.Sprintf("Execute \"%s\"", b.spell.name),
                Action: func() { b.spell.Execute(engine, engine.GetAvatar()) },
            })
        }

        if engine.IsPlayerControlled(b.GetHolder()) {
            actions = append(actions, util.MenuItem{
                Text: "Equip",
                Action: func() {
                    engine.ShowEquipMenu(b)
                },
            })
        }
    }
    return actions
}

func (b *Scroll) GetEmbeddedActions() []Action {
    if b.spell != nil {
        return []Action{b.spell}
    }
    return []Action{}
}

func (b *Scroll) read(engine Engine) {
    text := engine.GetScrollFile(b.filename)
    engine.ShowScrollableText(text, color.White, b.autoLayoutText)
}

func (b *Scroll) SetSpell(spell *Spell) {
    b.spell = spell
}

func (b *Scroll) Icon(uint64) int32 {
    return b.icon
}

func (b *Scroll) Unequip() {
    if b.wearer != nil {
        b.wearer.Unequip(b)
    }
}
func NewScroll(title, filename string) *Scroll {
    return &Scroll{
        BaseItem: BaseItem{
            name: title,
        },
        icon:           181,
        filename:       filename,
        autoLayoutText: true,
    }
}

func NewSpellScroll(title, filename string, spell *Spell) *Scroll {
    return &Scroll{
        BaseItem: BaseItem{
            name: title,
        },
        icon:           181,
        filename:       filename,
        spell:          spell,
        autoLayoutText: true,
    }
}
func NewSpellScrollFromSpellName(spellName string) *Scroll {
    spell := NewSpellFromName(spellName)
    return NewSpellScroll(spell.GetScrollTitle(), spell.GetScrollFile(), spell)
}
func NewScrollFromPredicate(encoded recfile.StringPredicate) *Scroll {
    // format is scroll title, filename, spell name
    return NewSpellScroll(
        encoded.GetString(0),
        encoded.GetString(1),
        NewSpellFromName(encoded.GetString(2)),
    )
}
func (b *Scroll) Encode() string {
    return recfile.ToPredicate("scroll", b.name, b.filename, b.spell.name)
}
