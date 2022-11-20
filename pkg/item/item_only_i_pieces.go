package item

import (
	"github.com/nitwhiz/quadis-server/pkg/game"
	"github.com/nitwhiz/quadis-server/pkg/piece"
	"time"
)

func NewOnlyIPieces() *Item {
	return &Item{
		Type: TypeOnlyIPieces,
		Activate: func(sourceGame *game.Game, room Room) {
			if room == nil {
				return
			}

			sourceGame.SetOverridePiece(&piece.I)

			<-time.After(time.Second * 5)

			sourceGame.SetOverridePiece(nil)
		},
	}
}
