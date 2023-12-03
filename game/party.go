package game

import (
    "Legacy/ega"
    "Legacy/geometry"
    "Legacy/gridmap"
    "Legacy/util"
    "fmt"
    "image/color"
    "strconv"
)

type Party struct {
    members            []*Actor
    partyInventory     [][]Item
    keys               map[string][]*Key
    gridMap            *gridmap.GridMap[*Actor, Item, Object]
    fov                *geometry.FOV
    gold               int
    food               int
    lockpicks          int
    stepsBeforeRest    int
    rules              *Rules
    prevPositions      []geometry.Point
    currentVehicle     *Vehicle
    splitControlled    *Actor
    usedKeys           map[string]bool
    activeSpellEffects map[OngoingSpellEffect]int
}

func (p *Party) Name() string {
    return "Party"
}

func (p *Party) Pos() geometry.Point {
    return p.members[0].Pos()
}

func NewParty(leader *Actor) *Party {
    p := &Party{
        members:            []*Actor{leader},
        partyInventory:     [][]Item{},
        keys:               make(map[string][]*Key),
        usedKeys:           make(map[string]bool),
        fov:                geometry.NewFOV(geometry.NewRect(-6, -6, 6, 6)),
        activeSpellEffects: make(map[OngoingSpellEffect]int),
    }
    leader.OnAddedToParty(p)
    return p
}

func (p *Party) InitWithRules(rules *Rules) {
    p.rules = rules
    p.stepsBeforeRest = rules.GetStepsBeforeRest()
    p.food = rules.GetPartyStartFood()
    p.gold = rules.GetPartyStartGold()
}
func (p *Party) SetRules(rules *Rules) {
    p.rules = rules
}
func (p *Party) AddItem(item Item) {
    if key, ok := item.(*Key); ok {
        if _, keyExists := p.keys[key.key]; !keyExists {
            p.keys[key.key] = []*Key{}
        }
        p.keys[key.key] = append(p.keys[key.key], key)
    } else {
        p.addToInventory(item)
    }
    item.SetHolder(p)
}

