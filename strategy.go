package main

import (
	"log"
	"strings"
)

type Strategy interface {
	LoadChar(char *Character, record []string)
	Prepare()
	Fitness(X *Genome) float64
	PrintStats(X *Genome)
}

func ParseStrategy(s string) Strategy {
	switch strings.ToLower(s) {
	case "armor":
		return &ArmorStrategy{}
	case "token":
		return &TokenStrategy{}
	}

	log.Fatalf("Unknown strategy: %s", s)
	return nil
}
