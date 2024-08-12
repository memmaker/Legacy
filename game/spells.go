package game

import (
    "Legacy/ega"
    "Legacy/geometry"
    "Legacy/gridmap"
    "Legacy/renderer"
    "fmt"
    "image/color"
    "math/rand"
)

type OngoingSpellEffect string

type Spell struct {
    BaseAction
    scrollTitle   string
    scrollFile    string
    labelWithCost string
}

func (s *Spell) LabelWithCost() string {
    return s.labelWithCost
}

func (s *Spell) SetMonetaryValue(value int) {
    s.monetaryValue = value
}

const (
    OngoingSpellEffectBirdsEye OngoingSpellEffect = "birds_eye"
)

func GetAllSpellScrolls() []*Scroll {
    names := GetAllSpellNames()
    var scrolls []*Scroll
    for _, name := range names {
        scrolls = append(scrolls, NewSpellScrollFromSpellName(name))
    }

    return scrolls
}

func GetAllSpellNames() []string {
    names := []string{
        "Nom De Plume",
        "Create Food",
        "Bird's Eye",
        "Raise as Undead",
        "Fireball",
        "Icebolt",
        "Healing word of Tauci",
    }
    return names
}

func GetSpellNamesByLevel(level int) []string {
    switch level {
    case 1:
        return []string{
            "Nom De Plume",
            "Create Food",
        }
    case 2:
        return []string{
            "Bird's Eye",
            "Healing word of Tauci",
        }
    case 5:
        return []string{
            "Icebolt",
        }
    case 6:
        return []string{
            "Fireball",
        }
    case 7:
        return []string{
            "Raise as Undead",
        }
    }
    return []string{}
}

func NewRandomScrollForVendor(level int) *Scroll {
    spellNames := GetSpellNamesByLevel(level)
    randomIndex := rand.Intn(len(spellNames))
    spellName := spellNames[randomIndex]
    return NewSpellScrollFromSpellName(spellName)
}

