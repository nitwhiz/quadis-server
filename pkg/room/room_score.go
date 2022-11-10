package room

import (
	"github.com/nitwhiz/quadis-server/pkg/event"
	"github.com/nitwhiz/quadis-server/pkg/game"
	"github.com/nitwhiz/quadis-server/pkg/score"
)

type playerScorePayload struct {
	Game  *game.Payload  `json:"game"`
	Score *score.Payload `json:"score"`
}

func (r *Room) publishScores() {
	r.gamesMutex.RLock()
	defer r.gamesMutex.RUnlock()

	var scores []*playerScorePayload

	for _, g := range r.games {
		scores = append(scores, &playerScorePayload{
			Game:  g.ToPayload(),
			Score: g.GetScore().ToPayload(),
		})
	}

	r.bus.Publish(&event.Event{
		Type:    event.TypeRoomScores,
		Origin:  event.OriginRoom(r.GetId()),
		Payload: scores,
	})
}
