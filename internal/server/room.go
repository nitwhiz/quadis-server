package server

import (
	"bloccs-server/pkg/event"
	"bloccs-server/pkg/game"
	"github.com/google/uuid"
	"log"
	"sync"
	"time"
)

type Room struct {
	ID            string
	Players       map[string]*Player
	playersMutex  *sync.Mutex
	games         map[string]*game.Game
	gamesMutex    *sync.Mutex
	eventBus      *event.Bus
	roomWaitGroup *sync.WaitGroup
	createAt      time.Time
	gamesRunning  bool
}

func NewRoom() *Room {
	b := event.NewBus()

	r := &Room{
		ID:            uuid.NewString(),
		Players:       map[string]*Player{},
		games:         map[string]*game.Game{},
		gamesMutex:    &sync.Mutex{},
		eventBus:      b,
		playersMutex:  &sync.Mutex{},
		roomWaitGroup: &sync.WaitGroup{},
		createAt:      time.Now(),
		gamesRunning:  false,
	}

	b.AddChannel(event.ChannelRoom)

	return r
}

func (r *Room) AreGamesRunning() bool {
	return r.gamesRunning
}

func (r *Room) ShouldClose() bool {
	r.playersMutex.Lock()
	defer r.playersMutex.Unlock()

	if len(r.Players) == 0 && time.Since(r.createAt) > time.Minute*15 {
		return true
	}

	return false
}

func (r *Room) Start() {
	r.gamesMutex.Lock()

	for _, g := range r.games {
		g.Start()
	}

	r.gamesMutex.Unlock()

	r.playersMutex.Lock()

	for _, player := range r.Players {
		go func(p *Player) {
			defer r.RemovePlayer(p)
			defer r.roomWaitGroup.Done()

			r.roomWaitGroup.Add(1)

			for {
				msg, err := p.ReadMessage()

				if err != nil {
					log.Println("error reading message", err)
					return
				}

				r.gamesMutex.Lock()

				if _, ok := r.games[p.ID]; ok {
					r.games[p.ID].Command(string(msg))
				}

				r.gamesMutex.Unlock()
			}
		}(player)
	}

	r.playersMutex.Unlock()

	r.gamesRunning = true

	r.eventBus.Publish(event.New(event.ChannelRoom, event.GameStart, nil))
}

// todo: room lifecycle: start, stop, resetGames <- gameovers; game summary screens

func (r *Room) Stop() {
	log.Println("stopping room")

	r.gamesMutex.Lock()
	defer r.gamesMutex.Unlock()

	for _, g := range r.games {
		g.Stop()
	}

	r.roomWaitGroup.Wait()

	r.gamesRunning = false
}
