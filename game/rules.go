package game

import (
    "Legacy/util"
    "fmt"
    "math"
)

type Rules struct {
}

func NewRules() *Rules {
    return &Rules{}
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
    xpMissing := min(0, xpForNext-xp)
    return canLevel, xpMissing
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
