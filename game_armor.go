package main

import (
	"fmt"
)

type Armor int

const (
	Cloth Armor = iota
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

var armorForClass = [...]Armor{
	Warrior:     Plate,
	Paladin:     Plate,
	Hunter:      Mail,
	Rogue:       Leather,
	Priest:      Cloth,
	DeathKnight: Plate,
	Shaman:      Mail,
	Mage:        Cloth,
	Warlock:     Cloth,
	Monk:        Leather,
	Druid:       Leather,
	DemonHunter: Leather,
}

func ArmorForClass(class Class) Armor {
	return armorForClass[class]
}
