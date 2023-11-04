package game

import (
    "Legacy/ega"
    "Legacy/geometry"
    "Legacy/gridmap"
    "Legacy/renderer"
    "Legacy/util"
    "fmt"
    "image/color"
    "strconv"
)

type Party struct {
    members            []*Actor
    partyInventory     [][]Item
    keys               map[string]*Key
    gridMap            *gridmap.GridMap[*Actor, Item, Object]
    fov                *geometry.FOV
    gold               int
    food               int
    lockpicks          int
    stepsBeforeRest    int
    needRestAfterSteps int
}

func (p *Party) Pos() geometry.Point {
    return p.members[0].Pos()
}

func NewParty(leader *Actor) *Party {
    stepsBeforeRest := 100
    p := &Party{
        members:            []*Actor{leader},
        partyInventory:     [][]Item{},
        keys:               make(map[string]*Key),
        fov:                geometry.NewFOV(geometry.NewRect(-6, -6, 6, 6)),
        gold:               1000,
        food:               3,
        needRestAfterSteps: stepsBeforeRest,
        stepsBeforeRest:    stepsBeforeRest,
    }
    leader.SetParty(p)
    return p
}

func (p *Party) AddItem(item Item) {
    if key, ok := item.(*Key); ok {
        p.keys[key.key] = key
    } else {
        p.addToInventory(item)
    }
    item.SetHolder(p)
}

func (p *Party) addToInventory(item Item) {
    for i, it := range p.partyInventory {
        if it[0].CanStackWith(item) {
            p.partyInventory[i] = append(p.partyInventory[i], item)
            return
        }
    }
    p.partyInventory = append(p.partyInventory, []Item{item})
}

func (p *Party) RemoveItem(item Item) {
    if key, ok := item.(*Key); ok {
        delete(p.keys, key.key)
        item.SetHolder(nil)
    } else {
        for i, it := range p.partyInventory {
            if it[0] == item {
                p.partyInventory[i] = append(it[:0], it[1:]...)
                if len(p.partyInventory[i]) == 0 {
                    p.partyInventory = append(p.partyInventory[:i], p.partyInventory[i+1:]...)
                }
                item.SetHolder(nil)
                return
            }
            if it[0].CanStackWith(item) {
                for j, stackItem := range it {
                    if stackItem == item {
                        p.partyInventory[i] = append(it[:j], it[j+1:]...)
                        item.SetHolder(nil)
                        return
                    }
                }
            }
        }
    }
}

type MemberStatus struct {
    Name        string
    HealthIcon  int32
    StatusColor color.Color
}

