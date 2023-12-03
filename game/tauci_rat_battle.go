package game

type TauciRatBattleStartEvent struct {
    engine Engine
    isOver bool
}

func (t *TauciRatBattleStartEvent) Init() {
    engine := t.engine
    //party := engine.GetParty()

    // find all the rats on the current level
    var rats []*Actor
    prisonMap := engine.GetGridMap()
    for _, actor := range prisonMap.Actors() {
        if actor.GetInternalName() == "grey_rat" {
            rats = append(rats, actor)
            prisonMap.RemoveActor(actor)
        }
    }
    // transition to the castle map
    engine.TransitionToNamedLocation("Tauci_Castle", "Throne_Room")
    //TODO
    //castleMap := engine.GetGridMap()

    // place 4 rats into the throne room

    // place all the other rats randomly on the map

    // set up two factions: party and rats vs. tauci

}

func (t *TauciRatBattleStartEvent) IsOver() bool {
    return t.isOver
}

func (t *TauciRatBattleStartEvent) Update() {
    // check for the end of the battle
    // possible outcomes
    // tauci are all dead
    // rats are all dead - party can surrender or continue
    // party is all dead - game over
    // flee - continue on return to this map?
}
