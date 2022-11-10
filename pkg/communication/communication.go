package communication

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"sync"
	"time"
)

type PreStopCallback func()

type Connection struct {
	ws              *websocket.Conn
	ctx             context.Context
	stop            context.CancelFunc
	wg              *sync.WaitGroup
	readMutex       *sync.Mutex
	writeMutex      *sync.Mutex
	output          chan string
	input           chan string
	isStopping      bool
	preStopCallback PreStopCallback
}

type Settings struct {
	WS              *websocket.Conn
	ParentContext   context.Context
	PreStopCallback PreStopCallback
}

func NewConnection(settings *Settings) *Connection {
	ctx, cancel := context.WithCancel(settings.ParentContext)

	conn := Connection{
		ws:              settings.WS,
		ctx:             ctx,
		stop:            cancel,
		wg:              &sync.WaitGroup{},
		readMutex:       &sync.Mutex{},
		writeMutex:      &sync.Mutex{},
		output:          make(chan string, 256),
		input:           make(chan string, 256),
		isStopping:      false,
		preStopCallback: settings.PreStopCallback,
	}

	conn.ws.SetPongHandler(func(string) error {
		return conn.ws.SetReadDeadline(time.Now().Add(time.Second * 10))
	})

	conn.listen()

	return &conn
}

func (c *Connection) listen() {
	go c.startPings()
	go c.startWriter()
	go c.startReader()
}

func (c *Connection) Stop() {
	if c.isStopping {
		return
	}

	c.isStopping = true

	log.Println("stopping connection ...")

	if c.preStopCallback != nil {
		c.preStopCallback()
	}

	c.stop()
	_ = c.ws.Close()

	c.wg.Wait()

	log.Println("connection stopped")
}

func (c *Connection) startPings() {
	c.wg.Add(1)
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(time.Second * 5):
			if err := c.ping(); err != nil {
				go c.Stop()
				return
			}
			break
		}
	}
}

func (c *Connection) ping() error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	_ = c.ws.SetWriteDeadline(time.Now().Add(time.Second * 2))

	if err := c.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
		return err
	}

	return nil
}

func (c *Connection) startReader() {
	c.wg.Add(1)
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-time.After(time.Microsecond * 250):
			if msg, err := c.tryRead(); msg != "" && err == nil {
				// todo: this channel is flood-able when nobody listens
				c.input <- msg
			} else if err != nil {
				go c.Stop()
				return
			}

			break
		}
	}
}

func (c *Connection) tryRead() (string, error) {
	c.readMutex.Lock()
	defer c.readMutex.Unlock()

	_ = c.ws.SetReadDeadline(time.Now().Add(time.Second * 10))
	t, msg, err := c.ws.ReadMessage()

	if t == websocket.CloseMessage {
		return "", errors.New("websocket closed")
	} else if _, ok := err.(*websocket.CloseError); ok {
		return "", errors.New("websocket closed")
	}

	if err != nil {
		log.Printf("conn read error: %s.\n", err)
		return "", err
	}

	if t == websocket.TextMessage {
		return string(msg), nil
	}

	return "", nil
}

func (c *Connection) startWriter() {
	c.wg.Add(1)
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.output:
			if err := c.tryWrite(msg); err != nil {
				go c.Stop()
				return
			}

			break
		}
	}
}

func (c *Connection) tryWrite(msg string) error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	_ = c.ws.SetWriteDeadline(time.Now().Add(time.Second * 5))
	err := c.ws.WriteMessage(websocket.TextMessage, []byte(msg))

	if _, ok := err.(*websocket.CloseError); ok {
		return errors.New("websocket closed")
	}

	if err != nil {
		log.Printf("conn write error: %s.\n", err)
		return err
	}

	return nil
}

func (c *Connection) GetInputChannel() chan string {
	return c.input
}

// Read blocks until there is an unread message from the websocket
func (c *Connection) Read() string {
	return <-c.input
}

// Write enqueues the message to be sent to the websocket, blocks if too many messages are enqueued
func (c *Connection) Write(msg string) {
	c.output <- msg
}
