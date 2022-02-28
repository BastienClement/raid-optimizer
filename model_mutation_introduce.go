package main

import (
	"log"
	"math/rand"
)

func (X *Genome) MutIntroduce(rng *rand.Rand) {
	type RaidStats struct {
		Count   float64
		Healers float64
	}

	stats := make([]RaidStats, X.RaidCount)
	playerRaids := make([]int, len(players))
	benched := make([]int, 0, len(roster))

	for cid, rid := range X.Distribution {
		char := &roster[cid]
		if rid >= 0 {
			playerRaids[char.Player] |= (1 << rid)
			stats[rid].Count += 1
			if char.Role == Healer {
				stats[rid].Healers += 1
			}
		} else {
			if char.Role == Tank {
				continue // Tanks are immune to introduction
			}
			benched = append(benched, cid)
		}
	}

	dist := make([]int, len(X.Distribution))
again:
	copy(dist, X.Distribution)

	if len(benched) > 0 {
		for _, bid := range rng.Perm(len(benched)) {
			cid := benched[bid]
			char := &roster[cid]

			var healerDiff float64
			if char.Role == Healer {
				healerDiff = 1
			}

			for _, rid := range rng.Perm(X.RaidCount) {
				if stats[rid].Count >= float64(maxRaidSize) {
					continue // This raid is already full...
				}

				if playerRaids[char.Player]&(1<<rid) != 0 {
					continue // This player is already playing here
				}

				newRatio := (stats[rid].Healers + healerDiff) / (stats[rid].Count + 1)
				if newRatio < healerMinRatio || newRatio > healerMaxRatio {
					continue // Introducing this character would break the healer ratio
				}

				dist[cid] = rid
				goto done
			}
		}
	}

	// If we cannot introduce anything, let's bench someone instead
	X.MutBench(rng)
	return

done:
	if checkViability && !Viable(dist, X.RaidCount) {
		log.Printf("Introduce goto again")
		goto again
	}
	copy(X.Distribution, dist)
}
