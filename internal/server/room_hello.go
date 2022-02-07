package server

import (
	"bloccs-server/pkg/event"
	"encoding/json"
	"log"
)

type HelloResponseMessage struct {
	Name string
}

func (r *Room) listenForHello(p *Player) (*HelloResponseMessage, error) {
	msg, err := p.ReadMessage()

	if err != nil {
		return nil, err
	}

	helloResponse := HelloResponseMessage{}

	err = json.Unmarshal(msg, &helloResponse)

	if err != nil || helloResponse.Name == "" {
		log.Println("handshakeHello: invalid hello response. listening again.", helloResponse)

		return r.listenForHello(p)
	}

	return &helloResponse, nil
}

func (r *Room) sendHello(p *Player) error {
	bs, err := json.Marshal(event.New("none", event.Hello, nil))

	if err != nil {
		return err
	}

	return p.SendMessage(bs)
}

func (r *Room) handshakeHello(p *Player) error {
	if err := r.sendHello(p); err != nil {
		return err
	}

	helloResponse, err := r.listenForHello(p)

	if err != nil {
		return err
	}

	p.Name = helloResponse.Name

	r.playersMutex.Lock()

	currPlayers := make([]event.PlayerPayload, 0)

	for _, p := range r.Players {
		currPlayers = append(currPlayers, event.PlayerPayload{
			ID:       p.ID,
			Name:     p.Name,
			CreateAt: p.CreateAt,
		})
	}

	r.playersMutex.Unlock()

	bs, err := json.Marshal(event.New("none", event.HelloAck, &event.HelloAckPayload{
		You: event.PlayerPayload{
			ID:       p.ID,
			Name:     p.Name,
			CreateAt: p.CreateAt,
		},
		Room: event.RoomPayload{
			ID:      r.ID,
			Players: currPlayers,
		},
	}))

	if err != nil {
		log.Println("cannot marshal ack message")

		return err
	}

	return p.SendMessage(bs)
}
