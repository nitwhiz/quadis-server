package server

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type Player struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	CreateAt  int64           `json:"create_at"`
	ConnMutex *sync.Mutex     `json:"-"`
	Conn      *websocket.Conn `json:"-"`
}

func NewPlayer(conn *websocket.Conn) *Player {
	return &Player{
		ID:        uuid.NewString(),
		Name:      "",
		CreateAt:  time.Now().UnixMilli(),
		ConnMutex: &sync.Mutex{},
		Conn:      conn,
	}
}

func (p *Player) ReadMessage() ([]byte, error) {
	_ = p.Conn.SetReadDeadline(time.Now().Add(time.Second * 5))

	_, msg, err := p.Conn.ReadMessage()

	if err != nil {
		log.Println("player read error")
		return nil, err
	}

	return msg, nil
}

func (p *Player) SendMessage(data []byte) error {
	p.ConnMutex.Lock()
	defer p.ConnMutex.Unlock()

	_ = p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 5))

	err := p.Conn.WriteMessage(websocket.TextMessage, data)

	if err != nil {
		log.Println("player write error")
		return err
	}

	return nil
}
