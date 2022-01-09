package server

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
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

func NewPlayer(name string, conn *websocket.Conn) *Player {
	return &Player{
		ID:        uuid.NewString(),
		Name:      name,
		CreateAt:  time.Now().UnixMilli(),
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
