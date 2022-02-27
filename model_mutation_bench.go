package main

import (
	"log"
	"math/rand"
)

func (X *Genome) MutBench(rng *rand.Rand) {
	type RaidStats struct {
		Count   float64
		Healers float64
	}

	var stats [RMAX]RaidStats
	var benchable [CMAX]int
	j := 0

	for cid, rid := range X.Distribution {
		char := &roster[cid]
		if rid >= 0 {
			stats[rid].Count += 1
			if char.Role == Healer {
				stats[rid].Healers += 1
			}
			if char.Main || char.Role == Tank {
				continue // Mains and tanks are immune to benching
			}
			benchable[j] = cid
			j++
		}
	}

	// Remove impossible benches
	for i:= 0; i < j; {
		var newRatio float64
		var healerDiff float64

		rid := X.Distribution[benchable[i]]

		if stats[rid].Count <= 10 {
			goto impossible // Raid is already at minimum size
		}

		if roster[benchable[i]].Role == Healer {
			healerDiff = 1
		}
		newRatio = (stats[rid].Healers - healerDiff) / (stats[rid].Count - 1)
		if newRatio < healerMinRatio || newRatio > healerMaxRatio {
			goto impossible // Benching this character would break the healer ratio
		}

		i++
		continue

	impossible:
		benchable[i] = benchable[j-1]
		j--
	}

	if j < 1 {
		// If we cannot bench anything, let's introduce someone instead
		X.MutIntroduce(rng)
		return
	}

	// Benching a random char
	X.Distribution[benchable[rng.Intn(j)]] = -1

	if checkViability && !X.Viable() {
		log.Fatalf("Bench failed")
	}
}
