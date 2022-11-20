package item

import (
	"github.com/nitwhiz/quadis-server/pkg/game"
	"time"
)

func NewLockRotation() *Item {
	return &Item{
		Type: TypeLockRotation,
		Activate: func(sourceGame *game.Game, room Room) {
			if room == nil {
				return
			}

			sourceGame.GetFallingPiece().SetRotationLocked(true)

			<-time.After(time.Second * 5)

			sourceGame.GetFallingPiece().SetRotationLocked(false)
		},
	}
}