func (p *Party) Status(engine Engine) []MemberStatus {
    var result []MemberStatus
    for _, member := range p.members {
        nameColor := ega.BrightWhite
        canLevelUp, _ := engine.CanLevelUp(member)
        if !member.IsAlive() {
            nameColor = ega.BrightRed
        } else if canLevelUp {
            nameColor = ega.BrightMagenta
        }
        result = append(result, MemberStatus{
            Name:        member.Name(),
            HealthIcon:  healthToIcon(member.health, member.maxHealth),
            StatusColor: nameColor,
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

func (p *Party) GetStackedInventory() [][]Item {
    return p.partyInventory
}

func (p *Party) GetFlatInventory() []Item {
    var result []Item
    for _, stack := range p.partyInventory {
        result = append(result, stack...)
    }
    return result
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
    p.food -= len(p.members)
    for _, member := range p.members {
        member.FullRest()
    }
    return true
}

func (p *Party) NeedsRest() bool {
    for _, member := range p.members {
        if member.health < member.maxHealth || member.HasNegativeBuffs() {
            return true
        }
    }
    return false
}

func (p *Party) AddFood(amount int) {
    p.food += amount
}

func (p *Party) AddGold(amount int) {
    p.gold += amount
}

func (p *Party) GetFood() int {
    return p.food
}

func (p *Party) AddLockpicks(amount int) {
    p.lockpicks += amount
}

func (p *Party) GetLockpicks() int {
    return p.lockpicks
}

func (p *Party) RemoveLockpicks(amount int) {
    p.lockpicks -= amount
}

func (p *Party) GetSpells() []*Spell {
    var result []*Spell
    for _, item := range p.partyInventory {
        if scroll, ok := item[0].(*Scroll); ok {
            if scroll.spell != nil {
                result = append(result, scroll.spell)
            }
        }
    }
    return result
}

func (p *Party) IsDefeated() bool {
    for _, member := range p.members {
        if member.IsAlive() {
            return false
        }
    }
    return true
}

func (p *Party) GetPartyOverview() []string {

    tableData := []util.TableRow{
        {Label: "Name", Columns: []string{"HP", "MP", "Arm.", "Dmg."}},
        {Label: "----", Columns: []string{"--", "--", "----", "----"}},
    }

    for i := 0; i < len(p.members); i++ {
        member := p.members[i]
        tableData = append(tableData, util.TableRow{
            Label: member.Name(),
            Columns: []string{
                strconv.Itoa(member.health),
                strconv.Itoa(member.mana),
                strconv.Itoa(member.GetTotalArmor()),
                strconv.Itoa(member.GetMeleeDamage()),
            }})
    }

    return util.TableLayout(tableData)
}

func (p *Party) RemoveMember(member *Actor) {
    for i, m := range p.members {
        if m == member {
            p.members = append(p.members[:i], p.members[i+1:]...)
            return
        }
    }
}

func (p *Party) GetMemberIndex(wearer ItemWearer) int {
    for i, member := range p.members {
        if member == wearer {
            return i
        }
    }
    return -1
}

func (p *Party) GetFinanceOverview(engine Engine) []string {
    var valueOfCarriedItems int
    var valueOfWornItems int
    for _, itemStack := range p.partyInventory {
        firstItemOfStack := itemStack[0]
        stackSize := len(itemStack)
        if wearable, ok := firstItemOfStack.(Wearable); ok && wearable.IsEquipped() {
            valueOfWornItems += wearable.GetValue() * stackSize
        } else {
            valueOfCarriedItems += firstItemOfStack.GetValue() * stackSize
        }
    }
    rules := engine.GetRules()
    foodValue := p.food * rules.GetBaseValueOfFood()
    lockpickValue := p.lockpicks * rules.GetBaseValueOfLockpick()
    netWorth := p.gold + valueOfCarriedItems + valueOfWornItems + foodValue + lockpickValue

    tableData := []util.TableRow{
        {Label: "Gold", Columns: []string{moneyFormat(p.gold)}},
        {Label: "Food", Columns: []string{moneyFormat(foodValue)}},
        {Label: "Lockpicks", Columns: []string{moneyFormat(lockpickValue)}},
        {Label: "Carried items", Columns: []string{moneyFormat(valueOfCarriedItems)}},
        {Label: "Worn items", Columns: []string{moneyFormat(valueOfWornItems)}},
        {Label: "Net worth", Columns: []string{moneyFormat(netWorth)}},
    }
    return util.TableLayout(tableData)
}

func (p *Party) GetKeys() []*Key {
    var result []*Key
    for _, key := range p.keys {
        result = append(result, key)
    }
    return result
}

func (p *Party) SetFood(foodCount int) {
    p.food = foodCount
}

func (p *Party) SetGold(amount int) {
    p.gold = amount
}

func (p *Party) SetLockpicks(amount int) {
    p.lockpicks = amount
}

func (p *Party) GetCurrentMapName() string {
    return p.gridMap.GetName()
}

func (p *Party) AddXPForEveryone(xp int) {
    for _, member := range p.members {
        member.AddXP(xp)
    }
}

func (p *Party) GetDefenseBuffs() []string {
    var result []string
    for _, member := range p.members {
        if !member.HasDefenseBuffs() {
            continue
        }
        result = append(result, member.Name())
        result = append(result, member.GetDefenseBuffsString()...)
    }
    return result
}

func (p *Party) GetOffenseBuffs() []string {
    var result []string
    for _, member := range p.members {
        if !member.HasOffenseBuffs() {
            continue
        }
        result = append(result, member.Name())
        result = append(result, member.GetOffenseBuffsString()...)
    }
    return result
}

func (p *Party) HasOffenseBuffs() bool {
    for _, member := range p.members {
        if member.HasOffenseBuffs() {
            return true
        }
    }
    return false
}

func (p *Party) HasDefenseBuffs() bool {
    for _, member := range p.members {
        if member.HasDefenseBuffs() {
            return true
        }
    }
    return false
}

func (p *Party) NeedsRestAfterMovement() bool {
    p.stepsBeforeRest--
    if p.stepsBeforeRest <= 0 {
        p.stepsBeforeRest = p.needRestAfterSteps
        return true
    }
    return false
}

func (p *Party) HasGold(cost int) bool {
    return p.gold >= cost
}

func (p *Party) GetNameOfBreakingTool() string {
    for _, itemStack := range p.partyInventory {
        firstItemOfStack := itemStack[0]
        if breakingTool, ok := firstItemOfStack.(*Tool); ok && breakingTool.kind == ToolTypePickaxe {
            return breakingTool.Name()
        }
    }
    return ""
}

func moneyFormat(value int) string {
    return strconv.Itoa(value) + "g"
}

func healthToIcon(health, maxhealth int) int32 {
    startIndex := int32(137)
    // 137 = 100%
    // 138 = 90%
    // 139 = 80%
    // 140 = 70%
    // 141 = 60%
    // ...
    ratio := float64(health) / float64(maxhealth)
    if ratio > 0.9 {
        return startIndex
    } else if ratio > 0.8 {
        return startIndex + 1
    } else if ratio > 0.7 {
        return startIndex + 2
    } else if ratio > 0.6 {
        return startIndex + 3
    } else if ratio > 0.5 {
        return startIndex + 4
    } else if ratio > 0.4 {
        return startIndex + 5
    } else if ratio > 0.3 {
        return startIndex + 6
    } else if ratio > 0.2 {
        return startIndex + 7
    } else if ratio > 0.1 {
        return startIndex + 8
    } else {
        return startIndex + 9
    }
}
