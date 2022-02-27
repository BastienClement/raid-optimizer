package main

import (
	"log"
	"math/rand"
)

func (X *Genome) MutSwap(rng *rand.Rand) {
	type RaidStats struct {
		Count   float64
		Healers float64
	}

	stats := make([]RaidStats, X.RaidCount)
	playerRaids := make([]int, len(players))

	for cid, rid := range X.Distribution {
		if rid >= 0 {
			char := &roster[cid]
			playerRaids[char.Player] |= (1 << rid)
			stats[rid].Count += 1
			if char.Role == Healer {
				stats[rid].Healers += 1
			}
		}
	}

	dist := make([]int, len(X.Distribution))
again:
	copy(dist, X.Distribution)

	for _, aid := range rng.Perm(len(dist)) {
		a, ar := roster[aid], dist[aid]
		for _, charIndex := range rng.Perm(len(playerCharacters[a.Player])) {
			bid := playerCharacters[a.Player][charIndex]
			if aid == bid {
				continue // We got the exact same char
			}

			b, br := roster[bid], dist[bid]

			if (a.Main && br == -1) || (b.Main && ar == -1) {
				continue // Cannot swap a main with a char on the bench
			}

			if a.Role != b.Role && (a.Role == Tank || b.Role == Tank) {
				continue // Cannot swap a tank with a non-tank
			}

			if a.Role != b.Role && (a.Role == Healer || b.Role == Healer) {
				// If only one of the char is a healer, we need to be careful not to break anything
				var aHealerDiff float64 = -1
				if a.Role == Healer {
					aHealerDiff = 1
				}

				if ar >= 0 {
					newRatioA := (stats[ar].Healers - aHealerDiff) / stats[ar].Count
					if newRatioA < healerMinRatio || newRatioA > healerMaxRatio {
						continue // Introducing this character would break the healer ratio
					}
				}

				if br >= 0 {
					newRatioB := (stats[br].Healers + aHealerDiff) / stats[br].Count
					if newRatioB < healerMinRatio || newRatioB > healerMaxRatio {
						continue // Introducing this character would break the healer ratio
					}
				}
			}

			dist[aid] = br
			dist[bid] = ar
			goto done
		}
	}

	log.Printf("Cannot swap anything")
	return

done:
	if checkViability && !Viable(dist, X.RaidCount) {
		log.Printf("Swap goto again")
		goto again
	}
	copy(X.Distribution, dist)
}
