package server

import (
	"bloccs-server/pkg/event"
	"bloccs-server/pkg/game"
	"log"
)

func (r *Room) passEvent(p *Player, e *event.Event) error {
	// todo: throttle events from other players; only send specific events instantly

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

	g := game.New(r.eventBus, r.ID, p.ID)

	r.gamesMutex.Lock()
	r.games[p.ID] = g
	r.gamesMutex.Unlock()

	r.Players[p.ID] = p

	r.eventBus.Subscribe(event.ChannelRoom, func(e *event.Event) {
		if err := r.passEvent(p, e); err != nil {
			log.Println("error passing event")
		}
	}, g)

	r.eventBus.Subscribe("update/.*", func(e *event.Event) {
		if err := r.passEvent(p, e); err != nil {
			log.Println("error passing event")
		}
	}, g)

	r.playersMutex.Unlock()

	r.eventBus.Publish(event.New(event.ChannelRoom, event.PlayerJoin, &event.PlayerJoinPayload{
		ID:       p.ID,
		Name:     p.Name,
		CreateAt: p.CreateAt,
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

	delete(r.Players, p.ID)

	r.eventBus.Publish(event.New(event.ChannelRoom, event.PlayerLeave, &event.PlayerLeavePayload{
		ID:       p.ID,
		Name:     p.Name,
		CreateAt: p.CreateAt,
	}))
}
