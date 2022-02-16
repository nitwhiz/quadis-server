package server

import (
	"bloccs-server/pkg/event"
	"bloccs-server/pkg/game"
	"bloccs-server/pkg/score"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type Room struct {
	ID            string
	Players       map[string]*Player
	playersMutex  *sync.Mutex
	eventBus      *event.Bus
	roomWaitGroup *sync.WaitGroup
	createAt      time.Time
	gamesRunning  bool
	isStopping    bool
}

func NewRoom() *Room {
	b := event.NewBus()

	r := &Room{
		ID:            uuid.NewString(),
		Players:       map[string]*Player{},
		eventBus:      b,
		playersMutex:  &sync.Mutex{},
		roomWaitGroup: &sync.WaitGroup{},
		createAt:      time.Now(),
		gamesRunning:  false,
		isStopping:    false,
	}

	b.AddChannel(event.ChannelRoom)

	return r
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
	r.playersMutex.Lock()

	for _, player := range r.Players {
		player.StartGame()
	}

	r.playersMutex.Unlock()

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
