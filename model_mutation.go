package main

import (
	"math/rand"
)

func (X *Genome) Mutate(rng *rand.Rand) {
	mutation := rng.Intn(100)
	/*if mutation < 10 {
		X.MutResize(rng) // Resize the raid comp
	} else */if mutation < 40 {
		X.MutSwap(rng) // Swap two characters
	} else if mutation < 70 {
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
