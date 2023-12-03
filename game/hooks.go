package game

import (
    "Legacy/geometry"
    "Legacy/util"
)

type MovementHook struct {
    Applies func(mover *Actor, pos geometry.Point) bool
    Action  func(mover *Actor, pos geometry.Point)
    Consume bool
}

type ActorActionHook struct {
    Applies func(actor *Actor) bool
    Action  func(actor *Actor) []util.MenuItem
}
type PlantHook struct {
    Applies func(item Item, owner *Actor) bool
    Action  func(item Item, owner *Actor)
    Consume bool
}

type LevelHooks struct {
    MovementHooks    []MovementHook
    ActorActionHooks []ActorActionHook
    PlantHooks       []PlantHook
}

// GetNPCMovementRoutine hook: when a slime walks over the powder it dies
// then allow throwing the powder
// context action hook: when the slime is sleeping, allow taking it as inventory item
func GetHooksForLevel(engine Engine, level string) LevelHooks {
    switch level {
    case "Tauci_Castle":
        return tauciCastleHooks(engine)
    }
    return LevelHooks{}
}

func tauciCastleHooks(engine Engine) LevelHooks {
    allHooks := LevelHooks{}
    slimePlantHook := PlantHook{
        Applies: func(item Item, owner *Actor) bool {
            return item.Name() == "yellow powder" && owner.internalName == "slime"
        },
        Action: func(item Item, owner *Actor) {
            engine.Kill(owner)
        },
        Consume: false,
    }
    allHooks.PlantHooks = append(allHooks.PlantHooks, slimePlantHook)

    slimeMovesHook := MovementHook{
        Applies: func(mover *Actor, pos geometry.Point) bool {
            currentMap := engine.GetGridMap()
            if currentMap.IsItemAt(pos) {
                item := currentMap.ItemAt(pos)
                return item.Name() == "yellow powder" && mover.internalName == "slime"
            }
            return false
        },
        Action: func(mover *Actor, pos geometry.Point) {
            mover.Damage(engine, mover.GetHealth()) // happens during combat
        },
    }
    allHooks.MovementHooks = append(allHooks.MovementHooks, slimeMovesHook)

    takeSlimeHook := ActorActionHook{
        Applies: func(actor *Actor) bool {
            return actor.internalName == "slime" && actor.IsSleeping()
        },
        Action: func(actorAt *Actor) []util.MenuItem {
            return []util.MenuItem{
                {
                    Text: "Take",
                    Action: func() {
                        currentMap := engine.GetGridMap()
                        if actorAt.internalName != "slime" || !actorAt.IsSleeping() {
                            return
                        }
                        currentMap.RemoveActor(actorAt)
                        slimeItem := NewFlavorItem("paralyzed slime", 1)
                        slimeItem.SetDescription([]string{"A slime that has been paralyzed.", "It is still alive,", "but it can't move."})
                        engine.TakeItem(slimeItem)
                    },
                },
            }
        },
    }
    allHooks.ActorActionHooks = append(allHooks.ActorActionHooks, takeSlimeHook)

    return allHooks
}
