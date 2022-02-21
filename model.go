package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"

	"github.com/MaxHalford/eaopt"
)

const debugViability = true

type Genome struct {
	RaidCount    int
	Distribution []int
}

func (X *Genome) Viable() bool {
	return Viable(X.Distribution, X.RaidCount)
}

func Viable(distribution []int, size int) bool {
	type RaidStats struct {
		Count     int
		RoleCount struct {
			Tanks   int
			Healers int
			Melees  int
			Rangeds int
		}
		PlayerIndex BitSet
	}

	raids := make([]RaidStats, size)
	for i := range raids {
		raids[i].PlayerIndex = MakeBitSet(len(players))
	}

	for cid, rid := range distribution {
		char := roster[cid]
		if rid < 0 {
			if char.Main {
				if debugViability {
					fmt.Printf("Benched main\n")
				}
				return false // We benched a main
			}
			continue
		}

		raid := &raids[rid]

		if raid.PlayerIndex.Get(char.Player) {
			if debugViability {
				fmt.Printf("Duplicate\n")
			}
			return false // Duplicate player in the same raid
		} else {
			raid.PlayerIndex.Set(char.Player, true)
		}

		raid.Count += 1
		switch char.Role {
		case Tank:
			raid.RoleCount.Tanks += 1
		case Healer:
			raid.RoleCount.Healers += 1
		case Melee:
			raid.RoleCount.Melees += 1
		case Ranged:
			raid.RoleCount.Rangeds += 1
		}
	}

	for _, raid := range raids {
		// Validate raid viability
		if raid.Count < 10 || raid.Count > 30 {
			if debugViability {
				fmt.Printf("Bad size\n")
			}
			return false
		}
		if raid.RoleCount.Tanks != 2 {
			if debugViability {
				fmt.Printf("Bad tank count\n")
			}
			return false
		}
		healerRatio := float64(raid.RoleCount.Healers) / float64(raid.Count)
		if healerRatio < 0.175 || healerRatio > 0.5 {
			if debugViability {
				fmt.Printf("Bad healer ratio\n")
			}
			return false
		}
	}

	return true
}

func (X *Genome) Evaluate() (float64, error) {
	return strategy.Fitness(X), nil
}

func (X *Genome) Mutate(rng *rand.Rand) {
	mutation := rng.Float64()
	if mutation < 0.1 {
		//X.MutResize(rng) // Resize the raid comp
	} else if mutation < 0.4 {
		X.MutSwap(rng) // Swap two characters
	} else if mutation < 0.7 {
		X.MutIntroduce(rng) // Introduce one character
	} else {
		X.MutBench(rng) // Bench one character
	}
}

func (X *Genome) MutResize(rng *rand.Rand) {
	// Start by shuffling raids around without touching raid comp
	// This has no risk to break viability but will allow shrink/expand to alway operate on the same raid
	raidShuffle := make([]int, X.RaidCount)
	for r := range raidShuffle {
		raidShuffle[r] = r
	}
	rng.Shuffle(X.RaidCount, func(i, j int) {
		raidShuffle[i], raidShuffle[j] = raidShuffle[j], raidShuffle[i]
	})
	for cid, rid := range X.Distribution {
		if rid >= 0 {
			X.Distribution[cid] = raidShuffle[rid]
		}
	}

	// Either Shrink or Expand the roster
	switch rng.Uint64() % 1 {
	case 0:
		X.MutResizeShrink(rng)
	case 1:
		X.MutResizeExpand(rng)
	}
}

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

	if !Viable(dist, droppedRaid) {
		log.Printf("Shrink goto again")
		goto again
	}
	X.RaidCount -= 1
	copy(X.Distribution, dist)
}

func (X *Genome) MutResizeExpand(rng *rand.Rand) {
	if X.RaidCount == maxRaids {
		return
	}

	// TODO: implement
}

