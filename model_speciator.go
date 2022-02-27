package main

import (
	"math/rand"

	"github.com/MaxHalford/eaopt"
)

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
