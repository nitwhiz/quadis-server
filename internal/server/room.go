package server

import (
	"bloccs-server/pkg/bloccs"
	"encoding/json"
	"github.com/google/uuid"
	"log"
	"sync"
)

type Room struct {
	ID             string             `json:"id"`
	Players        map[string]*Player `json:"players"`
	playersMutex   *sync.Mutex
	eventHandlers  map[string][]int
	updateHandlers map[string][]int
}

func NewRoom() *Room {
	return &Room{
		ID:             uuid.NewString(),
		Players:        map[string]*Player{},
		playersMutex:   &sync.Mutex{},
		eventHandlers:  map[string][]int{},
		updateHandlers: map[string][]int{},
	}
}

func getEventHandler(p *Player) bloccs.EventHandler {
	return func(event *bloccs.Event) {
		bs, err := json.Marshal(event)

		if err != nil {
			log.Println("json failed:", err)
		}

		if err = p.SendMessage(bs); err != nil {
			log.Println("send failed:", err)
		}
	}
}

func (r *Room) applyUpdateHandlers(p *Player) {
	// send p's events to p
	p.Game.AddUpdateHandler(func() {
		if !p.Game.Field.Dirty {
			return
		}

		bs, err := json.Marshal(&bloccs.Event{
			Type: bloccs.EventFieldUpdate,
			Data: map[string]interface{}{
				"player": p,
				"field":  p.Game.Field,
			},
		})

		if err != nil {
			log.Println("json failed:", err)
		}

		if err = p.SendMessage(bs); err != nil {
			log.Println("send failed:", err)
		}
	})

	for _, player := range r.Players {
		// send other player's events to p
		player.Game.AddUpdateHandler(func() {
			if !player.Game.Field.Dirty {
				return
			}

			bs, err := json.Marshal(&bloccs.Event{
				Type: bloccs.EventFieldUpdate,
				Data: map[string]interface{}{
					"player": player,
					"field":  player.Game.Field,
				},
			})

			if err != nil {
				log.Println("json failed:", err)
			}

			if err = p.SendMessage(bs); err != nil {
				log.Println("send failed:", err)
			}
		})

		// send p's updates to other players
		p.Game.AddUpdateHandler(func() {
			if !p.Game.Field.Dirty {
				return
			}

			bs, err := json.Marshal(&bloccs.Event{
				Type: bloccs.EventFieldUpdate,
				Data: map[string]interface{}{
					"player": p,
					"field":  p.Game.Field,
				},
			})

			if err != nil {
				log.Println("json failed:", err)
			}

			if err = player.SendMessage(bs); err != nil {
				log.Println("send failed:", err)
			}
		})
	}
}

func (r *Room) applyEventHandlers(p *Player) {
	// subscribe to their own game
	p.Game.EventBus.AddHandler(getEventHandler(p))

	for _, player := range r.Players {
		// subscribe to others' games
		player.Game.EventBus.AddHandler(getEventHandler(p))

		// subscribe others to their game
		p.Game.EventBus.AddHandler(getEventHandler(player))
	}
}

func (r *Room) AddPlayer(p *Player) {
	r.playersMutex.Lock()

	if _, ok := r.Players[p.ID]; ok {
		return
	}

	r.BroadcastEvent(&bloccs.Event{
		Type: bloccs.EventPlayerJoin,
		Data: map[string]interface{}{
			"player": p,
		},
	})

	r.applyUpdateHandlers(p)
	r.applyEventHandlers(p)

	r.Players[p.ID] = p

	r.playersMutex.Unlock()

	// todo: this is a concurrent read of players
	// todo: refactor mutex use in this project

	bs, _ := json.Marshal(&bloccs.Event{
		Type: bloccs.EventRoomInfo,
		Data: map[string]interface{}{
			"room": r,
		},
	})

	_ = p.SendMessage(bs)
}

func (r *Room) RemovePlayer(p *Player) {
	r.playersMutex.Lock()

	if _, ok := r.Players[p.ID]; !ok {
		return
	}

	p.Game.Stop()

	// todo: un-register event handlers

	delete(r.Players, p.ID)

	r.BroadcastEvent(&bloccs.Event{
		Type: bloccs.EventPlayerLeave,
		Data: map[string]interface{}{
			"player": p,
		},
	})

	r.playersMutex.Unlock()
}

func (r *Room) BroadcastEvent(event *bloccs.Event) {
	bs, err := json.Marshal(event)

	if err != nil {
		log.Printf("error in json for broadcast: %s\n", err.Error())
		return
	}

	for _, p := range r.Players {
		if err := p.SendMessage(bs); err != nil {
			log.Printf("error sending broadcast to %s: %s\n", p.ID, err.Error())
		}
	}
}

func (r *Room) Start() {
	r.playersMutex.Lock()

	for _, p := range r.Players {
		if p.Game != nil {
			p.Game.Start()
		}
	}

	r.playersMutex.Unlock()
}

func (r *Room) Stop() {
	r.playersMutex.Lock()

	for _, p := range r.Players {
		if p.Game != nil {
			p.Game.Stop()
		}
	}

	r.playersMutex.Unlock()
}
