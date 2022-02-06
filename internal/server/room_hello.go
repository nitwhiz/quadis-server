package server

import (
	"bloccs-server/pkg/event"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

type HelloResponseMessage struct {
	Name string
}

func (r *Room) Join(conn *websocket.Conn) error {
	p := NewPlayer(conn)

	if err := r.handshakeHello(p); err != nil {
		return err
	}

	r.AddPlayer(p)

	p.Conn.SetPongHandler(func(string) error {
		_ = p.Conn.SetReadDeadline(time.Now().Add(time.Second * 3))
		return nil
	})

	go func() {
		pingTicker := time.NewTicker(time.Second * 2)

		defer func() {
			log.Println("removing player")

			pingTicker.Stop()
			r.RemovePlayer(p)
		}()

		for {
			select {
			case <-pingTicker.C:
				p.Ping()
				break
			}
		}
	}()

	return nil
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

	var currPlayers []event.PlayerPayload

	for _, p := range r.Players {
		currPlayers = append(currPlayers, event.PlayerPayload{
			ID:       p.ID,
			Name:     p.Name,
			CreateAt: p.CreateAt,
		})
	}

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

	r.playersMutex.Unlock()

	if err != nil {
		log.Println("cannot marshal ack message")

		return err
	}

	return p.SendMessage(bs)
}
