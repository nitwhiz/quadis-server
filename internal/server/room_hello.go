package server

import (
	"bloccs-server/pkg/bloccs"
	"bloccs-server/pkg/event"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"time"
)

type HelloResponseMessage struct {
	Name string `json:"name"`
}

func (r *Room) Join(conn *websocket.Conn) error {
	p := NewPlayer(conn)

	err := r.handshakeHello(p)

	if err != nil {
		return err
	}

	// needed?
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(time.Second * 3))
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
				_ = conn.SetWriteDeadline(time.Now().Add(time.Second * 10))

				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}

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

func (r *Room) handshakeHello(p *Player) error {
	var err error
	var helloResponse *HelloResponseMessage

	go func() {
		helloResponse, err = r.listenForHello(p)

		if err != nil {
			return
		}

		p.Name = helloResponse.Name

		r.AddPlayer(p)

		r.playersMutex.Lock()

		bs, jsonErr := json.Marshal(event.New("none", bloccs.EventHelloAck, &event.Payload{
			"you":  p,
			"room": r,
		}))

		r.playersMutex.Unlock()

		if jsonErr != nil {
			log.Println("cannot marshal ack message")

			return
		}

		_ = p.SendMessage(bs)
	}()

	if err != nil {
		return err
	}

	bs, err := json.Marshal(event.New("none", bloccs.EventHello, nil))

	if err != nil {
		return err
	}

	return p.SendMessage(bs)
}
