package main

import (
	"fmt"
	"strings"
)

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
	default:
		panic(fmt.Sprintf("Unknown role: %s", str))
	}
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
	default:
		panic(fmt.Sprintf("Unknown role for spec: %s", spec))
	}
}
