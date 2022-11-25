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

			targetId := room.GetTargetGameId(sourceGame.GetId())

			if targetId == "" {
				return
			}

			room.UpdateItemAffection(targetId, TypeLockRotation)

			targetGame := room.GetGame(targetId)

			if targetGame == nil || targetGame.IsOver() {
				return
			}

			fp := targetGame.GetFallingPiece()

			fp.SetRotationLocked(true)

			<-time.After(time.Second * 5)

			fp.SetRotationLocked(false)

			room.UpdateItemAffection(targetId, TypeNone)
		},
	}
}
