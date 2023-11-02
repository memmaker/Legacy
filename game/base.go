package game

import "Legacy/geometry"

type GameObject struct {
    pos              geometry.Point
    isHidden         bool
    discoveryMessage []string
}

func (a *GameObject) Pos() geometry.Point {
    return a.pos
}

func (a *GameObject) SetPos(pos geometry.Point) {
    a.pos = pos
}

func (a *GameObject) IsHidden() bool {
    return a.isHidden
}

func (a *GameObject) SetDiscoveryMessage(hidden bool, message []string) {
    a.isHidden = hidden
    a.discoveryMessage = message
}

func (a *GameObject) SetHidden(hidden bool) {
    a.isHidden = hidden
}

func (a *GameObject) Discover() []string {
    a.isHidden = false
    return a.discoveryMessage
}
