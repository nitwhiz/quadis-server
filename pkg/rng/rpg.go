package rng

import (
	"bloccs-server/pkg/piece"
	"math/rand"
)

type RPG struct {
	rand *rand.Rand
	bag  []*piece.Piece
}

func NewRPG(seed int64) *RPG {
	r := RPG{
		rand: rand.New(rand.NewSource(seed)),
	}

	r.NextBag()

	return &r
}

func (r *RPG) NextBag() {
	var b []*piece.Piece

	for _, p := range piece.All {
		b = append(b, piece.New(p))
	}

	r.rand.Shuffle(len(b), func(i, j int) {
		b[i], b[j] = b[j], b[i]
	})

	r.bag = b
}

func (r *RPG) NextPiece() *piece.Piece {
	p, b := r.bag[0], r.bag[1:]

	if len(b) == 0 {
		r.NextBag()
	} else {
		r.bag = b
	}

	return p
}
