package game

import "fmt"

const DaysPerWeek = 7
const HoursPerDay = 24
const DaysPerYear = 360
const MinutesPerHour = 60
const MinutesPerDay = MinutesPerHour * HoursPerDay

type WorldTime struct {
    minutes int
    days    int
}

func (w WorldTime) HoursAndMinutes() (int, int) {
    return w.minutes / MinutesPerHour, w.minutes % MinutesPerHour
}

func (w WorldTime) YearsAndDays() (int, int) {
    return w.days / DaysPerYear, w.days % DaysPerYear
}

func (w WorldTime) WithAddedMinutes(minutes int) WorldTime {
    w.minutes += minutes
    if w.minutes >= MinutesPerDay {
        w.minutes -= MinutesPerDay
        w.days++
    }
    return w
}
func (w WorldTime) WithAddedDays(days int) WorldTime {
    w.days += days
    return w
}

func (w WorldTime) GetDate() string {
    year, day := w.YearsAndDays()
    return fmt.Sprintf("Year %d, Day %d", year, day)
}

func (w WorldTime) GetTime() string {
    hours, minutes := w.HoursAndMinutes()
    return fmt.Sprintf("%02d:%02d", hours, minutes)
}

func (w WorldTime) GetTimeAndDate() string {
    return fmt.Sprintf("%s, %s", w.GetTime(), w.GetDate())
}
func NewWorldTime() WorldTime {
    return WorldTime{
        minutes: 0,
        days:    10 * DaysPerYear,
    }
}