func (X *Genome) MutSwap(rng *rand.Rand) {
	type RaidStats struct {
		Count   float64
		Healers float64
	}

	stats := make([]RaidStats, X.RaidCount)
	playerRaids := make([]int, len(players))

	for cid, rid := range X.Distribution {
		if rid >= 0 {
			char := roster[cid]
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
		for _, bid := range rng.Perm(len(dist)) {
			a, ar := roster[aid], dist[aid]
			b, br := roster[bid], dist[bid]

			if a.Player != b.Player || aid == bid {
				continue // Not the same player or the exact same char
			}

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
	if !Viable(dist, X.RaidCount) {
		log.Printf("Swap goto again")
		goto again
	}
	copy(X.Distribution, dist)
}

func (X *Genome) MutIntroduce(rng *rand.Rand) {
	type RaidStats struct {
		Count   float64
		Healers float64
	}

	stats := make([]RaidStats, X.RaidCount)
	playerRaids := make([]int, len(players))

	for cid, rid := range X.Distribution {
		if rid >= 0 {
			char := roster[cid]
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

	for _, cid := range rng.Perm(len(dist)) {
		if dist[cid] != -1 {
			continue // This char is not on the bench
		}

		char := roster[cid]
		if char.Role == Tank {
			continue // Tanks are immune to introduction
		}

		var healerDiff float64
		if char.Role == Healer {
			healerDiff = 1
		}

		for _, rid := range rng.Perm(X.RaidCount) {
			if stats[rid].Count >= float64(raidSize) {
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

	// If we cannot introduce anything, let's bench someone instead
	X.MutBench(rng)
	return

done:
	if !Viable(dist, X.RaidCount) {
		log.Printf("Introduce goto again")
		goto again
	}
	copy(X.Distribution, dist)
}

func (X *Genome) MutBench(rng *rand.Rand) {
	type RaidStats struct {
		Count   float64
		Healers float64
	}
	stats := make([]RaidStats, X.RaidCount)
	for cid, rid := range X.Distribution {
		if rid >= 0 {
			char := roster[cid]
			stats[rid].Count += 1
			if char.Role == Healer {
				stats[rid].Healers += 1
			}
		}
	}

	dist := make([]int, len(X.Distribution))
again:
	copy(dist, X.Distribution)

	for _, cid := range rng.Perm(len(dist)) {
		rid := dist[cid]
		if rid < 0 {
			continue // Cannot bench a char that's already on the bench
		}

		if stats[rid].Count <= 10 {
			continue // Nothing left to bench here...
		}

		char := roster[cid]
		if char.Main || char.Role == Tank {
			continue // Mains and tanks are immune to benching
		}

		var healerDiff float64
		if char.Role == Healer {
			healerDiff = 1
		}

		newRatio := (stats[rid].Healers - healerDiff) / (stats[rid].Count - 1)
		if newRatio < healerMinRatio || newRatio > healerMaxRatio {
			continue // Benching this character would break the healer ratio
		}

		dist[cid] = -1
		goto done
	}

	// If we cannot bench anything, let's introduce someone instead
	X.MutIntroduce(rng)
	return

done:
	if !Viable(dist, X.RaidCount) {
		log.Printf("Bench goto again")
		goto again
	}
	copy(X.Distribution, dist)
}

func (X *Genome) Crossover(y eaopt.Genome, rng *rand.Rand) {
	/*Y := y.(*Genome)
		if X.RaidCount != Y.RaidCount {
			log.Printf("Cross-over with various sizes")
			return
		}

		A := make([]int, len(X.Distribution))
		B := make([]int, len(Y.Distribution))
	again:
		copy(A, X.Distribution)
		copy(B, Y.Distribution)

		PrintRaid(X)
		PrintRaid(Y)
		//log.Fatal()

		for _, chars := range playerCharacters {
			if rng.Float64() < 0.5 {
				for _, cid := range chars {
					A[cid], B[cid] = B[cid], A[cid]
				}
			}
		}

		if !Viable(A, X.RaidCount) {
			copy(X.Distribution, A)
			PrintRaid(X)
			log.Fatalf("Crossover A goto again")
			goto again
		}
		if !Viable(B, X.RaidCount) {
			log.Fatalf("Crossover B goto again")
			goto again
		}

		copy(X.Distribution, A)
		copy(Y.Distribution, B)*/
}

func (X *Genome) Clone() eaopt.Genome {
	Y := Genome{
		RaidCount:    X.RaidCount,
		Distribution: make([]int, len(X.Distribution)),
	}
	copy(Y.Distribution, X.Distribution)
	return &Y
}

func MakeRaid(rng *rand.Rand) *Genome {
	X := Genome{
		RaidCount:    rng.Intn(maxRaids-minRaids+1) + minRaids,
		Distribution: make([]int, len(roster)),
	}

	// Keep track of which raids the player is participating in
	playerRaids := make([]int, len(players))

	// Prepare tank spots
	tankCount := X.RaidCount * 2
	tankSpots := make([]int, tankCount)
	for i := range tankSpots {
		tankSpots[i] = i
	}

	// Find out how many healers are actually spottable for this number of raids
	spottableHealers := 0
	for _, chars := range playerCharacters {
		playerHealers := 0
		for _, cid := range chars {
			if roster[cid].Role == Healer {
				playerHealers += 1
			}
			if playerHealers == X.RaidCount {
				break
			}
		}
		spottableHealers += playerHealers
	}

	// Prepare healers spots
	healPerRaid := int(math.Ceil(float64(spottableHealers) / float64(X.RaidCount)))
	healCount := healPerRaid * X.RaidCount
	healSpots := make([]int, healCount)
	for i := range healSpots {
		healSpots[i] = i / healPerRaid
	}

again:
	rng.Shuffle(tankCount, func(i, j int) {
		tankSpots[i], tankSpots[j] = tankSpots[j], tankSpots[i]
	})
	rng.Shuffle(healCount, func(i, j int) {
		healSpots[i], healSpots[j] = healSpots[j], healSpots[i]
	})

	for _, char := range roster {
		playerRaids[char.Player] = 0
	}

	// *** Dispatch tanks ***

	startTankSpot := 0
	for _, group := range [][]int{roleIndex.Tank.Mains, roleIndex.Tank.Alts} {
		// Shuffle chars in the group
		chars := make([]int, len(group))
		copy(chars, group)
		rng.Shuffle(len(chars), func(i, j int) {
			chars[i], chars[j] = chars[j], chars[i]
		})

	tank:
		for _, cid := range chars {
			char := roster[cid]

			// Attempt to find a tank spot for this char
			for spot := startTankSpot; spot < tankCount; spot++ {
				raid := tankSpots[spot] / 2
				raidMask := 1 << raid

				if playerRaids[char.Player]&raidMask != 0 {
					continue // Player already in this raid
				}

				X.Distribution[cid] = raid
				playerRaids[char.Player] |= raidMask

				tankSpots[startTankSpot], tankSpots[spot] = tankSpots[spot], tankSpots[startTankSpot]
				startTankSpot += 1

				continue tank
			}

			if char.Main {
				log.Fatalf("Unable to place main tank: %s", char)
			}
			X.Distribution[cid] = -1
		}
	}
	if startTankSpot != tankCount {
		goto again // this might happen if playing with more than 2 tanks player
		log.Fatalf("Failed to populate every tank slots: %d != %d", startTankSpot, tankCount)
	}

	// *** Dispatch healers ***

	raidHealsCount := make([]int, X.RaidCount)
	startHealSpot := 0
	for _, group := range [][]int{roleIndex.Heal.Mains, roleIndex.Heal.Alts} {
		// Shuffle chars in the group
		chars := make([]int, len(group))
		copy(chars, group)
		rng.Shuffle(len(chars), func(i, j int) {
			chars[i], chars[j] = chars[j], chars[i]
		})

	healer:
		for _, cid := range chars {
			char := roster[cid]

			// Attempt to find a healer spot for this char
			for spot := startHealSpot; spot < healCount; spot++ {
				raid := healSpots[spot]
				raidMask := 1 << raid

				if playerRaids[char.Player]&raidMask != 0 {
					continue // Player already in this raid
				}

				X.Distribution[cid] = raid
				playerRaids[char.Player] |= raidMask
				raidHealsCount[raid] += 1

				healSpots[startHealSpot], healSpots[spot] = healSpots[spot], healSpots[startHealSpot]
				startHealSpot += 1

				continue healer
			}

			if char.Main {
				log.Fatalf("Unable to place main healer: %s", char)
			}
			X.Distribution[cid] = -1
		}
	}

	// *** Dispatch DPSes ***

	requiredSlotsCount := 0
	bonusSlotsCount := 0
	dpsCountPerRaid := make([]int, X.RaidCount*2)
	for i := 0; i < X.RaidCount; i++ {
		rh := float64(raidHealsCount[i])
		required := int(math.Ceil(Max(10.0-2.0-rh, rh/healerMaxRatio-rh-2.0)))
		bonus := int(math.Floor(rh/healerMinRatio-rh-2.0)) - required

		dpsCountPerRaid[i*2] = required
		dpsCountPerRaid[i*2+1] = bonus

		requiredSlotsCount += required
		bonusSlotsCount += bonus
	}

	requiredSlots := make([]int, 0, requiredSlotsCount)
	bonusSlots := make([]int, 0, bonusSlotsCount)
	for i := 0; i < X.RaidCount; i++ {
		for j, count := 0, dpsCountPerRaid[i*2]; j < count; j++ {
			requiredSlots = append(requiredSlots, i)
		}
		for j, count := 0, dpsCountPerRaid[i*2+1]; j < count; j++ {
			bonusSlots = append(bonusSlots, i)
		}
	}
	rng.Shuffle(requiredSlotsCount, func(i, j int) {
		requiredSlots[i], requiredSlots[j] = requiredSlots[j], requiredSlots[i]
	})
	rng.Shuffle(bonusSlotsCount, func(i, j int) {
		bonusSlots[i], bonusSlots[j] = bonusSlots[j], bonusSlots[i]
	})

	startRequiredSlot, startBonusSlot := 0, 0
	for _, group := range [][]int{roleIndex.Dps.Mains, roleIndex.Dps.Alts} {
		// Shuffle chars in the group
		chars := make([]int, len(group))
		copy(chars, group)
		rng.Shuffle(len(chars), func(i, j int) {
			chars[i], chars[j] = chars[j], chars[i]
		})

	dps:
		for _, cid := range chars {
			char := roster[cid]

			// Attempt to find a dps spot for this char
			for spot := startRequiredSlot; spot < requiredSlotsCount; spot++ {
				raid := requiredSlots[spot]
				raidMask := 1 << raid
				if playerRaids[char.Player]&raidMask != 0 {
					continue // Player already in this raid
				}
				X.Distribution[cid] = raid
				playerRaids[char.Player] |= raidMask
				requiredSlots[startRequiredSlot], requiredSlots[spot] = requiredSlots[spot], requiredSlots[startRequiredSlot]
				startRequiredSlot += 1
				continue dps
			}
			for spot := startBonusSlot; spot < bonusSlotsCount; spot++ {
				raid := bonusSlots[spot]
				raidMask := 1 << raid
				if playerRaids[char.Player]&raidMask != 0 {
					continue // Player already in this raid
				}
				X.Distribution[cid] = raid
				playerRaids[char.Player] |= raidMask
				bonusSlots[startBonusSlot], bonusSlots[spot] = bonusSlots[spot], bonusSlots[startBonusSlot]
				startBonusSlot += 1
				continue dps
			}

			if char.Main {
				log.Fatalf("Unable to place main dps: %s", char)
			}
			X.Distribution[cid] = -1
		}
	}

	if !X.Viable() {
		goto again
	}

	return &X
}

type Speciator struct{}

var _ eaopt.Speciator = (*Speciator)(nil)

func (s Speciator) Apply(indis eaopt.Individuals, rng *rand.Rand) ([]eaopt.Individuals, error) {
	species := make([]eaopt.Individuals, maxRaids)
	for _, indi := range indis {
		idx := indi.Genome.(*Genome).RaidCount - 1
		species[idx] = append(species[idx], indi)
	}
	return species[minRaids-1:], nil
}

func (s Speciator) Validate() error {
	return nil
}
