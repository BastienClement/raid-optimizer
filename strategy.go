package main

import (
	"fmt"
	"log"
	"math"
	"strings"
)

type Strategy interface {
	Prepare()
	Fitness(X *Genome) float64
	PrintStats(X *Genome)
}

func ParseStrategy(s string) Strategy {
	switch strings.ToLower(s) {
	case "armor":
		return &ArmorStrategy{}
	}

	log.Fatalf("Unknown strategy: %s", s)
	return nil
}

// Optimizing armor trading

type ArmorStrategy struct {
	targets map[Armor]float64
}

func (ArmorStrategy) String() string {
	return "Armor"
}

type ArmorRaidStats struct {
	ArmorReceiver map[Armor]int
	ArmorTrader   map[Armor]int
}

func (as *ArmorStrategy) Prepare() {
	log.Printf("Computing targets...")

	armorReceiver := map[Armor]int{Cloth: 0, Leather: 0, Mail: 0, Plate: 0}
	armorTrader := map[Armor]int{Cloth: 0, Leather: 0, Mail: 0, Plate: 0}

	for _, char := range roster {
		if char.Main {
			armorReceiver[ArmorForClass(char.Class)] += 1
		} else {
			armorTrader[ArmorForClass(char.Class)] += 1
		}
	}

	as.targets = map[Armor]float64{Cloth: 0, Leather: 0, Mail: 0, Plate: 0}
	for _, armor := range []Armor{Cloth, Leather, Mail, Plate} {
		as.targets[armor] = float64(armorTrader[armor]) / float64(armorReceiver[armor])
	}

	log.Printf("Theoretical optimums: %+v", as.targets)
}

func (as ArmorStrategy) ComputeStats(X *Genome) []ArmorRaidStats {
	raids := make([]ArmorRaidStats, X.RaidCount)
	for i := range raids {
		raids[i].ArmorReceiver = map[Armor]int{Cloth: 0, Leather: 0, Mail: 0, Plate: 0}
		raids[i].ArmorTrader = map[Armor]int{Cloth: 0, Leather: 0, Mail: 0, Plate: 0}
	}

	for cid, rid := range X.Distribution {
		char := roster[cid]
		if rid < 0 {
			continue // Benched
		}

		raid := &raids[rid]
		if char.Main {
			raid.ArmorReceiver[ArmorForClass(char.Class)] += 1
		} else {
			raid.ArmorTrader[ArmorForClass(char.Class)] += 1
		}
	}

	return raids
}

func (as ArmorStrategy) Fitness(X *Genome) float64 {
	raids := as.ComputeStats(X)

	armorRatio := map[Armor][]float64{Cloth: {}, Leather: {}, Mail: {}, Plate: {}}
	for _, raid := range raids {
		for _, armor := range []Armor{Cloth, Leather, Mail, Plate} {
			if raid.ArmorReceiver[armor] > 0 {
				armorRatio[armor] = append(armorRatio[armor], float64(raid.ArmorTrader[armor])/float64(raid.ArmorReceiver[armor]))
			}
		}
	}

	var delta float64
	for armor, ratios := range armorRatio {
		for _, ratio := range ratios {
			delta += math.Abs(as.targets[armor] - ratio)
		}
	}

	return delta
}

func (as ArmorStrategy) PrintStats(X *Genome) {
	stats := as.ComputeStats(X)

	armorRatio := map[Armor][]float64{Cloth: {}, Leather: {}, Mail: {}, Plate: {}}
	for rid, raid := range stats {
		fmt.Printf("[Raid %2d] ", rid+1)
		for _, armor := range []Armor{Cloth, Leather, Mail, Plate} {
			fmt.Printf("%s %2d:%-2d", armor, raid.ArmorReceiver[armor], raid.ArmorTrader[armor])
			var ratio float64
			if raid.ArmorReceiver[armor] > 0 {
				ratio = float64(raid.ArmorTrader[armor]) / float64(raid.ArmorReceiver[armor])
				armorRatio[armor] = append(armorRatio[armor], ratio)
			}
			fmt.Printf(" (%f)", ratio)
			fmt.Printf("\t")
		}
		fmt.Printf("\n")
	}

	fmt.Printf("[Average] ")
	for _, armor := range []Armor{Cloth, Leather, Mail, Plate} {
		var sum float64
		var count float64
		for _, ratio := range armorRatio[armor] {
			sum += ratio
			count += 1
		}
		fmt.Printf("%s        %f \t", armor, sum/count)
	}
	fmt.Printf("\n")

	fmt.Printf("[Optimal] ")
	for _, armor := range []Armor{Cloth, Leather, Mail, Plate} {
		fmt.Printf("%s        %f \t", armor, as.targets[armor])
	}
	fmt.Printf("\n")
}
