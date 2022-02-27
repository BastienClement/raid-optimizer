package main

import (
	"fmt"
)

type Specialization uint64

const (
	WarriorArm Specialization = 1 << iota
	WarriorFury
	WarriorProtection
	PaladinHoly
	PaladinProtection
	PaladinRetribution
	HunterBeastMaster
	HunterMarksmanship
	HunterSurvival
	RogueAssassination
	RogueOutlaw
	RogueSubtlety
	PriestDiscipline
	PriestHoly
	PriestShadow
	DeathKnightBlood
	DeathKnightFrost
	DeathKnightUnholy
	ShamanElemental
	ShamanEnhancement
	ShamanRestoration
	MageArcane
	MageFire
	MageFrost
	WarlockAffliction
	WarlockDemonology
	WarlockDestruction
	MonkBrewmaster
	MonkMistweaver
	MonkWindwalker
	DruidBalance
	DruidFeral
	DruidGuardian
	DruidRestoration
	DemonHunterHavoc
	DemonHunterVengeance
)

var specs = map[Class]map[string]Specialization{
	Warrior:     {"arm": WarriorArm, "fury": WarriorFury, "protection": WarriorProtection},
	Paladin:     {"holy": PaladinHoly, "protection": PaladinProtection, "retribution": PaladinRetribution},
	Hunter:      {"beastmaster": HunterBeastMaster, "marksmanship": HunterMarksmanship, "survival": HunterSurvival},
	Rogue:       {"assassination": RogueAssassination, "subtlety": RogueSubtlety, "outlaw": RogueOutlaw},
	Priest:      {"discipline": PriestDiscipline, "holy": PriestHoly, "shadow": PriestShadow},
	DeathKnight: {"blood": DeathKnightBlood, "frost": DeathKnightFrost, "unholy": DeathKnightUnholy},
	Shaman:      {"elemental": ShamanElemental, "enhancement": ShamanEnhancement, "restoration": ShamanRestoration},
	Mage:        {"arcane": MageArcane, "fire": MageFire, "frost": MageFrost},
	Warlock:     {"affliction": WarlockAffliction, "demonology": WarlockDemonology, "destruction": WarlockDestruction},
	Monk:        {"brewmaster": MonkBrewmaster, "mistweaver": MonkMistweaver, "windwalker": MonkWindwalker},
	Druid:       {"balance": DruidBalance, "feral": DruidFeral, "guardian": DruidGuardian, "restoration": DruidRestoration},
	DemonHunter: {"havoc": DemonHunterHavoc, "vengeance": DemonHunterVengeance},
}

func (spec Specialization) String() string {
	for _, clsSpecs := range specs {
		for str, s := range clsSpecs {
			if s == spec {
				return str
			}
		}
	}

	return fmt.Sprintf("<Spec %d>", spec)
}

func ParseSpec(cls Class, str string) Specialization {
	if spec, found := specs[cls][str]; found {
		return spec
	}

	panic(fmt.Sprintf("Unknown class/spec: %s/%s", cls, str))
}
