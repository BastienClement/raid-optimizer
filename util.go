package main

import "math/rand"

type BitSet []uint64

const (
	zero64 uint64 = 0
	one64  uint64 = 1
)

func MakeBitSet(size int) BitSet {
	len := size / 64
	if mod := size % 64; mod > 0 {
		len += 1
	}
	return make(BitSet, len)
}

func (s BitSet) Get(idx int) bool {
	word := idx / 64
	bit := idx % 64
	return s[word]&(one64<<bit) != zero64
}

func (s BitSet) Set(idx int, value bool) {
	word := idx / 64
	bit := idx % 64
	if value {
		s[word] |= one64 << bit
	} else {
		s[word] &= ^(one64 << bit)
	}
}

func Max[T int | float64](first T, rest ...T) T {
	max := first
	for _, n := range rest {
		if n > max {
			max = n
		}
	}
	return max
}

func Min[T int | float64](first T, rest ...T) T {
	min := first
	for _, n := range rest {
		if n < min {
			min = n
		}
	}
	return min
}

// Perm returns, as a slice of n ints, a pseudo-random permutation of the integers
// in the half-open interval [0,n).
func Perm8(rng *rand.Rand, n int8) []int8 {
	m := make([]int8, n)
	for i := int8(0); i < n; i++ {
		j := rng.Int31n(int32(i + 1))
		m[i] = m[j]
		m[j] = int8(i)
	}
	return m
}
