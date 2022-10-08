package server

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nitwhiz/bloccs-server/pkg/event"
	"github.com/nitwhiz/bloccs-server/pkg/game"
	"github.com/nitwhiz/bloccs-server/pkg/rng"
	"github.com/nitwhiz/bloccs-server/pkg/score"
	"log"
	"sync"
	"time"
)

type Room struct {
	ID                      string
	Players                 map[string]*Player
	playersMutex            *sync.Mutex
	eventBus                *event.Bus
	roomWaitGroup           *sync.WaitGroup
	createAt                time.Time
	gamesRunning            bool
	isStopping              bool
	bedrockTargetMap        map[string]string
	bedrockTargetMapMutex   *sync.RWMutex
	randomPlayerIdGenerator *rng.RSG
	gameOverPlayerCount     int
}

func NewRoom() *Room {
	b := event.NewBus()

	r := &Room{
		ID:                    uuid.NewString(),
		Players:               map[string]*Player{},
		eventBus:              b,
		playersMutex:          &sync.Mutex{},
		roomWaitGroup:         &sync.WaitGroup{},
		createAt:              time.Now(),
		gamesRunning:          false,
		isStopping:            false,
		bedrockTargetMap:      map[string]string{},
		bedrockTargetMapMutex: &sync.RWMutex{},
		gameOverPlayerCount:   0,
	}

	r.randomPlayerIdGenerator = rng.NewRSG(time.Now().UnixMilli(), func() []string {
		var pids []string

		for pid := range r.Players {
			pids = append(pids, pid)
		}

		return pids
	})

	b.AddChannel(event.ChannelRoom)

	return r
}

func (r *Room) UpdateBedrockTargetMap() {
	r.bedrockTargetMapMutex.Lock()
	defer r.bedrockTargetMapMutex.Unlock()

	r.playersMutex.Lock()
	defer r.playersMutex.Unlock()

	r.bedrockTargetMap = map[string]string{}

	var playerIds []string

	for pid, p := range r.Players {
		if !p.game.IsOver {
			playerIds = append(playerIds, pid)
		}
	}

	for _, pId := range playerIds {
		rpId := r.randomPlayerIdGenerator.NextElement()

		if pId != rpId {
			r.bedrockTargetMap[pId] = rpId
		}
	}

	r.eventBus.Publish(event.New(event.ChannelRoom, event.UpdateBedrockTargets, &event.UpdateBedrockTargetsPayload{
		Targets: r.bedrockTargetMap,
	}))
}

func (r *Room) GetPlayerCount() int {
	r.playersMutex.Lock()
	defer r.playersMutex.Unlock()

	return len(r.Players)
}

func (r *Room) Join(conn *websocket.Conn) error {
	p := NewPlayer(conn, game.New(r.eventBus, r.ID))

	if err := r.handshakeHello(p); err != nil {
		return err
	}

	r.AddPlayer(p)

	p.Listen(func() {
		r.RemovePlayer(p)
		r.UpdateBedrockTargetMap()
	})

	return nil
}

func (r *Room) AreGamesRunning() bool {
	return r.gamesRunning
}

func (r *Room) ShouldClose() bool {
	r.playersMutex.Lock()
	defer r.playersMutex.Unlock()

	if !r.isStopping && len(r.Players) == 0 && time.Since(r.createAt) > time.Minute*15 {
		return true
	}

	return false
}

func (r *Room) Start() {
	if r.gamesRunning {
		return
	}

	r.ResetGames()
	r.UpdateBedrockTargetMap()

	r.gameOverPlayerCount = 0

	for _, player := range r.Players {
		(func(player *Player) {
			// todo: rewrite
			player.StartGame(func() {
				r.eventBus.Publish(&event.Event{
					Channel: fmt.Sprintf("update/%s", player.ID),
					Type:    event.GameOver,
					Payload: &event.PlayerGameOverPayload{
						Player: event.PlayerPayload{
							ID:       player.ID,
							Name:     player.Name,
							CreateAt: player.CreateAt,
						},
					},
				})

				// todo: locking; race conditions for room struct access

				player.StopGame()
				r.UpdateBedrockTargetMap()

				r.gameOverPlayerCount++

				if r.gameOverPlayerCount >= len(r.Players)-1 {
					for _, sp := range r.Players {
						sp.StopGame()
					}

					r.gamesRunning = false
				}
			})
		})(player)
	}

	r.gamesRunning = true

	r.eventBus.Publish(event.New(event.ChannelRoom, event.GameStart, nil))
}

func (r *Room) GetScores() map[string]*score.Score {
	r.playersMutex.Lock()
	defer r.playersMutex.Unlock()

	scores := map[string]*score.Score{}

	for _, p := range r.Players {
		s, l := p.game.GetScore()

		playerScore := score.New()

		playerScore.Score = s
		playerScore.Lines = l

		scores[p.ID] = playerScore
	}

	return scores
}

func (r *Room) ResetGames() {
	r.playersMutex.Lock()
	defer r.playersMutex.Unlock()

	for _, p := range r.Players {
		p.ResetGame()
	}
}

func (r *Room) Stop() {
	r.isStopping = true

	log.Println("stopping room")

	r.playersMutex.Lock()
	defer r.playersMutex.Unlock()

	for _, p := range r.Players {
		p.StopGame()
	}

	r.roomWaitGroup.Wait()
	r.gamesRunning = false
}
