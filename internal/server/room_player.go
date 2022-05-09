package server

import (
	"bloccs-server/pkg/event"
	"log"
)

func (r *Room) passEvent(p *Player, e *event.Event) error {
	// todo: throttle events from other players; only send specific events instantly
	// todo: this data-races if event body is something not locked - may be irrelevant though

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
	defer r.playersMutex.Unlock()

	if _, ok := r.Players[p.ID]; ok {
		return
	}

	if len(r.Players) >= 7 {
		return
	}

	r.Players[p.ID] = p

	r.eventBus.Subscribe(event.ChannelRoom, func(e *event.Event) {
		if err := r.passEvent(p, e); err != nil {
			log.Println("error passing event")
			r.RemovePlayer(p)
		}
	}, p.ID)

	r.eventBus.Subscribe("update/.*", func(e *event.Event) {
		if e.Type == event.GameOver {

		}

		if err := r.passEvent(p, e); err != nil {
			log.Println("error passing event")
			r.RemovePlayer(p)
		}
	}, p.ID)

	r.eventBus.Publish(event.New(event.ChannelRoom, event.PlayerJoin, &event.PlayerJoinPayload{
		ID:       p.ID,
		Name:     p.Name,
		CreateAt: p.CreateAt,
	}))
}

func (r *Room) RemovePlayer(p *Player) {
	r.playersMutex.Lock()
	defer r.playersMutex.Unlock()

	if _, ok := r.Players[p.ID]; !ok {
		return
	}

	r.eventBus.Unsubscribe(p.ID)

	delete(r.Players, p.ID)

	r.eventBus.Publish(event.New(event.ChannelRoom, event.PlayerLeave, &event.PlayerLeavePayload{
		ID:       p.ID,
		Name:     p.Name,
		CreateAt: p.CreateAt,
	}))
}
