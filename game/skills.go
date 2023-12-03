package game

import (
    "Legacy/util"
)

type SkillName string

/*


-> These can be just flags
Art of theft (Sneak, Pickpocket, Lockpick, Steal)
 - More options per level
 - (Sneak, Pickpocket, Lockpick, Glasscutting, etc.)

-> These can be just flags
Social Skills (Deceive, Persuade, Intimidate, Bluff, Spot Lies) (Choose a specific technique per Level)
 - More options per level
 - (Persuasion, Intimidation, Bluff, Spot Lies, etc.)
 -> Do we still roll for failure?

-> Needs a counter
Outdoors Skill (Survival, Hunting, Fishing, Tracking)
 - More catch, more food per level
 - Harder to spot when resting

-> Needs a counter
Athletics Skill (Climbing, Swimming, Jumping, Running)
 - Reduce encumbrance per level
 - Increase speed per level

-> Just flags
Foreign Language Skill
 - Choose a language per Level
 - Animals, Monsters, etc.

-> Flags + Counter
Perception Skill (Spot Hidden, Danger Sense, Listen, Search)
 - More options per level
 - (Spot Hidden, Danger Sense)
 - Further levels increase range
*/

type SkillLevel int

func (l SkillLevel) ToString() string {
    switch l {
    case SkillLevelBeginner:
        return "Beginner"
    case SkillLevelAdvanced:
        return "Advanced"
    case SkillLevelExpert:
        return "Expert"
    case SkillLevelMaster:
        return "Master"
    }
    return "Unknown"
}

const (
    SkillLevelBeginner SkillLevel = 1
    SkillLevelAdvanced SkillLevel = 2
    SkillLevelExpert   SkillLevel = 3
    SkillLevelMaster   SkillLevel = 4
)

type SkillCheck struct {
    SkillName          SkillName
    Difficulty         DifficultyLevel
    IsVersusAntagonist bool
    VersusAttribute    AttributeName
}
type DifficultyLevel int

func (l DifficultyLevel) ReducedBy(level int) DifficultyLevel {
    currentLevel := int(l)
    newLevel := currentLevel - level
    if newLevel < -1 {
        newLevel = -1
    }
    return DifficultyLevel(newLevel)
}

func (l DifficultyLevel) ToString() string {
    switch l {
    case DifficultyLevelTrivial:
        return "Trivial"
    case DifficultyLevelVeryEasy:
        return "Very Easy"
    case DifficultyLevelEasy:
        return "Easy"
    case DifficultyLevelMedium:
        return "Medium"
    case DifficultyLevelHard:
        return "Hard"
    case DifficultyLevelVeryHard:
        return "Very Hard"
    case DifficultyLevelNearImpossible:
        return "Near Impossible"
    case DifficultyLevelImpossible:
        return "Impossible"
    }
    return "Unknown"
}
func DifficultyLevelFromInt(level int) DifficultyLevel {
    if level < -1 {
        return DifficultyLevelTrivial
    }
    if level > 6 {
        return DifficultyLevelImpossible
    }
    return DifficultyLevel(level)
}
func DifficultyLevelFromString(level string) DifficultyLevel {
    switch level {
    case "Trivial":
        return DifficultyLevelTrivial
    case "Very Easy":
        return DifficultyLevelVeryEasy
    case "Easy":
        return DifficultyLevelEasy
    case "Medium":
        return DifficultyLevelMedium
    case "Hard":
        return DifficultyLevelHard
    case "Very Hard":
        return DifficultyLevelVeryHard
    case "Near Impossible":
        return DifficultyLevelNearImpossible
    case "Impossible":
        return DifficultyLevelImpossible
    }
    return DifficultyLevelTrivial
}

const (
    DifficultyLevelTrivial        DifficultyLevel = -1
    DifficultyLevelVeryEasy       DifficultyLevel = 0
    DifficultyLevelEasy           DifficultyLevel = 1
    DifficultyLevelMedium         DifficultyLevel = 2
    DifficultyLevelHard           DifficultyLevel = 3
    DifficultyLevelVeryHard       DifficultyLevel = 4
    DifficultyLevelNearImpossible DifficultyLevel = 5
    DifficultyLevelImpossible     DifficultyLevel = 6
)

