package server

import (
	"bloccs-server/pkg/bloccs"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type Player struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	CreateAt  int64           `json:"join_time"`
	Game      *bloccs.Game    `json:"-"`
	ConnMutex *sync.Mutex     `json:"-"`
	Conn      *websocket.Conn `json:"-"`
}

func NewPlayer(name string, conn *websocket.Conn) *Player {
	return &Player{
		ID:        uuid.NewString(),
		Name:      name,
		CreateAt:  time.Now().UnixMilli(),
		Game:      bloccs.NewGame(),
		ConnMutex: &sync.Mutex{},
		Conn:      conn,
	}
}

func (p *Player) SendMessage(data []byte) error {
	p.ConnMutex.Lock()
	err := p.Conn.WriteMessage(websocket.TextMessage, data)
	p.ConnMutex.Unlock()

	if err != nil {
		return err
	}

	return nil
}

func (p *Player) sendFieldUpdate() {
	bs, err := json.Marshal(Message{
		Event: &bloccs.Event{
			Type: bloccs.EventUpdate,
			Data: map[string]interface{}{
				"field": p.Game.Field,
			},
		},
	})

	if err != nil {
		log.Println("json failed:", err)
	}

	if err = p.SendMessage(bs); err != nil {
		log.Println("send failed:", err)
	}
}

func (p *Player) Listen() {
	for {
		_, message, err := p.Conn.ReadMessage()

		if err != nil {
			log.Println("read failed:", err)
			break
		}

		if string(message) == "L" {
			p.Game.Field.MoveFallingPiece(-1, 0, 0)

			p.sendFieldUpdate()
		} else if string(message) == "R" {
			p.Game.Field.MoveFallingPiece(1, 0, 0)

			p.sendFieldUpdate()
		} else if string(message) == "D" {
			p.Game.Field.MoveFallingPiece(0, 1, 0)

			p.sendFieldUpdate()
		} else if string(message) == "X" {
			p.Game.Field.MoveFallingPiece(0, 0, 1)

			p.sendFieldUpdate()
		}
	}
}
