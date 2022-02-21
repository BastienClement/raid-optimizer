package main

import (
	"fmt"
	"strings"
)

type Class int

const (
	Warrior Class = 1 + iota
	Paladin
	Hunter
	Rogue
	Priest
	DeathKnight
	Shaman
	Mage
	Warlock
	Monk
	Druid
	DemonHunter
)

var classes = map[string]Class{
	"warrior":     Warrior,
	"paladin":     Paladin,
	"hunter":      Hunter,
	"rogue":       Rogue,
	"priest":      Priest,
	"dk":          DeathKnight,
	"deathknight": DeathKnight,
	"shaman":      Shaman,
	"mage":        Mage,
	"warlock":     Warlock,
	"monk":        Monk,
	"druid":       Druid,
	"dh":          DemonHunter,
	"demonhunter": DemonHunter,
}

func (cls Class) String() string {
	for str, c := range classes {
		if cls == c {
			return str
		}
	}

	return fmt.Sprintf("<Class %d>", cls)
}

func ParseClass(str string) Class {
	if cls, found := classes[str]; found {
		return cls
	}

	panic("Unknown class: " + str)
}

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

type Role int

const (
	Tank Role = 1 << iota
	Healer
	Melee
	Ranged
)

func (r Role) String() string {
	switch r {
	case Tank:
		return "Tank"
	case Healer:
		return "Heal"
	case Melee:
		return "Melee"
	case Ranged:
		return "Ranged"
	}

	return fmt.Sprintf("<Role %d>", r)
}

func ParseRole(str string) Role {
	switch strings.ToLower(str) {
	case "tank":
		return Tank
	case "healer":
		return Healer
	case "melee":
		return Melee
	case "ranged":
		return Ranged
	}

	panic(fmt.Sprintf("Unknown role: %s", str))
}

func GetRole(spec Specialization) Role {
	switch spec {
	case WarriorProtection, PaladinProtection, DeathKnightBlood, MonkBrewmaster, DruidGuardian, DemonHunterVengeance:
		return Tank
	case PaladinHoly, MonkMistweaver, PriestDiscipline, PriestHoly, ShamanRestoration, DruidRestoration:
		return Healer
	case WarriorArm, WarriorFury, PaladinRetribution, HunterSurvival, RogueAssassination, RogueOutlaw, RogueSubtlety,
		DeathKnightFrost, DeathKnightUnholy, ShamanEnhancement, MonkWindwalker, DruidFeral, DemonHunterHavoc:
		return Melee
	case HunterBeastMaster, HunterMarksmanship, PriestShadow, ShamanElemental, MageArcane, MageFire, MageFrost,
		WarlockAffliction, WarlockDemonology, WarlockDestruction, DruidBalance:
		return Ranged
	}

	panic(fmt.Sprintf("Unknown role for spec: %s", spec))
}

type Armor int

const (
	Cloth Armor = 1 + iota
	Leather
	Mail
	Plate
)

func (a Armor) String() string {
	switch a {
	case Cloth:
		return "Cloth"
	case Leather:
		return "Leather"
	case Mail:
		return "Mail"
	case Plate:
		return "Plate"
	}

	return fmt.Sprintf("<Armor %d>", a)
}

func ArmorForClass(class Class) Armor {
	switch class {
	case Priest, Mage, Warlock:
		return Cloth
	case Rogue, Monk, Druid, DemonHunter:
		return Leather
	case Hunter, Shaman:
		return Mail
	case Warrior, Paladin, DeathKnight:
		return Plate
	}

	panic(fmt.Sprintf("Unknown armor for class: %s", class))
}

func ClassColor(class Class) (int, int, int) {
	switch class {
	case Warrior:
		return 198, 155, 109
	case Paladin:
		return 244, 140, 186
	case Hunter:
		return 170, 211, 114
	case Rogue:
		return 255, 244, 104
	case Priest:
		return 255, 255, 255
	case DeathKnight:
		return 196, 30, 58
	case Shaman:
		return 0, 112, 221
	case Mage:
		return 63, 199, 235
	case Warlock:
		return 135, 136, 238
	case Monk:
		return 0, 255, 152
	case Druid:
		return 255, 124, 10
	case DemonHunter:
		return 163, 48, 201
	}

	panic("Unknown color")
}
