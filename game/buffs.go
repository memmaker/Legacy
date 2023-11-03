package game

type BuffType string

const (
    BuffTypeDefense BuffType = "defense"
    BuffTypeOffense BuffType = "offense"
)

type Buff struct {
    Name     string
    Strength int
}
