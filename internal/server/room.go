package server

import (
	"bloccs-server/pkg/bloccs"
	"bloccs-server/pkg/event"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type Room struct {
	Id                   string                  `json:"id"`
	Players              map[string]*Player      `json:"players"`
	Games                map[string]*bloccs.Game `json:"games"`
	eventHandlerIds      map[string]int
	eventHandlerIdsMutex *sync.Mutex
	playersMutex         *sync.Mutex
	gamesMutex           *sync.Mutex
	eventBus             *event.Bus
	createAt             time.Time
	gamesRunning         bool
	isStopping           bool
}

func NewRoom() *Room {
	b := event.NewBus()

	b.Start()

	r := &Room{
		Id:                   uuid.NewString(),
		Players:              map[string]*Player{},
		playersMutex:         &sync.Mutex{},
		Games:                map[string]*bloccs.Game{},
		gamesMutex:           &sync.Mutex{},
		eventHandlerIds:      map[string]int{},
		eventHandlerIdsMutex: &sync.Mutex{},
		eventBus:             b,
		createAt:             time.Now(),
		gamesRunning:         false,
		isStopping:           false,
	}

	return r
}

func (r *Room) GetId() string {
	return r.Id
}

func (r *Room) GetPlayerCount() int {
	r.playersMutex.Lock()
	defer r.playersMutex.Unlock()

	return len(r.Players)
}

func (r *Room) Join(conn *websocket.Conn) error {
	g := bloccs.NewGame(r.eventBus, &bloccs.GameSettings{
		FallingPieceSpeed: 1,
		Seed:              r.Id,
		FieldWidth:        10,
		FieldHeight:       20,
	})
	p := NewPlayer(conn, g.GetId(), g.GetCommandChannel())

	fmt.Println("new player joining")

	if err := r.handshakeHello(p); err != nil {
		return err
	}

	r.RegisterPlayer(p, g)

	p.Listen(func() {
		r.RemoveGame(g)
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
	r.gamesMutex.Lock()
	defer r.gamesMutex.Unlock()

	for _, g := range r.Games {
		g.Start()
	}

	r.gamesRunning = true
}

func (r *Room) Stop() {
	r.isStopping = true

	log.Printf("stopping room %s ...\n", r.Id)

	r.gamesMutex.Lock()
	defer r.playersMutex.Unlock()

	for _, g := range r.Games {
		g.Stop()
	}

	r.gamesRunning = false
}
