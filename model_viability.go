package main

import (
	"fmt"
)

const debugViability = false

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
		if raid.Count < minRaidSize || raid.Count > maxRaidSize {
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
