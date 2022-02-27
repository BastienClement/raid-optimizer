package main

import (
	"fmt"
	"strings"
)

type Token int

const (
	Mystic Token = iota
	Venerated
	Zenith
	Dreadful
)

var tokenForClass = [...]Token{
	Warrior:     Zenith,
	Paladin:     Venerated,
	Hunter:      Mystic,
	Rogue:       Zenith,
	Priest:      Venerated,
	DeathKnight: Dreadful,
	Shaman:      Venerated,
	Mage:        Mystic,
	Warlock:     Dreadful,
	Monk:        Zenith,
	Druid:       Mystic,
	DemonHunter: Dreadful,
}

func TokenForClass(c Class) Token {
	return tokenForClass[c]
}

func (t Token) String() string {
	switch t {
	case Mystic:
		return "Mystic"
	case Venerated:
		return "Venerated"
	case Zenith:
		return "Zenith"
	case Dreadful:
		return "Dreadful"
	}

	return fmt.Sprintf("<Token %d>", t)
}

type TokenSlot int

const (
	SlotHead TokenSlot = iota
	SlotShoulders
	SlotChest
	SlotHands
	SlotLegs
)

var slots = map[string]TokenSlot{
	"head":      SlotHead,
	"shoulders": SlotShoulders,
	"chest":     SlotChest,
	"hands":     SlotHands,
	"legs":      SlotLegs,
}

func (ts TokenSlot) String() string {
	switch ts {
	case SlotHead:
		return "Head"
	case SlotShoulders:
		return "Shoulders"
	case SlotChest:
		return "Chest"
	case SlotHands:
		return "Hands"
	case SlotLegs:
		return "Legs"
	}

	return fmt.Sprintf("<TokenSloth %d>", ts)
}

type TokenSlotSet int

func (tss TokenSlotSet) Has(ts TokenSlot) bool {
	return tss&(1<<ts) != 0
}

func (tss *TokenSlotSet) Set(ts TokenSlot) {
	*tss |= (1 << ts)
}

func (tss TokenSlotSet) String() (str string) {
	if tss.Has(SlotHead) {
		str += "H"
	} else {
		str += "-"
	}
	if tss.Has(SlotShoulders) {
		str += "S"
	} else {
		str += "-"
	}
	if tss.Has(SlotChest) {
		str += "C"
	} else {
		str += "-"
	}
	if tss.Has(SlotHands) {
		str += "H"
	} else {
		str += "-"
	}
	if tss.Has(SlotLegs) {
		str += "L"
	} else {
		str += "-"
	}
	return
}

func ParseTokenSlots(str string) (set TokenSlotSet) {
	if str != "" {
		for _, s := range strings.Split(str, "/") {
			s = strings.ToLower(s)
			if slot, found := slots[s]; found {
				set.Set(slot)
				continue
			}

			panic(fmt.Sprintf("Unknown slot: %s", s))
		}
	}
	return
}
