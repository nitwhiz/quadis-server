package server

import (
	"bloccs-server/pkg/event"
	"encoding/json"
	"errors"
	"fmt"
)

const EventHello = "hello"
const EventHelloAck = "hello_ack"

type HelloResponseMessage struct {
	Name string `json:"name"`
}

type HelloAckPayload struct {
	Player *Player `json:"player"`
}

func (r *Room) listenForHello(p *Player) (*HelloResponseMessage, error) {
	msg, err := p.ReadMessage()

	if err != nil {
		return nil, err
	}

	helloResponse := HelloResponseMessage{}

	err = json.Unmarshal(msg, &helloResponse)

	if err != nil || helloResponse.Name == "" {
		return nil, errors.New("invalid hello response")
	}

	return &helloResponse, nil
}

func (r *Room) sendHello(p *Player) error {
	fmt.Println("sending hello")
	return r.sendEventToPlayer(event.New(EventHello, nil, nil), p)
}

func (r *Room) handshakeHello(p *Player) error {
	fmt.Println("handshaking")

	if err := r.sendHello(p); err != nil {
		return err
	}

	helloResponse, err := r.listenForHello(p)

	if err != nil {
		return err
	}

	p.Name = helloResponse.Name

	return r.sendEventToPlayer(event.New(EventHelloAck, r, &HelloAckPayload{Player: p}), p)
}
