package item

import (
	"github.com/nitwhiz/quadis-server/pkg/game"
)

func NewTornado() *Item {
	return &Item{
		Type: TypeTornado,
		Activate: func(sourceGame *game.Game, room Room) {
			if room == nil {
				return
			}

			targetId := room.GetTargetGameId(sourceGame.GetId())

			if targetId == "" {
				return
			}

			room.UpdateItemAffection(sourceGame.GetId(), TypeTornado)

			targetGame := room.GetGame(targetId)

			if targetGame == nil || targetGame.IsOver() {
				return
			}

			targetGame.GetField().ShuffleTokens()

			room.UpdateItemAffection(sourceGame.GetId(), TypeNone)
		},
	}
}
