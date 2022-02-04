package server

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type Player struct {
	ID             string
	Name           string
	CreateAt       int64
	Conn           *websocket.Conn
	connWriteMutex *sync.Mutex
	connReadMutex  *sync.Mutex
}

func NewPlayer(conn *websocket.Conn) *Player {
	return &Player{
		ID:             uuid.NewString(),
		Name:           "",
		CreateAt:       time.Now().UnixMilli(),
		Conn:           conn,
		connWriteMutex: &sync.Mutex{},
		connReadMutex:  &sync.Mutex{},
	}
}

func (p *Player) Ping() {
	p.connWriteMutex.Lock()
	defer p.connWriteMutex.Unlock()

	_ = p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 10))

	if err := p.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		return
	}
}

func (p *Player) ReadMessage() ([]byte, error) {
	p.connReadMutex.Lock()
	defer p.connReadMutex.Unlock()

	_ = p.Conn.SetReadDeadline(time.Now().Add(time.Second * 5))

	_, msg, err := p.Conn.ReadMessage()

	if err != nil {
		log.Println("player read error")
		return nil, err
	}

	return msg, nil
}

func (p *Player) SendMessage(data []byte) error {
	p.connWriteMutex.Lock()
	defer p.connWriteMutex.Unlock()

	_ = p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 5))

	err := p.Conn.WriteMessage(websocket.TextMessage, data)

	if err != nil {
		log.Println("player write error")
		return err
	}

	return nil
}
