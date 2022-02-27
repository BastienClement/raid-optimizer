package main

import (
	"github.com/MaxHalford/eaopt"
)

type Genome struct {
	RaidCount    int
	Distribution []int
}

func (X *Genome) Clone() eaopt.Genome {
	Y := Genome{
		RaidCount:    X.RaidCount,
		Distribution: make([]int, len(X.Distribution)),
	}
	copy(Y.Distribution, X.Distribution)
	return &Y
}
