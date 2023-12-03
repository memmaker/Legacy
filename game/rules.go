package game

import (
    "Legacy/util"
    "fmt"
    "math"
    "math/rand"
    "strings"
)

type Rules struct {
    mirrorMap       map[string]TeleportTarget
    difficultyTable map[DifficultyLevel]float64
}
type TeleportTarget struct {
    MapName  string
    Location string
}

func NewRules() *Rules {
    return &Rules{
        difficultyTable: map[DifficultyLevel]float64{
            DifficultyLevelTrivial:        0.99,
            DifficultyLevelVeryEasy:       0.8,
            DifficultyLevelEasy:           0.65,
            DifficultyLevelMedium:         0.5,
            DifficultyLevelHard:           0.35,
            DifficultyLevelVeryHard:       0.2,
            DifficultyLevelNearImpossible: 0.1,
            DifficultyLevelImpossible:     0.01,
        },

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

func (r *Rules) RollSkillCheck(skillLevel SkillLevel, difficultyLevel DifficultyLevel) bool {
    relativeDiff := r.GetRelativeDifficulty(skillLevel, difficultyLevel)
    chance := r.difficultyTable[relativeDiff]
    return RollChance(chance)
}

func RollChance(chance float64) bool {
    return rand.Float64() < chance
}

func (r *Rules) GetBaseValueOfLockpick() int {
    return 50
}

func (r *Rules) GetBaseValueOfFood() int {
    return 100
}

func (r *Rules) NeededXpForLevel(level int) int {
    //403.4093
    //⋅
    //1.3069
    //LEVEL
    x := 0.035
    y := 1.8
    return int(math.Pow(float64(level)/x, y))
    //return int(7892.43*math.Pow(math.E, 0.0503654*float64(level)) - 7800.12)
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
    totalXP := 0
    xpForLastLevel := 0
    for i := from; i <= to; i++ {
        xpForLevel := r.NeededXpForLevel(i)
        diff := xpForLevel - xpForLastLevel
        row := util.TableRow{
            Label: fmt.Sprintf("%d", i),
            Columns: []string{
                fmt.Sprintf("%d (+%d)", xpForLevel, diff),
            },
        }
        rows = append(rows, row)
        xpForLastLevel = xpForLevel
        totalXP += xpForLevel
    }
    rows = append(rows, util.TableRow{
        Label: "---",
        Columns: []string{
            "------",
        },
    })
    rows = append(rows, util.TableRow{
        Label: "Sum",
        Columns: []string{
            fmt.Sprintf("%d", totalXP),
        },
    })
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
    a.attributes.Increment(Level)
    healthBonus := a.GetAttributes().GetAttribute(Endurance)
    a.maxHealth += healthBonus
    a.health = a.maxHealth
}

func (r *Rules) GetMeleeDamage(attacker *Actor, victim *Actor) int {
    baseMeleeDamage := attacker.GetMeleeDamage()
    victimArmor := victim.GetTotalArmor()
    meleeDamage := max(baseMeleeDamage-victimArmor, 0)
    if meleeDamage == 0 && baseMeleeDamage > 0 {
        meleeDamage = 1
    }
    return meleeDamage
}

func (r *Rules) GetRangedDamage(attacker *Actor, victim *Actor) int {
    baseRangedDamage := attacker.GetRangedDamage()
    victimArmor := victim.GetTotalArmor()
    rangedDamage := max(0, baseRangedDamage-victimArmor)
    if rangedDamage == 0 && baseRangedDamage > 0 {
        rangedDamage = 1
    }
    return rangedDamage
}

func (r *Rules) DoesMeleeAttackHit(attacker *Actor, defender *Actor) bool {
    chance := r.GetMeleeHitChance(attacker, defender)
    return RollChance(chance)
}

func (r *Rules) DoesRangedAttackHit(attacker *Actor, defender *Actor) bool {
    chance := r.GetRangedHitChance(attacker, defender)
    return RollChance(chance)
}

func (r *Rules) GetMeleeHitChance(attacker *Actor, defender *Actor) float64 {
    skillActor := attacker.GetSkills().GetSkillLevel(PhysicalSkillMeleeCombat)
    skillAdversary := defender.GetSkills().GetSkillLevel(PhysicalSkillMeleeCombat)
    relativeDifficulty := DifficultyLevelFromInt(max(-1, min(7, (int(skillAdversary)-int(skillActor))-1)))
    chance := r.difficultyTable[relativeDifficulty]
    return chance
}

func (r *Rules) GetRangedHitChance(attacker *Actor, defender *Actor) float64 {
    skillActor := attacker.GetSkills().GetLevel(PhysicalSkillRangedCombat)
    absoluteDiff := max(defender.GetAbsoluteDifficultyByAttribute(Agility), defender.GetAbsoluteDifficultyByAttribute(Perception))
    relativeDiff := r.GetRelativeDifficulty(skillActor, absoluteDiff).ReducedBy(1)
    chance := r.difficultyTable[relativeDiff]
    return chance
}

func (r *Rules) GetMinutesPerStepInLevels() int {
    return 1
}

func (r *Rules) GetMinutesPerStepOnWorldmap() int {
    return 20
}

func (r *Rules) GetRelativeDifficulty(skill SkillLevel, difficulty DifficultyLevel) DifficultyLevel {
    diffAsInt := int(difficulty)
    skillAsInt := int(skill - 1)
    relativeDifficulty := diffAsInt - skillAsInt
    fromInt := DifficultyLevelFromInt(relativeDifficulty)
    return fromInt
}
func (r *Rules) GetVersusDifficulty(skillActor, skillAdversary SkillLevel) DifficultyLevel {
    //0..4 - 0..4 = -4..4 + 2 = -2..6
    relativeDifficulty := (int(skillAdversary) - int(skillActor)) + 2
    fromInt := DifficultyLevelFromInt(relativeDifficulty)
    return fromInt
}
func (r *Rules) GetSkillCheckTable() []string {
    var rows []util.TableRow
    for i := 1; i <= 4; i++ {
        skillLevel := SkillLevel(i)
        for j := 1; j <= 4; j++ {
            difficultyLevel := DifficultyLevelFromInt(j)
            relativeDifficulty := r.GetRelativeDifficulty(skillLevel, difficultyLevel)
            successes := 0
            for k := 1; k <= 1000; k++ {
                if r.RollSkillCheck(skillLevel, difficultyLevel) {
                    successes++
                }
            }
            successRate := int((float64(successes) / 1000.0) * 100.0)

            row := util.TableRow{
                Label:   skillLevel.ToString(),
                Columns: []string{difficultyLevel.ToString(), relativeDifficulty.ToString(), fmt.Sprintf("%d%%", successRate)},
            }
            rows = append(rows, row)
        }

    }
    layout := util.TableLayout(rows)
    return layout
}

func (r *Rules) GetCriminalOffenseEvent(nameOfCurrentMap string) string {
    return "tauci_criminal_offense"
}