func NewSpellFromName(name string) *Spell {
    switch name {
    case "Nom De Plume":
        spell := NewSpell(name, 0, func(engine Engine, caster *Actor) {
            engine.AskUserForString("Now known as: ", 8, func(text string) {
                caster.SetName(text)
            })
        })
        spell.SetDescription([]string{
            "Allows you to change your name.",
        })
        spell.SetScrollTitle("Name of a rose")
        spell.SetScrollFile("nom_de_plume")
        spell.SetNoCombatUtility()
        spell.SetMonetaryValue(200)
        return spell
    case "Create Food":
        spell := NewSpell(name, 10, func(engine Engine, caster *Actor) {
            engine.AddFood(10)
        })
        spell.SetDescription([]string{
            "Create 10 rations of food",
            "out of thin air.",
        })
        spell.SetScrollTitle("A poor man's feast")
        spell.SetScrollFile("create_food")
        spell.SetNoCombatUtility()
        spell.SetMonetaryValue(3000)
        return spell
    case "Healing word of Tauci":
        spell := NewSpell(name, 10, func(engine Engine, caster *Actor) {
            healthIncrease := 10 * caster.GetLevel()
            newHealth := min(caster.GetHealth()+healthIncrease, caster.GetMaxHealth())
            caster.SetHealth(newHealth)
            engine.Print("You feel better!")
        })
        spell.SetDescription([]string{
            "Heal yourself.",
            "Heals: level of caster x 10 HP",
        })
        spell.SetScrollTitle("On health & wealth")
        spell.SetScrollFile("healing_word_of_tauci")
        spell.SetCombatUtilityForUseAtLocation(func(engine Engine, caster *Actor, userPosition geometry.Point, allyPositions, enemyPositions map[geometry.Point]*Actor) int {
            if caster.GetHealth() < caster.GetMaxHealth() {
                return caster.GetMaxHealth() - caster.GetHealth()
            } else {
                return -50
            }
        })
        spell.SetMonetaryValue(5000)
        return spell
    case "Bird's Eye":
        spell := NewSpell(name, 10, func(engine Engine, caster *Actor) {
            engine.GetParty().AddSpellEffect(OngoingSpellEffectBirdsEye, 5)
        })
        spell.SetDescription([]string{
            "You see the world from a bird's eye view.",
        })
        spell.SetScrollTitle("Change of perspective")
        spell.SetScrollFile("birds_eye")
        spell.SetNoCombatUtility()
        spell.SetMonetaryValue(7000)
        return spell
    case "Raise as Undead":
        targetedSpell := NewTargetedSpell(name, 10, func(engine Engine, caster *Actor, pos geometry.Point) {
            engine.RaiseAsUndeadAt(caster, pos)
        })
        targetedSpell.SetActionColor(ega.BrightBlack)
        deadActorsInRange := allVisibleTilesInRadiusWith(4, func(engine Engine, pos geometry.Point) bool {
            gridMap := engine.GetGridMap()
            if !gridMap.IsDownedActorAt(pos) {
                return false
            }
            downedActorAt := gridMap.DownedActorAt(pos)
            return downedActorAt != nil && !downedActorAt.IsAlive()
        })
        targetedSpell.SetValidTargets(deadActorsInRange)
        targetedSpell.SetScrollTitle("The dead shall rise")
        targetedSpell.SetScrollFile("raise_as_undead")
        targetedSpell.SetDescription([]string{
            "Raise a dead person as undead.",
            "Range: 4 tiles.",
        })
        targetedSpell.SetCombatUtilityForTargetedUseOnLocation(func(engine Engine, caster *Actor, target geometry.Point, allyPositions, enemyPositions map[geometry.Point]*Actor) int {
            if engine.IsPlayerControlled(caster) {
                party := engine.GetParty()
                if len(party.GetMembers()) == 4 {
                    return -10
                }
            }

            gridMap := engine.GetGridMap()
            if !gridMap.IsDownedActorAt(target) {
                return 0
            }
            downedActorAt := gridMap.DownedActorAt(target)
            if downedActorAt != nil && !downedActorAt.IsAlive() {
                return 100
            }
            return 0
        })
        targetedSpell.SetMonetaryValue(25000)
        return targetedSpell
    case "Fireball":
        radius := 3
        fireball := NewTargetedSpell(name, 10, func(engine Engine, caster *Actor, pos geometry.Point) {
            explosionIcon := int32(28)
            fireballDamagePerTile := 10 * caster.GetLevel()
            currentMap := engine.GetGridMap()
            hitPositions := currentMap.GetDijkstraMapWithActorsNotBlocking(pos, radius)
            for p, _ := range hitPositions {
                hitPos := p
                engine.CombatHitAnimation(hitPos, renderer.AtlasEntitiesGrayscale, explosionIcon, ega.BrightRed, func() {
                    engine.FixedDamageAt(caster, hitPos, fireballDamagePerTile)
                })
            }
        })

        fireball.SetValidTargets(allVisibleTilesInRadius(15))
        fireball.SetDescription([]string{
            "Burn everything in a 3 tile radius.",
            "Damage: level of caster x 10 HP",
            "Range: 15 tiles",
        })
        fireball.SetScrollTitle("Fire - the great equalizer")
        fireball.SetScrollFile("fireball")
        fireball.SetActionColor(ega.BrightRed)
        fireball.SetCombatUtilityForTargetedUseOnLocation(func(engine Engine, caster *Actor, target geometry.Point, allyPositions, enemyPositions map[geometry.Point]*Actor) int {
            if target == caster.Pos() {
                return -100
            }
            currentMap := engine.GetGridMap()
            hitPositions := currentMap.GetDijkstraMapWithActorsNotBlocking(target, radius)
            totalUtility := 0
            for p, _ := range hitPositions {
                hitPos := p
                if _, ok := enemyPositions[hitPos]; ok {
                    totalUtility += 10
                }
                if _, ok := allyPositions[hitPos]; ok {
                    totalUtility -= 25
                }
                if hitPos == caster.Pos() {
                    totalUtility -= 50
                }
            }
            return totalUtility
        })
        fireball.SetMonetaryValue(15000)
        return fireball
    case "Icebolt":
        radius := 3
        icebolt := NewTargetedSpell(name, 10, func(engine Engine, caster *Actor, pos geometry.Point) {
            explosionIcon := int32(28)
            iceboltDamage := 1 * caster.GetLevel()
            currentMap := engine.GetGridMap()
            hitPositions := currentMap.GetDijkstraMapWithActorsNotBlocking(pos, radius)
            for p, _ := range hitPositions {
                hitPos := p
                engine.CombatHitAnimation(hitPos, renderer.AtlasEntitiesGrayscale, explosionIcon, ega.BrightBlue, func() {
                    engine.FixedDamageAt(caster, hitPos, iceboltDamage)
                    engine.FreezeActorAt(hitPos, 3)
                })
            }
        })
        icebolt.SetValidTargets(allVisibleTilesInRadius(12))
        icebolt.SetDescription([]string{
            "Freeze everything in a 3 tile radius.",
            "Damage: level of caster x 1 HP",
            "Applies freeze for 3 turns.",
            "Range: 12 tiles",
        })
        icebolt.SetScrollTitle("Frozen in time")
        icebolt.SetScrollFile("icebolt")
        icebolt.SetActionColor(ega.BrightBlue)
        icebolt.SetCombatUtilityForTargetedUseOnLocation(func(engine Engine, caster *Actor, target geometry.Point, allyPositions, enemyPositions map[geometry.Point]*Actor) int {
            currentMap := engine.GetGridMap()
            hitPositions := currentMap.GetDijkstraMapWithActorsNotBlocking(target, radius)
            totalUtility := 0
            for p, _ := range hitPositions {
                hitPos := p
                if _, ok := enemyPositions[hitPos]; ok {
                    totalUtility += 20
                }
                if _, ok := allyPositions[hitPos]; ok {
                    totalUtility -= 20
                }
                if hitPos == caster.Pos() {
                    totalUtility -= 50
                }
            }
            return totalUtility
        })
        icebolt.SetMonetaryValue(13000)
        return icebolt
    }

    return nil
}
func allVisibleTilesInRadius(radius int) func(engine Engine, caster *Actor, usePos geometry.Point) map[geometry.Point]bool {
    return func(engine Engine, caster *Actor, usePos geometry.Point) map[geometry.Point]bool {
        gridMap := engine.GetGridMap()
        potentialPositions := gridmap.GetLocationsInRadius(usePos, float64(radius), func(pos geometry.Point) bool {
            return gridMap.Contains(pos) && engine.GetParty().CanSee(pos)
        })
        validPositions := make(map[geometry.Point]bool)
        for _, pos := range potentialPositions {
            validPositions[pos] = true
        }

        return validPositions
    }
}
func allVisibleTilesInRadiusWith(radius int, keep func(engine Engine, pos geometry.Point) bool) func(engine Engine, caster *Actor, usePos geometry.Point) map[geometry.Point]bool {
    return func(engine Engine, caster *Actor, usePos geometry.Point) map[geometry.Point]bool {
        gridMap := engine.GetGridMap()
        potentialPositions := gridmap.GetLocationsInRadius(usePos, float64(radius), func(pos geometry.Point) bool {
            return gridMap.Contains(pos) && engine.GetParty().CanSee(pos) && keep(engine, pos)
        })
        validPositions := make(map[geometry.Point]bool)
        for _, pos := range potentialPositions {
            validPositions[pos] = true
        }

        return validPositions
    }
}
func NewSpell(name string, manaCost int, effect func(engine Engine, caster *Actor)) *Spell {
    return &Spell{
        BaseAction: BaseAction{
            name:   name,
            effect: effect,
            canPayCost: func(engine Engine, user *Actor) bool {
                if !user.HasMana(manaCost) {
                    engine.Print("Not enough mana!")
                    return false
                }
                return true
            },
            payCost: func(engine Engine, user *Actor) {
                engine.ManaSpent(user, manaCost)
            },
            color: color.White,
        },
        labelWithCost: fmt.Sprintf("%s (%d)", name, manaCost),
    }
}

func NewTargetedSpell(name string, manaCost int, effect func(engine Engine, caster *Actor, pos geometry.Point)) *Spell {
    return &Spell{
        BaseAction: BaseAction{
            name:                 name,
            targetedEffect:       effect,
            closeModalsForEffect: true,
            canPayCost: func(engine Engine, user *Actor) bool {
                if !user.HasMana(manaCost) {
                    engine.Print("Not enough mana!")
                    return false
                }
                return true
            },
            payCost: func(engine Engine, user *Actor) {
                engine.ManaSpent(user, manaCost)
            },
            color: color.White,
        },
        labelWithCost: fmt.Sprintf("%s (%d)", name, manaCost),
    }
}
