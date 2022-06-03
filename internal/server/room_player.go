package server

import (
	"bloccs-server/pkg/bloccs"
	"bloccs-server/pkg/event"
	"log"
)

const EventPlayerJoin = "player_join"
const EventPlayerLeave = "player_leave"

// todo: could be done smarter maybe?
func isInGameEvent(eventType event.Type) bool {
	return eventType != EventHello &&
		eventType != EventHelloAck &&
		eventType != EventPlayerJoin &&
		eventType != EventPlayerLeave &&
		eventType != bloccs.EventGameStart
}

func (r *Room) sendEventToPlayer(e *event.Event, p *Player) error {
	if !r.gamesRunning && isInGameEvent(e.Type) {
		return nil
	}

	// todo: this data-races if event body is written during marshalling not locked

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

func (r *Room) RegisterPlayer(p *Player, g *bloccs.Game) {
	r.playersMutex.Lock()
	defer r.playersMutex.Unlock()

	r.gamesMutex.Lock()
	defer r.gamesMutex.Unlock()

	if _, ok := r.Players[p.GetId()]; ok {
		return
	}

	if _, ok := r.Games[g.GetId()]; ok {
		return
	}

	if len(r.Players) >= 7 {
		return
	}

	r.Players[p.GetId()] = p
	r.Games[g.GetId()] = g

	r.eventBus.Subscribe(event.All, func(event *event.Event) {
		if err := r.sendEventToPlayer(event, p); err != nil {
			log.Println("error passing event")

			r.RemoveGame(g)
		}
	})

	r.eventBus.Publish(event.New(EventPlayerJoin, p, nil))
}

func (r *Room) RemoveGame(g *bloccs.Game) {
	r.playersMutex.Lock()
	defer r.playersMutex.Unlock()

	r.gamesMutex.Lock()
	defer r.gamesMutex.Unlock()

	r.eventHandlerIdsMutex.Lock()
	defer r.eventHandlerIdsMutex.Unlock()

	id := g.GetId()

	if _, ok := r.Games[id]; ok {
		delete(r.Games, id)

		g.Stop()
	}

	if p, ok := r.Players[id]; ok {
		delete(r.Players, id)

		if hid, ok := r.eventHandlerIds[id]; ok {
			r.eventBus.Unsubscribe(hid)
			delete(r.eventHandlerIds, id)
		}

		r.eventBus.Publish(event.New(EventPlayerLeave, p, nil))
	}
}
