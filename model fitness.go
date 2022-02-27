package main

import (
	"math"
)

func (X *Genome) Evaluate() (float64, error) {
	return math.Round(strategy.Fitness(X)*1000) + secondaryFitness(X), nil
}

// Evaluates secondary fitness critera. Must return a value <1.0.
func secondaryFitness(X *Genome) float64 {
	const (
		ArcaneIntellect uint8 = 1 << iota
		PwFortitude
		BattleShout
		ChaosBrand
		MysticTouch
	)

	const AllBuffs = ArcaneIntellect | PwFortitude | BattleShout | ChaosBrand | MysticTouch

	var raidBuffs [RMAX]uint8
	var count [RMAX]int

	for cid, rid := range X.Distribution {
		if rid < 0 {
			continue
		}
		count[rid] += 1

		if raidBuffs[rid] == AllBuffs {
			continue
		}

		char := roster[cid]
		var buf uint8
		switch char.Class {
		case Mage:
			buf = ArcaneIntellect
		case Priest:
			buf = PwFortitude
		case Warrior:
			buf = BattleShout
		case DemonHunter:
			buf = ChaosBrand
		case Monk:
			buf = MysticTouch
		}

		if buf > 0 {
			raidBuffs[rid] |= buf
		}
	}

	var missingBuffsMalus float64
	for rid := 0; rid < X.RaidCount; rid++ {
		for buff := ArcaneIntellect; buff <= MysticTouch; buff <<= 1 {
			if raidBuffs[rid]&buff == 0 {
				missingBuffsMalus += 1 / float64(X.RaidCount)
			}
		}
	}

	var min, max int = 30, 0
	for r := 0; r < X.RaidCount; r++ {
		if count[r] > max {
			max = count[r]
		}
		if count[r] < min {
			min = count[r]
		}
	}

	return (missingBuffsMalus/5)/10 + (float64(max-min)/20)/10000
}
