package rng

import (
	"math/rand"
)

type BagGenerator[ElementType any] func() []ElementType

type RBG[ElementType any] struct {
	rand   *rand.Rand
	bag    []ElementType
	bagGen BagGenerator[ElementType]
}

func NewRBG[ElementType any](seed int64, bagGen BagGenerator[ElementType]) *RBG[ElementType] {
	r := RBG[ElementType]{
		rand:   rand.New(rand.NewSource(seed)),
		bagGen: bagGen,
	}

	r.NextBag()

	return &r
}

func (r *RBG[any]) NextBag() {
	b := r.bagGen()

	r.rand.Shuffle(len(b), func(i, j int) {
		b[i], b[j] = b[j], b[i]
	})

	r.bag = b
}

func (r *RBG[any]) NextElement() any {
	e, b := r.bag[0], r.bag[1:]

	if len(b) == 0 {
		r.NextBag()
	} else {
		r.bag = b
	}

	return e
}
