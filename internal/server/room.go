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
	gamesMutex    *sync.Mutex
	eventBus      *event.Bus
	playersMutex  *sync.Mutex
	roomWaitGroup *sync.WaitGroup
}

func NewRoom() *Room {
	b := event.NewBus()

	r := &Room{
		ID:            uuid.NewString(),
		Players:       map[string]*Player{},
		games:         map[string]*bloccs.Game{},
		gamesMutex:    &sync.Mutex{},
		eventBus:      b,
		playersMutex:  &sync.Mutex{},
		roomWaitGroup: &sync.WaitGroup{},
	}

	b.AddChannel(bloccs.ChannelRoom)

	return r
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

	r.gamesMutex.Lock()
	r.games[p.ID] = g
	r.gamesMutex.Unlock()

	r.Players[p.ID] = p

	r.eventBus.Subscribe(bloccs.ChannelRoom, func(e *event.Event) {
		if err := passEvent(p, e); err != nil {
			log.Println("error passing event")
		}
	}, g)

	r.eventBus.AddChannel(fmt.Sprintf("player/%s", p.ID))

	r.eventBus.Subscribe("game_update/.*", func(e *event.Event) {
		if err := passEvent(p, e); err != nil {
			log.Println("error passing event")
		}
	}, g)

	r.eventBus.Subscribe(fmt.Sprintf("player/%s", p.ID), func(e *event.Event) {
		if err := passEvent(p, e); err != nil {
			log.Println("error passing event")
		}
	}, g)

	r.playersMutex.Unlock()

	// own join is received, too, maybe that's not a problem
	r.eventBus.Publish(event.New(bloccs.ChannelRoom, bloccs.EventPlayerJoin, &event.Payload{
		"player": p,
	}))
}

func (r *Room) RemovePlayer(p *Player) {
	r.playersMutex.Lock()

	defer func() {
		_ = p.Conn.Close()
		r.playersMutex.Unlock()
	}()

	if _, ok := r.Players[p.ID]; !ok {
		return
	}

	r.gamesMutex.Lock()

	r.eventBus.Unsubscribe(r.games[p.ID])

	if g, ok := r.games[p.ID]; ok {
		g.Stop()
		delete(r.games, p.ID)
	}

	r.gamesMutex.Unlock()

	r.eventBus.RemoveChannel(fmt.Sprintf("player/%s", p.ID))

	delete(r.Players, p.ID)

	r.eventBus.Publish(event.New(bloccs.ChannelRoom, bloccs.EventPlayerLeave, &event.Payload{
		"player": p,
	}))
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
					if r.games[p.ID].Command(string(msg)) {
						if r.games[p.ID].Field.FallingPiece.Dirty {
							r.games[p.ID].PublishFallingPieceUpdate()
						}

						if r.games[p.ID].Field.Dirty {
							r.games[p.ID].PublishFieldUpdate()
						}
					}
				}

				r.gamesMutex.Unlock()
			}
		}(player)
	}

	r.playersMutex.Unlock()
}

func (r *Room) Stop() {
	r.gamesMutex.Lock()

	for _, g := range r.games {
		g.Stop()
	}
	r.gamesMutex.Unlock()

	r.roomWaitGroup.Wait()
}
