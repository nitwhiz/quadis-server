package room

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nitwhiz/quadis-server/pkg/communication"
	"github.com/nitwhiz/quadis-server/pkg/event"
	"github.com/nitwhiz/quadis-server/pkg/game"
	"github.com/nitwhiz/quadis-server/pkg/player"
	"github.com/nitwhiz/quadis-server/pkg/prom"
)

func (r *Room) RemoveGame(id string) {
	r.gamesMutex.Lock()

	if g, ok := r.games[id]; ok {
		r.bus.Unsubscribe(id)

		g.ToggleOver(true)
		delete(r.games, id)

		prom.TotalGames.Sub(1)

		r.bus.Publish(&event.Event{
			Type:    event.TypeLeave,
			Origin:  event.OriginRoom(r.GetId()),
			Payload: g.ToPayload(),
		})

		r.gamesMutex.Unlock()

		if r.targets != nil {
			r.targets.Randomize()
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
			r.mu.Lock()
			defer r.mu.Unlock()

			r.gameOverCount += 1

			if r.gameOverCount >= len(r.games)-1 {
				go r.StopGames(false)
				go r.publishScores()
			}
		},
		ActivateItemCallback: func(g *game.Game) {
			r.itemDistribution.ActivateItem(g)
		},
		Seed: r.randomSeed.NextInt64(),
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

	prom.TotalGames.Add(1)

	r.gamesMutex.Unlock()

	// todo: move isHost into player struct; promote new host if host leaves the game
	err = r.HandshakeAck(c, g, isHost)

	if err != nil {
		return err
	}

	r.bus.Subscribe(gameId, c)

	r.bus.Publish(&event.Event{
		Type:    event.TypeJoin,
		Origin:  event.OriginRoom(r.GetId()),
		Payload: g.ToPayload(),
	})

	if r.targets != nil {
		r.targets.Randomize()
	}

	return nil
}

func (r *Room) GetHostGame() *game.Game {
	r.gamesMutex.RLock()
	defer r.gamesMutex.RUnlock()

	for _, g := range r.games {
		if g.IsHost() {
			return g
		}
	}

	return nil
}

func (r *Room) GetRunningGamesCount() int {
	r.gamesMutex.RLock()
	defer r.gamesMutex.RUnlock()

	runningCount := 0

	for _, g := range r.games {
		if !g.IsOver() {
			runningCount += 1
		}
	}

	return runningCount
}
