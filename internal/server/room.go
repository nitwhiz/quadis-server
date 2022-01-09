package server

import (
	"bloccs-server/pkg/bloccs"
	"bloccs-server/pkg/event"
	"fmt"
	"github.com/google/uuid"
	"log"
	"sync"
)

type Room struct {
	ID            string             `json:"id"`
	Players       map[string]*Player `json:"players"`
	games         map[string]*bloccs.Game
	eventBus      *event.Bus
	playersMutex  *sync.Mutex
	roomWaitGroup *sync.WaitGroup
}

func NewRoom() *Room {
	return &Room{
		ID:            uuid.NewString(),
		Players:       map[string]*Player{},
		games:         map[string]*bloccs.Game{},
		eventBus:      event.NewBus(),
		playersMutex:  &sync.Mutex{},
		roomWaitGroup: &sync.WaitGroup{},
	}
}

func passEvent(p *Player, e *event.Event) error {
	bs, err := e.GetAsBytes()

	if err != nil {
		log.Printf("cannot get bytes of message")
		return err
	}

	if err := p.SendMessage(bs); err != nil {
		return err
	}

	return nil
}

func (r *Room) AddPlayer(p *Player) {
	r.playersMutex.Lock()

	if _, ok := r.Players[p.ID]; ok {
		return
	}

	if len(r.Players) >= 7 {
		return
	}

	g := bloccs.NewGame(r.eventBus, p.ID)

	r.eventBus.AddChannel(fmt.Sprintf("player/%s", p.ID))

	r.eventBus.Subscribe("game_update/.*", func(e *event.Event) {
		if err := passEvent(p, e); err != nil {
			r.RemovePlayer(p)
		}
	})

	r.eventBus.Subscribe(fmt.Sprintf("player/%s", p.ID), func(e *event.Event) {
		if err := passEvent(p, e); err != nil {
			r.RemovePlayer(p)
		}
	})

	r.games[p.ID] = g
	r.Players[p.ID] = p

	r.playersMutex.Unlock()

	r.eventBus.Publish(event.New(event.ChanBroadcast, bloccs.EventPlayerJoin, &event.Payload{
		"player": p,
	}))

	r.eventBus.Publish(event.New(fmt.Sprintf("player/%s", p.ID), bloccs.EventRoomInfo, &event.Payload{
		"room": r,
		"you":  p,
	}))
}

func (r *Room) RemovePlayer(p *Player) {
	defer func() {
		_ = p.Conn.Close()
		r.playersMutex.Unlock()
	}()

	r.playersMutex.Lock()

	if _, ok := r.Players[p.ID]; !ok {
		return
	}

	if g, ok := r.games[p.ID]; ok {
		g.Stop()
		// todo: game mutex: delete(r.games, p.ID)
	}

	r.eventBus.RemoveChannel(fmt.Sprintf("player/%s", p.ID))
	delete(r.Players, p.ID)

	r.eventBus.Publish(event.New(event.ChanBroadcast, bloccs.EventPlayerLeave, &event.Payload{
		"player": p,
	}))
}

func (r *Room) Start() {
	// todo: game mutex
	// todo: don't allow join after start

	for _, g := range r.games {
		g.Start()
	}

	r.playersMutex.Lock()

	for _, player := range r.Players {
		go func(p *Player) {
			defer r.roomWaitGroup.Done()
			defer r.RemovePlayer(p)

			r.roomWaitGroup.Add(1)

			for {
				_, msg, err := p.Conn.ReadMessage()

				if err != nil {
					log.Println("error reading message", err)
					return
				}

				if _, ok := r.games[p.ID]; ok {
					if r.games[p.ID].Command(string(msg)) {
						if r.games[p.ID].Field.FallingPiece.Dirty {
							r.games[p.ID].PublishFallingPieceUpdate()
						}

						if r.games[p.ID].Field.Dirty {
							r.games[p.ID].PublishFieldUpdate()
						}
					}
				}
			}
		}(player)
	}

	r.playersMutex.Unlock()
}

func (r *Room) Stop() {
	for _, g := range r.games {
		g.Stop()
	}

	r.roomWaitGroup.Wait()
}
