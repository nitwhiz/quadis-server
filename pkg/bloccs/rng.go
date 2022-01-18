package bloccs

import "math/rand"

type RNG struct {
	rand *rand.Rand
	bag  []*Piece
}

func NewRNG(seed int64) *RNG {
	r := RNG{
		rand: rand.New(rand.NewSource(seed)),
	}

	r.NextBag()

	return &r
}

func (r *RNG) NextBag() {
	var b []*Piece

	for _, p := range Pieces {
		b = append(b, NewPiece(p))
	}

	r.rand.Shuffle(len(b), func(i, j int) {
		b[i], b[j] = b[j], b[i]
	})

	r.bag = b
}

func (r *RNG) NextPiece() *Piece {
	p, b := r.bag[0], r.bag[1:]

	if len(b) == 0 {
		r.NextBag()
	} else {
		r.bag = b
	}

	return p
}