func (p *Party) onItemEquipStatusChanged(items []Item) {
    for _, item := range items {
        p.RemoveItem(item)
    }
    for _, item := range items {
        p.AddItem(item)
    }
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

func (p *Party) RemoveItem(item Item) bool {
    if key, ok := item.(*Key); ok {
        delete(p.keys, key.key)
        item.SetHolder(nil)
        return true
    } else {
        for i, it := range p.partyInventory {
            if it[0] == item {
                p.partyInventory[i] = append(it[:0], it[1:]...)
                if len(p.partyInventory[i]) == 0 {
                    p.partyInventory = append(p.partyInventory[:i], p.partyInventory[i+1:]...)
                }
                item.SetHolder(nil)
                return true
            } // this looks weird..
            if it[0].CanStackWith(item) {
                for j, stackItem := range it {
                    if stackItem == item {
                        p.partyInventory[i] = append(it[:j], it[j+1:]...)
                        item.SetHolder(nil)
                        return true
                    }
                }
            }
        }
    }
    return false
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

func (p *Party) GetInventory() [][]Item {
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
        npc.OnAddedToParty(p)
    }
}

func (p *Party) Move(relativeMovement geometry.Point) {
    if p.splitControlled != nil {
        p.gridMap.MoveActor(p.splitControlled, p.splitControlled.Pos().Add(relativeMovement))
    } else if p.IsInVehicle() {
        p.vehicleMovement(relativeMovement)
    } else {
        p.walkMovement(relativeMovement)
    }
}

func (p *Party) walkMovement(relativeMovement geometry.Point) {
    p.updatePreviousPositions()
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
func (p *Party) resetToPreviousPositions() {
    for i, pos := range p.prevPositions {
        p.gridMap.MoveActor(p.members[i], pos)
    }
}
func (p *Party) updatePreviousPositions() {
    p.prevPositions = make([]geometry.Point, len(p.members))
    for i, member := range p.members {
        p.prevPositions[i] = member.Pos()
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

func (p *Party) GetSplitActions(g Engine) []util.MenuItem {
    var items []util.MenuItem
    if p.HasFollowers() {
        for _, m := range p.members {
            member := m
            items = append(items, util.MenuItem{
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

func (p *Party) TryRest(engine Engine) bool {
    if p.food < len(p.members) {
        return false
    }
    p.food -= len(p.members)
    for _, member := range p.members {
        member.FullRest(engine)
    }
    return true
}

func (p *Party) NeedsRest() bool {
    for _, member := range p.members {
        if member.health < member.maxHealth || member.HasStatusEffects() {
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

func (p *Party) GetSpells() []*Action {
    var result []*Action
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
            m.OnRemovedFromParty()
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
    for _, keys := range p.keys {
        for _, key := range keys {
            result = append(result, key)
        }
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

func (p *Party) NeedsRestAfterMovement() bool {
    p.stepsBeforeRest--
    if p.stepsBeforeRest <= 0 {
        p.stepsBeforeRest = p.rules.GetStepsBeforeRest()
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

func (p *Party) GetMemberIcon(wearer ItemWearer) int32 {
    wearerIcons := []rune{'Ⅰ', 'Ⅱ', 'Ⅲ', 'Ⅳ'}
    return wearerIcons[p.GetMemberIndex(wearer)]
}

func (p *Party) GetFilteredInventory(keep func(item Item) bool) []Item {
    var result []Item
    for _, itemStack := range p.partyInventory {
        firstItemOfStack := itemStack[0]
        if keep(firstItemOfStack) {
            result = append(result, firstItemOfStack)
        }
    }
    return result
}

func (p *Party) GetFilteredStackedInventory(keep func(item Item) bool) [][]Item {
    var result [][]Item
    for _, itemStack := range p.partyInventory {
        firstItemOfStack := itemStack[0]
        if keep(firstItemOfStack) {
            result = append(result, itemStack)
        }
    }
    return result
}

func (p *Party) HasItems() bool {
    return len(p.partyInventory) > 0
}

func (p *Party) EnterVehicle(v *Vehicle) {
    // remove everyone from the map, TODO: nice animation
    for _, member := range p.members {
        p.gridMap.RemoveActor(member)
    }
    p.currentVehicle = v
}

func (p *Party) IsInVehicle() bool {
    return p.currentVehicle != nil
}

func (p *Party) vehicleMovement(movement geometry.Point) {
    if p.currentVehicle == nil {
        return
    }
    destPosition := p.currentVehicle.Pos().Add(movement)
    if !p.gridMap.Contains(destPosition) {
        return
    }
    destCell := p.gridMap.GetCell(destPosition)
    if p.currentVehicle.CanMoveTo(destCell) {
        p.gridMap.MoveObject(p.currentVehicle, destPosition)
        for _, member := range p.members {
            member.SetPos(p.currentVehicle.Pos())
        }
    }
}

func (p *Party) TryExitVehicle() bool {
    if p.currentVehicle == nil {
        return false
    }
    freeNeighbor := p.gridMap.GetRandomFreeNeighbor(p.currentVehicle.Pos())
    if freeNeighbor == p.currentVehicle.Pos() {
        return false
    }
    p.spawnPartyAt(freeNeighbor)
    p.currentVehicle = nil
    return true
}

func (p *Party) PlaceOnMap(currentMap *gridmap.GridMap[*Actor, Item, Object], spawnPos geometry.Point) {
    p.SetGridMap(currentMap)
    p.spawnPartyAt(spawnPos)
}

func (p *Party) spawnPartyAt(spawnPos geometry.Point) {
    p.gridMap.AddActor(p.members[0], spawnPos)

    if p.HasFollowers() {
        followerCount := len(p.GetMembers()) - 1
        freeCells := p.gridMap.GetFreeCellsForDistribution(spawnPos, followerCount, func(pos geometry.Point) bool {
            return p.gridMap.Contains(pos) && p.gridMap.IsCurrentlyPassable(pos)
        })
        if len(freeCells) < followerCount {
            println(fmt.Sprintf("ERROR: not enough free cells for followers at %v", spawnPos))
        } else {
            for i, follower := range p.GetMembers()[1:] {
                followerPos := freeCells[i]
                p.gridMap.AddActor(follower, followerPos)
            }
        }
    }
}

func (p *Party) GetControlledActor() *Actor {
    if p.splitControlled != nil {
        return p.splitControlled
    }
    return p.members[0]
}

func (p *Party) IsSplit() bool {
    return p.splitControlled != nil
}

func (p *Party) SwitchControlTo(actor *Actor) {
    p.splitControlled = actor
}

func (p *Party) ReturnControlToLeader() {
    p.splitControlled = nil
}

func (p *Party) GetMinutesPerStepOnWorldmap() int {
    if p.IsInVehicle() {
        return p.currentVehicle.GetMinutesPerStep()
    }
    return p.rules.GetMinutesPerStepOnWorldmap()
}

func (p *Party) HasAnyRangedWeaponsEquipped() bool {
    for _, member := range p.members {
        if member.HasRangedWeaponEquipped() {
            return true
        }
    }
    return false
}

func (p *Party) UsedKey(key string) {
    p.usedKeys[key] = true
}
func (p *Party) HasUsedKey(key string) bool {
    _, exists := p.usedKeys[key]
    return exists
}

func (p *Party) CanSee(pos geometry.Point) bool {
    return p.fov.Visible(pos)
}

func (p *Party) AddSpellEffect(effect OngoingSpellEffect, duration int) {
    p.activeSpellEffects[effect] = duration
}

func (p *Party) HasSpellEffect(effect OngoingSpellEffect) bool {
    _, exists := p.activeSpellEffects[effect]
    return exists
}

func (p *Party) HasTool(toolType ToolType) bool {
    for _, itemStack := range p.partyInventory {
        firstItemOfStack := itemStack[0]
        if tool, ok := firstItemOfStack.(*Tool); ok && tool.kind == toolType {
            return true
        }
    }
    return false
}

func (p *Party) RemoveTool(toolType ToolType) {
    for _, itemStack := range p.partyInventory {
        firstItemOfStack := itemStack[0]
        if tool, ok := firstItemOfStack.(*Tool); ok && tool.kind == toolType {
            p.RemoveItem(tool)
            return
        }
    }
}

func (p *Party) RemoveAllItems() [][]Item {
    result := p.partyInventory
    for _, itemStack := range p.partyInventory {
        if _, ok := itemStack[0].(Equippable); ok {
            for _, item := range itemStack {
                equippableItem := item.(Equippable)
                equippableItem.Unequip()
            }
        }
    }
    p.partyInventory = [][]Item{}
    return result
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
