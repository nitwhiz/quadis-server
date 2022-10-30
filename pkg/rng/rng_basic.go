package rng

import "math/rand"

type Basic struct {
	rand *rand.Rand
}

func NewBasic(seed int64) *Basic {
	return &Basic{
		rand: rand.New(rand.NewSource(seed)),
	}
}

func (r *Basic) NextFloat64() float64 {
	return r.rand.Float64()
}

func (r *Basic) NextInt64() int64 {
	return r.rand.Int63()
}

func (r *Basic) Probably(probability float64) bool {
	return r.NextFloat64() < probability
}
