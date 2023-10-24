package game

import "image/color"

type Party struct {
    members        []*Actor
    partyInventory []Item
}

func NewParty(leader *Actor) *Party {
    return &Party{
        members:        []*Actor{leader},
        partyInventory: []Item{},
    }
}

func (p *Party) AddItem(item Item) {
    p.partyInventory = append(p.partyInventory, item)
}

func (p *Party) RemoveItem(item Item) {
    for i, it := range p.partyInventory {
        if it == item {
            p.partyInventory = append(p.partyInventory[:i], p.partyInventory[i+1:]...)
            return
        }
    }
}

type MemberStatus struct {
    Name        string
    HealthIcon  rune
    StatusColor color.Color
}

func (p *Party) Status() []MemberStatus {
    var result []MemberStatus
    for _, member := range p.members {
        result = append(result, MemberStatus{
            Name:        paddedName(member.Name()),
            HealthIcon:  healthToIcon(member.Health),
            StatusColor: color.White,
        })
    }
    return result
}

func paddedName(name string) string {
    if len(name) < 8 {
        delta := 8 - len(name)
        for i := 0; i < delta; i++ {
            name += " "
        }
        return name
    } else if len(name) > 8 {
        return name[:8]
    }
    return name
}

func (p *Party) GetMember(index int) *Actor {
    if index < 0 || index >= len(p.members) {
        return nil
    }
    return p.members[index]
}

func (p *Party) GetInventoryDetails() []string {
    var result []string
    for _, item := range p.partyInventory {
        result = append(result, item.ShortDescription())
    }
    return result
}

func healthToIcon(health int) rune {
    switch {
    case health < 7:
        return 'Ө'
    case health < 4:
        return 'ӽ'
    default:
        return 'ө'
    }
}
