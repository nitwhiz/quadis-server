package rng

import (
	"github.com/nitwhiz/quadis-server/pkg/piece"
)

type Piece struct {
	*Bag[*piece.Piece]
}

func NewPiece(seed int64) *Piece {
	return &Piece{
		NewBag[*piece.Piece](seed, func() []*piece.Piece {
			var b []*piece.Piece

			for _, p := range piece.All {
				b = append(b, piece.New(p))
			}

			return b
		}),
	}
}
