package game

import (
    "Legacy/geometry"
    "Legacy/gridmap"
    "Legacy/renderer"
    "fmt"
    "image/color"
)

type Party struct {
    members        []*Actor
    partyInventory []Item
    keys           map[string]*Key
    gridMap        *gridmap.GridMap[*Actor, Item, Object]
    fov            *geometry.FOV
    gold           int
    food           int
}

func (p *Party) Pos() geometry.Point {
    return p.members[0].Pos()
}

func NewParty(leader *Actor) *Party {
    p := &Party{
        members:        []*Actor{leader},
        partyInventory: []Item{},
        keys:           make(map[string]*Key),
        fov:            geometry.NewFOV(geometry.NewRect(-6, -6, 6, 6)),
        gold:           1000,
        food:           3,
    }
    leader.SetParty(p)
    return p
}

func (p *Party) AddItem(item Item) {
    if key, ok := item.(*Key); ok {
        p.keys[key.key] = key
    } else {
        p.partyInventory = append(p.partyInventory, item)
    }
    item.SetHolder(p)
}

func (p *Party) RemoveItem(item Item) {
    if key, ok := item.(*Key); ok {
        delete(p.keys, key.key)
    } else {
        for i, it := range p.partyInventory {
            if it == item {
                p.partyInventory = append(p.partyInventory[:i], p.partyInventory[i+1:]...)
                return
            }
        }
    }
    item.SetHolder(nil)
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

func (p *Party) GetInventory() []Item {
    allItems := p.partyInventory
    for _, key := range p.keys {
        allItems = append(allItems, key)
    }
    return allItems
}

func (p *Party) GetMembers() []*Actor {
    return p.members
}

func (p *Party) AddMember(npc *Actor) {
    if len(p.members) < 4 {
        p.members = append(p.members, npc)
        npc.SetParty(p)
    }
}

func (p *Party) Move(relativeMovement geometry.Point) {
    leader := p.members[0]
    leaderPos := leader.Pos()
    newPos := leaderPos.Add(relativeMovement)
    if !p.gridMap.Contains(newPos) {
        return
    }
    if p.gridMap.IsActorAt(newPos) {
        actorAtDest := p.gridMap.ActorAt(newPos)
        if p.IsMember(actorAtDest) {
            p.gridMap.SwapPositions(leader, actorAtDest)
            return
        }
    }

    p.gridMap.MoveActor(leader, newPos)

    for i := 1; i < len(p.members); i++ {
        follower := p.members[i]
        followerPos := follower.Pos()
        p.gridMap.MoveActor(follower, leaderPos)
        leaderPos = followerPos
    }
}

func (p *Party) SetGridMap(g *gridmap.GridMap[*Actor, Item, Object]) {
    p.gridMap = g
}

func (p *Party) HasKey(key string) bool {
    _, ok := p.keys[key]
    return ok
}

func (p *Party) IsFull() bool {
    return len(p.members) == 4
}

func (p *Party) IsMember(npc *Actor) bool {
    for _, member := range p.members {
        if member == npc {
            return true
        }
    }
    return false
}

func (p *Party) HasFollowers() bool {
    return len(p.members) > 1
}

func (p *Party) GetFoV() *geometry.FOV {
    return p.fov
}

func (p *Party) GetSplitActions(g Engine) []renderer.MenuItem {
    var items []renderer.MenuItem
    if p.HasFollowers() {
        for _, m := range p.members {
            member := m
            items = append(items, renderer.MenuItem{
                Text:   fmt.Sprintf("Control \"%s\"", member.Name()),
                Action: func() { g.SwitchAvatarTo(member) },
            })
        }
    }
    return items
}

func (p *Party) GetGold() int {
    return p.gold
}

func (p *Party) RemoveGold(price int) {
    p.gold -= price
}

func (p *Party) TryRest() bool {
    if p.food < len(p.members) {
        return false
    }
    for _, member := range p.members {
        member.Health = 10
    }
    return true
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
