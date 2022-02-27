package main

import (
	"flag"
	"fmt"
	"log"
	"math"
)

const (
	TokenRoleReceiver int = iota
	TokenRoleTrader
	TokenRoleNone
)

type TokenStrategy struct {
	targetSlots TokenSlotSet
	targets     [4][5]float64
	as          ArmorStrategy
}

func (TokenStrategy) String() string {
	return "Token"
}

type TokenRaidStats struct {
	ArmorReceiver [4][5]int
	ArmorTrader   [4][5]int
}

func (TokenStrategy) LoadChar(char *Character, record []string) {
	char.TokenSlots = ParseTokenSlots(record[5])
}

func (TokenStrategy) TokenRole(c Character, s TokenSlot) int {
	if c.TokenSlots.Has(s) || c.TokenSlots.Count() >= 4 {
		return TokenRoleTrader
	} else if c.Main {
		return TokenRoleReceiver
	} else {
		return TokenRoleNone
	}
}

func (ts *TokenStrategy) Prepare() {
	ts.targetSlots = ParseTokenSlots(flag.Arg(1))
	log.Printf("Computing token targets (%s)...", ts.targetSlots)

	var tokenReceiver, tokenTrader [4][5]int

	for _, char := range roster {
		for slot := SlotHead; slot <= SlotLegs; slot++ {
			if !ts.targetSlots.Has(slot) {
				continue
			}

			switch ts.TokenRole(char, slot) {
			case TokenRoleReceiver:
				tokenReceiver[TokenForClass(char.Class)][slot] += 1
			case TokenRoleTrader:
				tokenTrader[TokenForClass(char.Class)][slot] += 1
			}
		}
	}

	for t := Mystic; t <= Dreadful; t++ {
		for s := SlotHead; s <= SlotLegs; s++ {
			if tokenReceiver[t][s] > 0 {
				ts.targets[t][s] = float64(tokenTrader[t][s]) / float64(tokenReceiver[t][s])
			}
		}
	}

	log.Printf("Theoretical optimums: %+v", ts.targets)
	ts.as.Prepare()
}

func (ts TokenStrategy) ComputeStats(X *Genome) [RMAX]TokenRaidStats {
	var raids [RMAX]TokenRaidStats

	for cid, rid := range X.Distribution {
		char := roster[cid]
		if rid < 0 {
			continue // Benched
		}

		for slot := SlotHead; slot <= SlotLegs; slot++ {
			if !ts.targetSlots.Has(slot) {
				continue
			}
			switch ts.TokenRole(char, slot) {
			case TokenRoleReceiver:
				raids[rid].ArmorReceiver[TokenForClass(char.Class)][slot] += 1
			case TokenRoleTrader:
				raids[rid].ArmorTrader[TokenForClass(char.Class)][slot] += 1
			}
		}
	}

	return raids
}

func (ts TokenStrategy) Fitness(X *Genome) float64 {
	raids := ts.ComputeStats(X)

	var delta float64
	for rid := 0; rid < X.RaidCount; rid++ {
		for t := Mystic; t <= Dreadful; t++ {
			for s := SlotHead; s <= SlotLegs; s++ {
				if !ts.targetSlots.Has(s) {
					continue
				}
				if raids[rid].ArmorReceiver[t][s] > 0 {
					ratio := float64(raids[rid].ArmorTrader[t][s]) / float64(raids[rid].ArmorReceiver[t][s])
					delta += math.Abs(ts.targets[t][s] - ratio)
				}
			}
		}
	}

	return delta*100000 + ts.as.Fitness(X)
}

func (ts TokenStrategy) PrintStats(X *Genome) {
	stats := ts.ComputeStats(X)

	armorRatio := [4][5][]float64{{}, {}, {}, {}}
	for s := SlotHead; s <= SlotLegs; s++ {
		if !ts.targetSlots.Has(s) {
			continue
		}
		for rid := 0; rid < X.RaidCount; rid++ {
			fmt.Printf("[Raid %2d] ", rid+1)
			for t := Mystic; t <= Dreadful; t++ {
				fmt.Printf("%s %s %2d:%-2d", t, s, stats[rid].ArmorReceiver[t][s], stats[rid].ArmorTrader[t][s])
				var ratio float64
				if stats[rid].ArmorReceiver[t][s] > 0 {
					ratio = float64(stats[rid].ArmorTrader[t][s]) / float64(stats[rid].ArmorReceiver[t][s])
					armorRatio[t][s] = append(armorRatio[t][s], ratio)
				}
				fmt.Printf(" (%f)", ratio)
				fmt.Printf("\t")
			}
			fmt.Printf("\n")
		}

		fmt.Printf("[Average] ")
		for t := Mystic; t <= Dreadful; t++ {
			var sum float64
			var count float64
			for _, ratio := range armorRatio[t][s] {
				sum += ratio
				count += 1
			}
			fmt.Printf("%s %s        %f \t", t, s, sum/count)
		}
		fmt.Printf("\n")

		fmt.Printf("[Optimal] ")
		for t := Mystic; t <= Dreadful; t++ {
			fmt.Printf("%s %s        %f \t", t, s, ts.targets[t][s])
		}
		fmt.Printf("\n\n")
	}
	ts.as.PrintStats(X)
}
