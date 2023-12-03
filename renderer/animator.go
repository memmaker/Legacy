package renderer

import (
    "Legacy/geometry"
    "Legacy/gocoro"
    "Legacy/util"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
)

type Animator struct {
    hitAnimations            []*TileAnimation
    animationRoutine         gocoro.Coroutine
    animationsRunning        bool
    mapToOnScreenCoordinates func(geometry.Point) (bool, geometry.Point)
}

func NewAnimator(mapToOnScreenCoordinates func(geometry.Point) (bool, geometry.Point)) *Animator {
    return &Animator{
        hitAnimations:            []*TileAnimation{},
        animationRoutine:         gocoro.NewCoroutine(),
        mapToOnScreenCoordinates: mapToOnScreenCoordinates,
    }
}

func (a *Animator) Update() {
    a.animationsRunning = false
    // handle player input
    if a.animationRoutine.Running() {
        a.animationRoutine.Update()
        a.animationsRunning = true
    }

    for i := len(a.hitAnimations) - 1; i >= 0; i-- {
        hitAnim := a.hitAnimations[i]
        hitAnim.TicksAlive++
        if hitAnim.IsFinished() {
            a.hitAnimations = append(a.hitAnimations[:i], a.hitAnimations[i+1:]...)
            if hitAnim.WhenDone != nil {
                hitAnim.WhenDone()
            }
        }
        a.animationsRunning = true
    }
}

func (a *Animator) Draw(gridRenderer *DualGridRenderer, screen *ebiten.Image) {
    // draw the combat UI
    offset := gridRenderer.GetScaledSmallGridSize()
    offsetPoint := geometry.Point{X: offset, Y: offset}
    for _, hitAnim := range a.hitAnimations {
        mapPosition := hitAnim.CurrentPosition()
        if isOnScreen, screenPos := a.mapToOnScreenCoordinates(mapPosition); isOnScreen {
            gridRenderer.DrawOnBigGridWithColor(screen, screenPos, offsetPoint, hitAnim.UseTiles, hitAnim.Icon(), hitAnim.TintColor)
        }
    }
}
func (a *Animator) AddDefaultHitAnimation(pos geometry.Point, atlas AtlasName, icon int32, tintColor color.Color, done func()) {
    animation := &TileAnimation{
        Positions: []geometry.Point{pos},
        Frames:    []int32{icon},
        WhenDone:  done,
        UseTiles:  atlas,
        TintColor: tintColor,
    }
    animation.EndAfterTime(0.33)
    a.hitAnimations = append(a.hitAnimations, animation)
}

func (a *Animator) AddHitAnimation(projAnim *TileAnimation) {
    a.hitAnimations = append(a.hitAnimations, projAnim)
}

func (a *Animator) IsRunning() bool {
    return a.animationsRunning
}

func (a *Animator) RunAnimationScript(script func(exe *gocoro.Execution)) {
    err := a.animationRoutine.Run(script)
    if err != nil {
        println(err.Error())
    }
}

type TileAnimation struct {
    Positions []geometry.Point

    TicksAlive                 int
    MoveIntervalInSeconds      float64
    AnimationIntervalInSeconds float64

    Frames    []int32
    UseTiles  AtlasName
    WhenDone  func()
    IsDone    func() bool
    TintColor color.Color
}

func (h *TileAnimation) CurrentPathIndex() int {
    if h.MoveIntervalInSeconds == 0 {
        return 0
    }
    interval := h.getCurrentInterval()
    return min(interval, len(h.Positions)-1)
}

func (h *TileAnimation) getCurrentInterval() int {
    tps := max(60, ebiten.ActualTPS())
    interval := h.TicksAlive / int(h.MoveIntervalInSeconds*tps)
    return interval
}
func (h *TileAnimation) CurrentPosition() geometry.Point {
    return h.Positions[h.CurrentPathIndex()]
}

func (h *TileAnimation) AnimateMovement() bool {
    return len(h.Positions) > 1 && h.MoveIntervalInSeconds != 0
}

func (h *TileAnimation) AnimateIcon() bool {
    return len(h.Frames) > 1 && h.AnimationIntervalInSeconds != 0
}

func (h *TileAnimation) IsFinished() bool {
    return h.IsDone()
}

func (h *TileAnimation) Icon() int32 {
    if len(h.Frames) == 0 {
        return 0
    }
    if !h.AnimateIcon() {
        return h.Frames[0]
    }

    frameOffset := util.GetLoopingFrameFromTick(uint64(h.TicksAlive), h.AnimationIntervalInSeconds, len(h.Frames))
    return h.Frames[frameOffset]
}

func (h *TileAnimation) EndAfterTime(timeInSeconds float64) {
    h.IsDone = func() bool {
        tps := max(60, ebiten.ActualTPS())
        return float64(h.TicksAlive)/tps >= timeInSeconds
    }
}

func (h *TileAnimation) EndAfterPathComplete() {
    h.IsDone = func() bool {
        return h.getCurrentInterval() == len(h.Positions)
    }
}
