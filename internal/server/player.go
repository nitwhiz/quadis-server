package server

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/nitwhiz/bloccs-server/pkg/game"
	"log"
	"sync"
	"time"
)

type Player struct {
	ID                  string
	Name                string
	CreateAt            int64
	game                *game.Game
	Conn                *websocket.Conn
	connWriteMutex      *sync.Mutex
	connReadMutex       *sync.Mutex
	Disconnect          chan bool
	listenLoopWaitGroup *sync.WaitGroup
	cancelListenLoop    context.CancelFunc
}

type StopCallbackFunc func()

func NewPlayer(conn *websocket.Conn, game *game.Game) *Player {
	return &Player{
		ID:                  game.ID,
		Name:                "",
		CreateAt:            time.Now().UnixMilli(),
		Conn:                conn,
		game:                game,
		connWriteMutex:      &sync.Mutex{},
		connReadMutex:       &sync.Mutex{},
		Disconnect:          make(chan bool),
		listenLoopWaitGroup: &sync.WaitGroup{},
		cancelListenLoop:    nil,
	}
}

func (p *Player) ResetGame() {
	p.game.Reset()
}

func (p *Player) StopGame() {
	p.game.Stop()
}

func (p *Player) StartGame(gameOverHandler func()) {
	p.game.Start(gameOverHandler)
}

func (p *Player) Listen(stopCallback StopCallbackFunc) {
	p.Conn.SetPongHandler(func(string) error {
		return p.Conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	})

	ctx, cancel := context.WithCancel(context.Background())

	p.cancelListenLoop = cancel

	go func() {
		defer func(Conn *websocket.Conn) {
			p.listenLoopWaitGroup.Done()

			if p.cancelListenLoop != nil {
				p.cancelListenLoop()
			}

			p.StopGame()

			p.listenLoopWaitGroup.Wait()

			_ = Conn.Close()
			stopCallback()
		}(p.Conn)

		p.listenLoopWaitGroup.Add(1)

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Millisecond):
				break
			}

			p.connReadMutex.Lock()

			_ = p.Conn.SetReadDeadline(time.Now().Add(time.Second * 10))
			t, msg, err := p.Conn.ReadMessage()

			p.connReadMutex.Unlock()

			if t == websocket.CloseMessage {
				return
			} else if _, ok := err.(*websocket.CloseError); ok {
				return
			}

			if err != nil {
				continue
			}

			if t == websocket.TextMessage {
				p.game.Command(string(msg))
			}
		}
	}()

	go func() {
		pingTicker := time.NewTicker(time.Second * 5)

		defer func() {
			pingTicker.Stop()

			if p.cancelListenLoop != nil {
				p.cancelListenLoop()
			}

			p.listenLoopWaitGroup.Done()
		}()

		p.listenLoopWaitGroup.Add(1)

		for {
			select {
			case <-ctx.Done():
				return
			case <-pingTicker.C:
				if err := p.Ping(); err != nil {
					return
				}
				break
			}
		}
	}()
}

func (p *Player) Ping() error {
	p.connWriteMutex.Lock()
	defer p.connWriteMutex.Unlock()

	_ = p.Conn.SetWriteDeadline(time.Now().Add(time.Second * 2))

	if err := p.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		return err
	}

	return nil
}

func (p *Player) ReadMessage() ([]byte, error) {
	p.connReadMutex.Lock()
	defer p.connReadMutex.Unlock()

	_ = p.Conn.SetReadDeadline(time.Now().Add(time.Second * 10))

	_, msg, err := p.Conn.ReadMessage()

	if err != nil {
		log.Println("player read error", err)
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
		log.Println("player write error", err)
		return err
	}

	return nil
}
