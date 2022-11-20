package piece

import "github.com/nitwhiz/quadis-server/pkg/rng"

type Generator struct {
	*rng.Bag[*Piece]
}

func NewGenerator(seed int64) *Generator {
	return &Generator{
		rng.NewBag[*Piece](seed, func() []*Piece {
			var b []*Piece

			for _, p := range All {
				b = append(b, p)
			}

			return b
		}),
	}
}
