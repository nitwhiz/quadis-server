package rng

import (
	"github.com/nitwhiz/bloccs-server/pkg/piece"
)

type RPG struct {
	*RBG[*piece.Piece]
}

func NewRPG(seed int64) *RPG {
	return &RPG{
		NewRBG[*piece.Piece](seed, func() []*piece.Piece {
			var b []*piece.Piece

			for _, p := range piece.All {
				b = append(b, piece.New(p))
			}

			return b
		}),
	}
}
