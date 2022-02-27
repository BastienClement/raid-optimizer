package main

import (
	"log"
	"math"
	"math/rand"
)

func (X *Genome) MutResizeShrink(rng *rand.Rand) {
	if X.RaidCount == minRaids {
		return
	}

	droppedRaid := X.RaidCount - 1
	basePlayerRaids := make([]int, len(players))

	type RaidStats struct {
		Count         float64
		Healers       float64
		ExtraCapacity float64
	}

	baseStats := make([]RaidStats, droppedRaid)

	droppedMains := make([]int, 0)
	droppedAlts := make([]int, 0)

	for cid, rid := range X.Distribution {
		if rid >= 0 {
			char := roster[cid]
			basePlayerRaids[char.Player] |= 1 << rid

			if rid == droppedRaid {
				if char.Main {
					droppedMains = append(droppedMains, cid)
				} else {
					droppedAlts = append(droppedAlts, cid)
				}
			} else {
				baseStats[rid].Count += 1
				if char.Role == Healer {
					baseStats[rid].Healers += 1
				}
			}
		}
	}

	for rid := range baseStats {
		baseStats[rid].ExtraCapacity = math.Floor(baseStats[rid].Healers/healerMinRatio - baseStats[rid].Count)
	}

	dist := make([]int, len(roster))
	playerRaids := make([]int, len(basePlayerRaids))
	stats := make([]RaidStats, droppedRaid)
again:
	copy(dist, X.Distribution)
	copy(playerRaids, basePlayerRaids)
	copy(stats, baseStats)

	rng.Shuffle(len(droppedMains), func(i, j int) {
		droppedMains[i], droppedMains[j] = droppedMains[j], droppedMains[i]
	})
	rng.Shuffle(len(droppedAlts), func(i, j int) {
		droppedAlts[i], droppedAlts[j] = droppedAlts[j], droppedAlts[i]
	})

	// Inject mains back into remaining groups
nextMain:
	for _, cid := range droppedMains {
		char := roster[cid]
	nextRaid:
		for _, rid := range rng.Perm(droppedRaid) {
		retry:
			if playerRaids[char.Player]&(1<<rid) == 0 {
				// Player is not in that raid, the actual method to inject the char depends on the role
				switch char.Role {
				case Tank:
					// For a tank, we need to boot one of them onto the bench
					for _, oid := range rng.Perm(len(dist)) {
						if dist[oid] != rid {
							continue // This char is not in the target raid
						}
						other := roster[oid]
						if other.Role == Tank && !other.Main {
							// We found a non-main tank in the target raid. Boot it to the bench.
							dist[oid] = -1
							playerRaids[other.Player] &= ^(1 << rid)
							dist[cid] = rid
							playerRaids[char.Player] |= (1 << rid)

							continue nextMain
						}
					}

					// Unable to inject in the current raid, let's try the next one
					continue nextRaid

				case Healer:
					// For a healer, we attempt to add it to the comp if it does not break ratio
					newRatio := (stats[rid].Healers + 1) / (stats[rid].Count + 1)
					if newRatio > healerMaxRatio {
						continue nextRaid
					}

					dist[cid] = rid
					playerRaids[char.Player] |= (1 << rid)

					// Update stats
					stats[rid].Healers += 1
					stats[rid].Count += 1
					stats[rid].ExtraCapacity = math.Floor(stats[rid].Healers/healerMinRatio - stats[rid].Count)
					continue nextMain

				case Melee, Ranged:
					// For a DPS, we attempt to use one of the ExtraCapacity if possible, otherwise we boot an alt from the raid
					if stats[rid].ExtraCapacity > 0 {
						stats[rid].Count += 1
						stats[rid].ExtraCapacity -= 1
					} else {
						for _, oid := range rng.Perm(len(dist)) {
							if dist[oid] != rid {
								continue // This char is not in the target raid
							}
							other := roster[oid]
							if other.Role == char.Role && !other.Main {
								// We found a non-main dps in the target raid. Boot it to the bench.
								dist[oid] = -1
								playerRaids[other.Player] &= ^(1 << rid)
								goto altDpsReplaced
							}
						}

						// We failed to replace someone in this raid
						continue nextRaid
					}

				altDpsReplaced:
					dist[cid] = rid
					playerRaids[char.Player] |= (1 << rid)
					continue nextMain
				}
			} else {
				// Player is already in the raid, attempt to replace it
				for oid := range dist {
					if dist[oid] != rid {
						continue // This char is not in the target raid
					}
					other := roster[oid]
					if other.Player == char.Player {
						if other.Role == char.Role || other.Role == Melee || other.Role == Ranged {
							// Same role or replacing DPS, just replace alt
							dist[oid] = -1
							dist[cid] = rid
							continue nextMain
						} else {
							// Switching from tank/healer to non-tank/healer, we need to bring another tank/healer from the bench
							for jid := range dist {
								if dist[jid] < 0 {
									joker := roster[jid]
									if joker.Role != other.Role || joker.Player == char.Player || playerRaids[joker.Player]&(1<<rid) != 0 {
										continue
									}

									dist[oid] = -1
									dist[jid] = rid

									playerRaids[other.Player] &= ^(1 << rid)
									playerRaids[joker.Player] |= (1 << rid)

									goto retry
								}
							}
						}
					}
				}
			}
		}

		log.Printf("Failed to insert %s into remaining raids, aborting", char)
		return
	}

	// Handle dropped alts
	for _, cid := range droppedAlts {
		dist[cid] = -1 // Put them on the bench
	}

	if checkViability && !Viable(dist, X.RaidCount-1) {
		log.Printf("Shrink goto again")
		goto again
	}
	X.RaidCount -= 1
	copy(X.Distribution, dist)
}