type SkillCategory string

const (
    SkillCategoryCombat     SkillCategory = "Melee Combat"
    SkillCategoryRanged     SkillCategory = "Ranged Combat"
    SkillCategoryThieving   SkillCategory = "Art of Theft"
    SkillCategorySocial     SkillCategory = "Social Skills"
    SkillCategoryOutdoors   SkillCategory = "Outdoor Survival"
    SkillCategoryAthletics  SkillCategory = "Athletics"
    SkillCategoryLanguages  SkillCategory = "Languages"
    SkillCategoryPerception SkillCategory = "Perception"
)
const (
    PhysicalSkillMeleeCombat  SkillName = "Melee Combat"
    PhysicalSkillRangedCombat SkillName = "Ranged Combat"
    PhysicalSkillBackstab     SkillName = "Backstab"
    PhysicalSkillTackle       SkillName = "Tackle"
    PhysicalSkillTools        SkillName = "Tool usage"
    PhysicalSkillRepair       SkillName = "Repair"

    ThievingSkillLockpicking  SkillName = "Lockpicking"
    ThievingSkillPickpocket   SkillName = "Pickpocket"
    ThievingSkillSneak        SkillName = "Sneak"
    ThievingSkillGlasscutting SkillName = "Glasscutting"
    //    ThievingSkillDisguise     SkillName = "Disguise"
    //  ThievingSkillForgery      SkillName = "Forgery"

    SocialSkillDeceive    SkillName = "Deceive"
    SocialSkillPersuade   SkillName = "Persuade"
    SocialSkillIntimidate SkillName = "Intimidate"
    SocialSkillBluff      SkillName = "Bluff"
    SocialSkillSpotLies   SkillName = "Spot Lies"

    OutdoorsSkillSurvival  SkillName = "Survival"
    OutdoorsSkillHunting   SkillName = "Hunting"
    OutdoorsSkillHerbalism SkillName = "Herbalism"

    AthleticsSkillClimbing SkillName = "Climbing"
    AthleticsSkillSwimming SkillName = "Swimming"

    LanguageSkillCommon   SkillName = "Common"
    LanguageSkillAnimals  SkillName = "Animals"
    LanguageSkillMonsters SkillName = "Monsters"

    PerceptionSkillSpotHidden  SkillName = "Spot Hidden"
    PerceptionSkillDangerSense SkillName = "Danger Sense"
    PerceptionSkillAssess      SkillName = "Assess"

    CombatSkillJellyJab SkillName = "Jelly Jab"
)

type SkillSet struct {
    skills map[SkillName]SkillLevel
}

func lengthOfLongestSkillName(names []SkillName) int {
    result := 0
    for _, skillName := range names {
        if len(skillName) > result {
            result = len(skillName)
        }
    }
    return result

}
func getSkillNames(category SkillCategory) []SkillName {
    switch category {
    case SkillCategoryThieving:
        return []SkillName{
            ThievingSkillLockpicking,
            ThievingSkillPickpocket,
            ThievingSkillSneak,
            ThievingSkillGlasscutting,
        }
    case SkillCategorySocial:
        return []SkillName{
            SocialSkillDeceive,
            SocialSkillPersuade,
            SocialSkillIntimidate,
            SocialSkillBluff,
            SocialSkillSpotLies,
        }
    case SkillCategoryOutdoors:
        return []SkillName{
            OutdoorsSkillSurvival,
            OutdoorsSkillHunting,
        }
    case SkillCategoryAthletics:
        return []SkillName{
            AthleticsSkillClimbing,
            AthleticsSkillSwimming,
        }
    case SkillCategoryLanguages:
        return []SkillName{
            LanguageSkillCommon,
            LanguageSkillAnimals,
            LanguageSkillMonsters,
        }
    case SkillCategoryPerception:
        return []SkillName{
            PerceptionSkillSpotHidden,
            PerceptionSkillDangerSense,
        }
    }
    return []SkillName{}
}
func GetAllSkillNames() []SkillName {
    return []SkillName{
        ThievingSkillLockpicking,  // lvl 0-10
        ThievingSkillPickpocket,   // lvl 0-10
        ThievingSkillSneak,        // lvl 0-10
        ThievingSkillGlasscutting, // lvl 0-10
        SocialSkillDeceive,        // lvl 0-10
        SocialSkillPersuade,       // lvl 0-10
        SocialSkillIntimidate,     // lvl 0-10
        SocialSkillBluff,          // lvl 0-10
        SocialSkillSpotLies,       // lvl 0-10
        OutdoorsSkillSurvival,     // lvl 0-10
        OutdoorsSkillHunting,      // lvl 0-10
        AthleticsSkillClimbing,
        AthleticsSkillSwimming,
        LanguageSkillCommon,
        LanguageSkillAnimals,
        LanguageSkillMonsters,
        PhysicalSkillTackle,
        PhysicalSkillTools,
        PerceptionSkillSpotHidden,  // lvl 0-10
        PerceptionSkillDangerSense, // lvl 0-10
    }
}
func NewSkillSet() SkillSet {
    return SkillSet{
        skills: make(map[SkillName]SkillLevel),
    }
}

