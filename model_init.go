package main

import (
	"log"
	"math"
	"math/rand"
)

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
		goto again 
		// FIXME: this might happen if playing with more than 2 tanks player because we may assign players with less 
		// chars first leaving the player with the most chars to fill every last slots...
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
