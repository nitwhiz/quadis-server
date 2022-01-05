package server

import (
	"bloccs-server/pkg/bloccs"
	"encoding/json"
	"github.com/google/uuid"
	"log"
)

type Room struct {
	ID             string             `json:"id"`
	Players        map[string]*Player `json:"players"`
	eventHandlers  map[string][]int
	updateHandlers map[string][]int
}

func NewRoom() *Room {
	return &Room{
		ID:             uuid.NewString(),
		Players:        map[string]*Player{},
		eventHandlers:  map[string][]int{},
		updateHandlers: map[string][]int{},
	}
}

func getEventHandler(p *Player) bloccs.EventHandler {
	return func(event *bloccs.Event) {
		bs, err := json.Marshal(Message{
			Event: event,
		})

		if err != nil {
			log.Println("json failed:", err)
		}

		if err = p.SendMessage(bs); err != nil {
			log.Println("send failed:", err)
		}
	}
}

func (r *Room) applyUpdateHandlers(p *Player) {
	for _, player := range r.Players {
		// send other player's events to p
		player.Game.AddUpdateHandler(func() {
			bs, err := json.Marshal(Message{
				Event: &bloccs.Event{
					Type: bloccs.EventUpdate,
					Data: map[string]interface{}{
						"field": player.Game.Field,
					},
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
			bs, err := json.Marshal(Message{
				Event: &bloccs.Event{
					Type: bloccs.EventUpdate,
					Data: map[string]interface{}{
						"field": player.Game.Field,
					},
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
	if _, ok := r.Players[p.ID]; ok {
		return
	}

	r.BroadcastMessage(&Message{
		Event: &bloccs.Event{
			Type: bloccs.EventPlayerJoin,
			Data: map[string]interface{}{
				"player": p,
			},
		},
	})

	r.applyUpdateHandlers(p)
	r.applyEventHandlers(p)

	r.Players[p.ID] = p

	bs, _ := json.Marshal(&Message{
		Event: &bloccs.Event{
			Type: bloccs.EventRoomInfo,
			Data: map[string]interface{}{
				"room": r,
			},
		},
	})

	_ = p.SendMessage(bs)
}

func (r *Room) RemovePlayer(p *Player) {
	if _, ok := r.Players[p.ID]; !ok {
		return
	}

	p.Game.Stop()

	// todo: un-register event handlers

	delete(r.Players, p.ID)

	r.BroadcastMessage(&Message{
		Event: &bloccs.Event{
			Type: bloccs.EventPlayerLeave,
			Data: map[string]interface{}{
				"player": p,
			},
		},
	})
}

func (r *Room) BroadcastMessage(msg *Message) {
	bs, err := json.Marshal(msg)

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
	for _, p := range r.Players {
		if p.Game != nil {
			p.Game.Start()
		}
	}
}

func (r *Room) Stop() {
	for _, p := range r.Players {
		if p.Game != nil {
			p.Game.Stop()
		}
	}
}
