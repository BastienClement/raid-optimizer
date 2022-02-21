package main

import "math/rand"

type Xoshiro256ssSource struct {
	s [4]uint64
}

var _ rand.Source64 = &Xoshiro256ssSource{}

func rol64(x uint64, k int) uint64 {
	return (x << k) | (x >> (64 - k))
}

func (x *Xoshiro256ssSource) Int63() int64 {
	return int64(x.Uint64() >> 1)
}

func (x *Xoshiro256ssSource) Uint64() uint64 {
	result := rol64(x.s[1]*5, 7) * 9
	t := x.s[1] << 17

	x.s[2] ^= x.s[0]
	x.s[3] ^= x.s[1]
	x.s[1] ^= x.s[2]
	x.s[0] ^= x.s[3]

	x.s[2] ^= t
	x.s[3] = rol64(x.s[3], 45)

	return result
}

func (x *Xoshiro256ssSource) Seed(s int64) {
	t := uint64(s)
	x.s[0] = uint64(t)
	x.s[1] = rol64(t, 16)
	x.s[2] = rol64(t, 32)
	x.s[3] = rol64(t, 48)
}
