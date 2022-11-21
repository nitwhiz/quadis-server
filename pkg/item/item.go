package item

import (
	"github.com/nitwhiz/quadis-server/pkg/game"
)

const TypeTornado = "tornado"
const TypeOnlyIPieces = "only_i_pieces"
const TypeLockRotation = "lock_rotation"

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

var All = []*Item{
	NewTornado(),
	NewOnlyIPieces(),
	NewLockRotation(),
}
