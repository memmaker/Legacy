package game

import (
    "Legacy/util"
    "math/rand"
    "strconv"
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

    OutdoorsSkillSurvival SkillName = "Survival"
    OutdoorsSkillHunting  SkillName = "Hunting"

    AthleticsSkillClimbing SkillName = "Climbing"
    AthleticsSkillSwimming SkillName = "Swimming"

    LanguageSkillCommon   SkillName = "Common"
    LanguageSkillAnimals  SkillName = "Animals"
    LanguageSkillMonsters SkillName = "Monsters"

    PerceptionSkillSpotHidden  SkillName = "Spot Hidden"
    PerceptionSkillDangerSense SkillName = "Danger Sense"
)

type RNGInfo struct {
    FailsInRow     int // FailsInRow tells us how many tries ago the last fail was, eg. 1 means the last try failed, 2 means the try before that failed, etc.
    SuccessesInRow int // SuccessesInRow tells us how many tries ago the last success was, eg. 1 means the last try succeeded, 2 means the try before that succeeded, etc.
}
type SkillSet struct {
    skills     map[SkillName]int
    randomBags map[SkillName]RNGInfo
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
func getAllSkillNames() []SkillName {
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
        PerceptionSkillSpotHidden,  // lvl 0-10
        PerceptionSkillDangerSense, // lvl 0-10
    }
}
func NewSkillSet() SkillSet {
    return SkillSet{
        skills:     make(map[SkillName]int),
        randomBags: make(map[SkillName]RNGInfo),
    }
}

func (s *SkillSet) AddAll() {
    for _, skill := range getAllSkillNames() {
        s.skills[skill] = 10
    }
}

func (s *SkillSet) GetSkillLevel(skill SkillName) int {
    return s.skills[skill]
}

func (s *SkillSet) HasSkill(skill SkillName) bool {
    value, ok := s.skills[skill]
    return ok && value > 0
}

func (s *SkillSet) IncrementSkill(skill SkillName) {
    currentLevel := s.skills[skill]
    if currentLevel >= 10 {
        return
    }

    s.skills[skill] = currentLevel + 1
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

func (s *SkillSet) TenBasedSkillCheck(skill SkillName, modifier int) bool {
    skillValue := s.skills[skill]
    if skillValue <= 0 {
        return false // fail, for not investing
    }

    s.ensureBag(skill)

    bag := s.randomBags[skill]
    skillCheckResult := rand.Intn(10)+modifier < skillValue
    if skillCheckResult {
        bag.SuccessesInRow = 1
        bag.FailsInRow = 0
        s.randomBags[skill] = bag
        return true // quick exit, we are only interested in skewing the results in the positive direction
    }

    if skillValue <= 5 {
        lastSuccess := bag.SuccessesInRow + 1
        expectedSuccess := 10 / skillValue // eg. 10 / 5 = 2, so every 2nd try should succeed
        // es(1) = 10/1 = 10
        // es(2) = 10/2 = 5
        // es(3) = 10/3 = 3
        // es(4) = 10/4 = 2
        // es(5) = 10/5 = 2

        if lastSuccess > expectedSuccess { // 3 > 2 -> the last
            dist := lastSuccess - expectedSuccess
            bonusChance := float64(1/10) * float64(dist) // eg. 10% * 1 = 10% bonus chance
            skillCheckResult = rand.Float64() < bonusChance
        }
    } else if skillValue > 5 {
        lastSuccess := bag.SuccessesInRow + 1
        expectedSuccess := 10 / (10 - skillValue) //  10/5 -> 2, 10/4 -> 2, 10/3 -> 3, 10/2 -> 5, 10/1 -> 10, 10/0 -> 0
        // es(6) = 10/4 = 2
        // es(7) = 10/3 = 3
        // es(8) = 10/2 = 5
        // es(9) = 10/1 = 10
        // es(10) = 10/0 = 0

        if lastSuccess > expectedSuccess {
            dist := lastSuccess - expectedSuccess
            bonusChance := float64(1/10) * float64(dist)
            skillCheckResult = rand.Float64() < bonusChance
        }
    }

    if skillCheckResult {
        bag.SuccessesInRow = 1
        bag.FailsInRow = bag.FailsInRow + 1
    } else {
        bag.SuccessesInRow = bag.SuccessesInRow + 1
        bag.FailsInRow = 1
    }
    s.randomBags[skill] = bag

    return skillCheckResult
}

func (s *SkillSet) ensureBag(skill SkillName) {
    if _, ok := s.randomBags[skill]; !ok {
        s.randomBags[skill] = RNGInfo{}
    }
}

func (s *SkillSet) GetChanceOfSuccess(skill SkillName, modifier int) float64 {
    skillValue := s.skills[skill] - modifier
    if skillValue <= 0 {
        return 0 // fail, for not investing
    }
    s.ensureBag(skill)
    return float64(skillValue) / 10
}

func (s *SkillSet) AsStringTable() []string {
    var result []util.TableRow
    names := getAllSkillNames()
    for _, skillName := range names {
        value := s.skills[skillName]
        if value <= 0 {
            continue
        }
        result = append(result, util.TableRow{Label: string(skillName), Columns: []string{strconv.Itoa(value)}})
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
