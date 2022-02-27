package main

import (
	"fmt"
	"log"
	"math"
)

type ArmorStrategy struct {
	targets [4]float64
}

func (ArmorStrategy) String() string {
	return "Armor"
}

type ArmorRaidStats struct {
	ArmorReceiver [4]int
	ArmorTrader   [4]int
}

func (ArmorStrategy) LoadChar(char *Character, record []string) {
}

func (as *ArmorStrategy) Prepare() {
	log.Printf("Computing armor targets...")

	var armorReceiver, armorTrader [4]int

	for _, char := range roster {
		if char.Main {
			armorReceiver[ArmorForClass(char.Class)] += 1
		} else {
			armorTrader[ArmorForClass(char.Class)] += 1
		}
	}

	for i := Cloth; i <= Plate; i++ {
		as.targets[i] = float64(armorTrader[i]) / float64(armorReceiver[i])
	}

	log.Printf("Theoretical optimums: %+v", as.targets)
}

func (as ArmorStrategy) ComputeStats(X *Genome) [RMAX]ArmorRaidStats {
	var raids [RMAX]ArmorRaidStats

	for cid, rid := range X.Distribution {
		char := roster[cid]
		if rid < 0 {
			continue // Benched
		}

		if char.Main {
			raids[rid].ArmorReceiver[ArmorForClass(char.Class)] += 1
		} else {
			raids[rid].ArmorTrader[ArmorForClass(char.Class)] += 1
		}
	}

	return raids
}

func (as ArmorStrategy) Fitness(X *Genome) float64 {
	raids := as.ComputeStats(X)

	var delta float64
	for rid := 0; rid < X.RaidCount; rid++ {
		for i := Cloth; i <= Plate; i++ {
			if raids[rid].ArmorReceiver[i] > 0 {
				ratio := float64(raids[rid].ArmorTrader[i]) / float64(raids[rid].ArmorReceiver[i])
				delta += math.Abs(as.targets[i] - ratio)
			}
		}
	}

	return delta
}

func (as ArmorStrategy) PrintStats(X *Genome) {
	stats := as.ComputeStats(X)

	armorRatio := [4][]float64{{}, {}, {}, {}}
	for rid := 0; rid < X.RaidCount; rid++ {
		fmt.Printf("[Raid %2d] ", rid+1)
		for i := Cloth; i <= Plate; i++ {
			fmt.Printf("%s %2d:%-2d", i, stats[rid].ArmorReceiver[i], stats[rid].ArmorTrader[i])
			var ratio float64
			if stats[rid].ArmorReceiver[i] > 0 {
				ratio = float64(stats[rid].ArmorTrader[i]) / float64(stats[rid].ArmorReceiver[i])
				armorRatio[i] = append(armorRatio[i], ratio)
			}
			fmt.Printf(" (%f)", ratio)
			fmt.Printf("\t")
		}
		fmt.Printf("\n")
	}

	fmt.Printf("[Average] ")
	for i := Cloth; i <= Plate; i++ {
		var sum float64
		var count float64
		for _, ratio := range armorRatio[i] {
			sum += ratio
			count += 1
		}
		fmt.Printf("%s        %f \t", i, sum/count)
	}
	fmt.Printf("\n")

	fmt.Printf("[Optimal] ")
	for i := Cloth; i <= Plate; i++ {
		fmt.Printf("%s        %f \t", i, as.targets[i])
	}
	fmt.Printf("\n")
}
