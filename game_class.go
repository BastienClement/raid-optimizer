package main

import (
	"fmt"
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
