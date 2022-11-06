package room

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nitwhiz/quadis-server/pkg/communication"
	"github.com/nitwhiz/quadis-server/pkg/event"
	"github.com/nitwhiz/quadis-server/pkg/game"
	"github.com/nitwhiz/quadis-server/pkg/player"
)

func (r *Room) RemoveGame(id string) {
	r.gamesMutex.Lock()

	if g, ok := r.games[id]; ok {
		r.bus.Unsubscribe(id)

		g.ToggleOver()
		delete(r.games, id)

		r.bus.Publish(&event.Event{
			Type:    event.TypeLeave,
			Origin:  event.OriginRoom(r.GetId()),
			Payload: g.ToPayload(),
		})

		r.gamesMutex.Unlock()

		if r.bedrockDistribution != nil {
			r.bedrockDistribution.Randomize()
		}
	} else {
		r.gamesMutex.Unlock()
	}
}

func (r *Room) CreateGame(ws *websocket.Conn) error {
	gameId := uuid.NewString()

	c := communication.NewConnection(&communication.Settings{
		WS:            ws,
		ParentContext: r.ctx,
		PreStopCallback: func() {
			r.RemoveGame(gameId)
		},
	})

	hrm, err := r.HandshakeGreeting(c)

	if err != nil {
		return err
	}

	gameSettings := game.Settings{
		Id:            gameId,
		EventBus:      r.bus,
		Connection:    c,
		Player:        player.New(hrm.PlayerName),
		ParentContext: r.ctx,
		OverCallback: func() {
			defer r.mu.Unlock()
			r.mu.Lock()

			r.gameOverCount += 1

			if r.gameOverCount >= len(r.games)-1 {
				go r.StopGames()
				go r.publishScores()
			}
		},
	}

	if r.rules.BedrockEnabled {
		gameSettings.BedrockChannel = r.bedrockDistribution.Channel
	}

	g := game.New(&gameSettings)

	isHost := false

	r.gamesMutex.Lock()

	// this is not 100% correct, but it's correct enough
	isHost = len(r.games) == 0

	r.games[gameId] = g

	r.gamesMutex.Unlock()

	// todo: move isHost into player struct; promote new host if host leaves the game
	err = r.HandshakeAck(c, g, isHost)

	r.bus.Subscribe(gameId, c)

	r.bus.Publish(&event.Event{
		Type:    event.TypeJoin,
		Origin:  event.OriginRoom(r.GetId()),
		Payload: g.ToPayload(),
	})

	if r.bedrockDistribution != nil {
		r.bedrockDistribution.Randomize()
	}

	if err != nil {
		return err
	}

	return nil
}

func (r *Room) GetRunningGamesCount() int {
	defer r.gamesMutex.RUnlock()
	r.gamesMutex.RLock()

	runningCount := 0

	for _, g := range r.games {
		if !g.IsOver() {
			runningCount += 1
		}
	}

	return runningCount
}
