package game

import "fmt"

type Flags struct {
    flags map[string]int
}

func NewFlags() *Flags {
    return &Flags{
        flags: make(map[string]int),
    }
}

func (f *Flags) SetFlag(key string, value int) {
    f.flags[key] = value
}

func (f *Flags) GetFlag(key string) int {
    if val, ok := f.flags[key]; ok {
        return val
    }
    return 0
}

func (f *Flags) HasFlag(key string) bool {
    val, ok := f.flags[key]
    return ok && val > 0
}

func (f *Flags) IncrementFlag(key string) {
    f.IncrementFlagBy(key, 1)
}

func (f *Flags) IncrementFlagBy(key string, amount int) {
    f.flags[key] = f.GetFlag(key) + amount
}

func (f *Flags) AllSet(flags []string) bool {
    for _, flag := range flags {
        if !f.HasFlag(flag) {
            return false
        }
    }
    return true
}

func (f *Flags) GetDebugInfo() []string {
    result := make([]string, 0)
    for key, val := range f.flags {
        result = append(result, fmt.Sprintf("%s: %d", key, val))
    }
    return result
}