func (s *SkillSet) AddMasteryInAllSkills() {
    for _, skill := range GetAllSkillNames() {
        s.skills[skill] = SkillLevelMaster
    }
}

func (s *SkillSet) GetSkillLevel(skill SkillName) int {
    return int(s.skills[skill])
}

func (s *SkillSet) HasSkill(skill SkillName) bool {
    _, ok := s.skills[skill]
    return ok
}

func (s *SkillSet) IncrementSkill(skill SkillName) {
    currentLevel := s.skills[skill]
    if currentLevel >= 4 {
        return
    }
    s.skills[skill] = SkillLevel(currentLevel + 1)
}

// what do we really want?
// a chance to always succeed, no matter how low the odds.
// no streaks of bad luck, eg. three misses in a row with a 95% hit chance

// the first part can just be covered by keeping track of the number of tries and inserting a success
// skill 1: expect a success every 10th try
// skill 2: expect a success every 5th try
// skill 3: expect a success every 3rd try
// skill 4: expect a success every 2nd-3rd try
// skill 5: expect a success every 2nd try (50%)

// this here is expected to usually succeed, but sometimes fail
// we want to avoid streaks of bad luck there is a minimum failure distance that we should ensure
// skill 6: expect a failure every 2nd-3rd try
//  -> fail dist: 2 - so if we had a fail, the next try should succeed with an additional 50% chance
// skill 7: expect a failure every 3rd try
//  -> fail dist: 3 - so if we had a fail, the next two tries should succeed with an additional 66% chance
// skill 8: expect a failure every 5th try
//  -> fail dist: 5 - so if we had a fail, the next four tries should succeed with an additional 80% chance
// skill 9: expect a failure every 10th try
// -> fail dist: 10 - so if we had a fail, the next nine tries should succeed with an additional 90% chance
// skill 10: always succeed

func (s *SkillSet) AsStringTable() []string {
    var result []util.TableRow
    names := GetAllSkillNames()
    for _, skillName := range names {
        if s.GetSkillLevel(skillName) == 0 {
            continue
        }
        value := s.GetLevel(skillName).ToString()
        result = append(result, util.TableRow{Label: string(skillName), Columns: []string{value}})
    }

    if len(result) == 0 {
        return []string{"No skills"}
    }
    return util.TableLayout(result)
}

func (s *SkillSet) HasSkills(skills map[string]int) bool {
    for skill, level := range skills {
        if s.GetSkillLevel(SkillName(skill)) < level {
            return false
        }
    }
    return true
}

func (s *SkillSet) HasSkillAt(skill string, level int) bool {
    return s.GetSkillLevel(SkillName(skill)) >= level
}

func (s *SkillSet) GetLevel(skill SkillName) SkillLevel {
    return s.skills[skill]
}

func (s *SkillSet) DecrementSkill(name SkillName) {
    currentLevel := s.skills[name]
    if currentLevel <= 0 {
        return
    }
    newLevel := currentLevel - 1
    if newLevel <= 0 {
        delete(s.skills, name)
        return
    }
    s.skills[name] = SkillLevel(newLevel)
}

func (s *SkillSet) SetSkill(name SkillName, value int) {
    if value <= 0 {
        delete(s.skills, name)
        return
    }
    value = min(4, max(1, value))
    s.skills[name] = SkillLevel(value)
}
