package game

import (
    "Legacy/util"
    "fmt"
    "math"
    "strings"
)

type Rules struct {
    mirrorMap map[string]TeleportTarget
}
type TeleportTarget struct {
    MapName  string
    Location string
}

func NewRules() *Rules {
    return &Rules{
        mirrorMap: map[string]TeleportTarget{
            "celador": {
                MapName:  "WorldMap",
                Location: "worldmap_spawn",
            },
            "tauci king": {
                MapName:  "Tauci_Castle",
                Location: "Throne_Room",
            },
            "tauci mines": {
                MapName:  "Tauci_Mines_Level_1",
                Location: "ladder_up",
            },
            "tauci woods": {
                MapName:  "Tauci_Woods",
                Location: "entrance",
            },
        },
    }
}

func (r *Rules) GetBaseValueOfLockpick() int {
    return 50
}

func (r *Rules) GetBaseValueOfFood() int {
    return 100
}

func (r *Rules) NeededXpForLevel(level int) int {
    return int(7892.43*math.Pow(math.E, 0.0503654*float64(level)) - 7800.12)
}

func (r *Rules) CanLevelUp(level, xp int) (bool, int) {
    xpForNext := r.NeededXpForLevel(level + 1)
    canLevel := xp >= xpForNext
    xpMissing := max(0, xpForNext-xp)
    return canLevel, xpMissing
}

func (r *Rules) GetTrainerCost(level int) int {
    return (level * level) * 10
}

func (r *Rules) GetXPTable(from, to int) []string {
    var rows []util.TableRow
    for i := from; i <= to; i++ {
        row := util.TableRow{
            Label: fmt.Sprintf("%d", i),
            Columns: []string{
                fmt.Sprintf("%d", r.NeededXpForLevel(i)),
            },
        }
        rows = append(rows, row)
    }
    return util.TableLayout(rows)
}

func (r *Rules) GetTargetTravelLocation(text string) (string, string) {
    if text == "" {
        return "", ""
    }
    text = strings.ToLower(text)
    var mapName, locationName string
    if target, ok := r.mirrorMap[text]; ok {
        mapName = target.MapName
        locationName = target.Location
    }
    return mapName, locationName
}

func (r *Rules) GetStepsBeforeRest() int {
    return 1000
}

func (r *Rules) GetPartyStartGold() int {
    return 1000
}

func (r *Rules) GetPartyStartFood() int {
    return 10
}

func (r *Rules) LevelUp(a *Actor) {
    a.level++
    healthBonus := 10
    a.maxHealth += healthBonus
    a.health = a.maxHealth
    a.baseArmor += 1
    a.baseMeleeDamage += 1
    a.baseRangedDamage += 1
}

func (r *Rules) GetMeleeDamage(attacker *Actor, victim *Actor) int {
    baseMeleeDamage := attacker.GetMeleeDamage()
    victimArmor := victim.GetTotalArmor()
    meleeDamage := baseMeleeDamage - victimArmor
    return meleeDamage
}

func (r *Rules) GetRangedDamage(attacker *Actor, victim *Actor) int {
    baseRangedDamage := attacker.GetRangedDamage()
    victimArmor := victim.GetTotalArmor()
    rangedDamage := baseRangedDamage - victimArmor
    return rangedDamage
}

func (r *Rules) CalculateMeleeHit(attacker *Actor, npc *Actor) bool {
    return true
}

func (r *Rules) GetMinutesPerStepInLevels() int {
    return 1
}

func (r *Rules) GetMinutesPerStepOnWorldmap() int {
    return 20
}
