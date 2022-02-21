package main

import (
	"fmt"
	"sort"
)

func PrintRaid(X *Genome) {
	raids := make([][]Character, X.RaidCount+1)
	for i := range raids {
		raids[i] = make([]Character, 0)
	}

	for cid, rid := range X.Distribution {
		raids[rid+1] = append(raids[rid+1], roster[cid])
	}

	var longest int
	for _, raid := range raids {
		if len(raid) > longest {
			longest = len(raid)
		}

		sort.Slice(raid, func(i, j int) bool {
			a, b := raid[i], raid[j]
			if a.Role != b.Role {
				return a.Role < b.Role
			} else if a.Class != b.Class {
				return a.Class < b.Class
			} else {
				return a.Name < b.Name
			}
		})
	}

	type RaidStats struct {
		RoleCount map[Role]int
	}

	stats := make([]RaidStats, X.RaidCount+1)

	for i := 0; i <= X.RaidCount; i++ {
		stats[i].RoleCount = map[Role]int{Tank: 0, Healer: 0, Melee: 0, Ranged: 0}
	}

	for row := 0; row < longest; row++ {
		for col := 0; col <= X.RaidCount; col++ {
			if row >= len(raids[col]) {
				fmt.Printf("%s   ", Character{})
				continue
			}

			char := raids[col][row]
			stats[col].RoleCount[char.Role] += 1

			fmt.Printf("%s   ", char)
		}
		fmt.Print("\n")
	}

	fmt.Printf("%v\n", stats)
	strategy.PrintStats(X)
}
