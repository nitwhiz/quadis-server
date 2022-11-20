package item

import (
	"github.com/nitwhiz/quadis-server/pkg/game"
	"log"
)

const TypeTornado = "tornado"

// todo: the necessity of this interface must mean that the architecture could be better (?)

type Room interface {
	GetTargetGameId(gameId string) string
	GetGame(id string) *game.Game
	GetGames() map[string]*game.Game
}

type ActivateFunc func(sourceGame *game.Game, room Room)

type Item struct {
	Type     string
	Activate ActivateFunc
}

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

			targetGame := room.GetGame(targetId)

			if targetGame == nil || targetGame.IsOver() {
				return
			}

			targetGame.GetField().ShuffleTokens()

			log.Println("tornado applied!")
		},
	}
}
