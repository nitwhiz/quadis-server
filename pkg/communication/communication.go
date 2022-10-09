package communication

import (
	"context"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type Connection struct {
	ws         *websocket.Conn
	ctx        context.Context
	stop       context.CancelFunc
	wg         *sync.WaitGroup
	readMutex  *sync.Mutex
	writeMutex *sync.Mutex
	output     chan string
	input      chan string
}

func NewConnection(ws *websocket.Conn) *Connection {
	ctx, cancel := context.WithCancel(context.Background())

	conn := Connection{
		ws:         ws,
		ctx:        ctx,
		stop:       cancel,
		wg:         &sync.WaitGroup{},
		readMutex:  &sync.Mutex{},
		writeMutex: &sync.Mutex{},
		output:     make(chan string, 256),
		input:      make(chan string, 256),
	}

	// todo: disconnected connection never aborts; never sends either
	conn.ws.SetPongHandler(func(string) error {
		return conn.ws.SetReadDeadline(time.Now().Add(time.Second * 10))
	})

	conn.listen()

	return &conn
}

func (c *Connection) listen() {
	c.startPings()

	c.startWriter()
	c.startReader()
}

func (c *Connection) Stop() {
	log.Println("stopping connection")

	c.stop()
	c.wg.Wait()
}

func (c *Connection) startPings() {
	go func() {
		pingTicker := time.NewTicker(time.Second * 5)
		defer pingTicker.Stop()

		defer c.wg.Done()
		c.wg.Add(1)

		defer func(ws *websocket.Conn) {
			_ = ws.Close()
		}(c.ws)

		for {
			select {
			case <-c.ctx.Done():
				return
			case <-pingTicker.C:
				if err := c.ping(); err != nil {
					return
				}
				break
			}
		}
	}()
}

func (c *Connection) ping() error {
	defer c.writeMutex.Unlock()
	c.writeMutex.Lock()

	_ = c.ws.SetWriteDeadline(time.Now().Add(time.Second * 2))

	if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
		return err
	}

	return nil
}

func (c *Connection) startReader() {
	go func() {
		defer c.wg.Done()
		c.wg.Add(1)

		for {
			select {
			case <-c.ctx.Done():
				return
			case <-time.After(time.Millisecond):
				break
			}

			if msg := c.tryRead(); msg != "" {
				c.input <- msg
			}
		}
	}()
}

func (c *Connection) tryRead() string {
	defer c.readMutex.Unlock()
	c.readMutex.Lock()

	_ = c.ws.SetReadDeadline(time.Now().Add(time.Second * 10))
	t, msg, err := c.ws.ReadMessage()

	if t == websocket.CloseMessage {
		return ""
	} else if _, ok := err.(*websocket.CloseError); ok {
		c.Stop()

		return ""
	}

	if err != nil {
		log.Printf("conn read error: %s, not closing connection.\n", err)
		return ""
	}

	if t == websocket.TextMessage {
		return string(msg)
	}

	return ""
}

func (c *Connection) startWriter() {
	go func() {
		defer c.wg.Done()
		c.wg.Add(1)

		for {
			select {
			case <-c.ctx.Done():
				return
			case msg := <-c.output:
				c.tryWrite(msg)
				break
			}
		}
	}()
}

func (c *Connection) tryWrite(msg string) {
	defer c.writeMutex.Unlock()
	c.writeMutex.Lock()

	_ = c.ws.SetWriteDeadline(time.Now().Add(time.Second * 5))
	err := c.ws.WriteMessage(websocket.TextMessage, []byte(msg))

	if _, ok := err.(*websocket.CloseError); ok {
		c.Stop()
		return
	}

	if err != nil {
		log.Printf("conn write error: %s, not closing connection.\n", err)
	}
}

func (c *Connection) GetInputChannel() chan string {
	return c.input
}

func (c *Connection) Read() string {
	return <-c.input
}

func (c *Connection) Write(msg string) {
	c.output <- msg
}
