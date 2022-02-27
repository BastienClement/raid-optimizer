package main

import (
	"math/rand"

	"github.com/MaxHalford/eaopt"
)

func (X *Genome) Crossover(Y eaopt.Genome, rng *rand.Rand) {
	// TODO: maybe implement crossover ? This seems hard to do and mutate-only is working well enough.
	X.Mutate(rng)
	Y.Mutate(rng)
}
