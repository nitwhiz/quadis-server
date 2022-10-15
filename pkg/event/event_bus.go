package event

import (
	"context"
	"github.com/nitwhiz/quadis-server/pkg/communication"
	"log"
	"sync"
	"time"
)

type Handler func(event *Event)

type Bus struct {
	connections      map[string]*communication.Connection
	connectionsMutex *sync.RWMutex
	wg               *sync.WaitGroup
	channel          chan *Event
	ctx              context.Context
	stop             context.CancelFunc
}

// NewBus returns an event bus made for broadcasting events to websocket connections
func NewBus(parentContext context.Context) *Bus {
	ctx, cancel := context.WithCancel(parentContext)

	b := Bus{
		connections:      map[string]*communication.Connection{},
		connectionsMutex: &sync.RWMutex{},
		wg:               &sync.WaitGroup{},
		channel:          make(chan *Event, 256),
		ctx:              ctx,
		stop:             cancel,
	}

	b.startListener()

	return &b
}

func (b *Bus) startListener() {
	go func() {
		defer b.wg.Done()
		b.wg.Add(1)

		for {
			select {
			case <-b.ctx.Done():
				return
			case event := <-b.channel:
				go func() {
					defer b.wg.Done()
					b.wg.Add(1)

					b.handleEvent(event)
				}()

				break
			case <-time.After(time.Millisecond * 10):
				break
			}
		}
	}()
}

func (b *Bus) handleEvent(event *Event) {
	defer b.connectionsMutex.RUnlock()
	b.connectionsMutex.RLock()

	sMsg, err := event.Serialize()

	if err != nil {
		log.Printf("serialization error: %s, ignoring.\n", err)
		return
	}

	for _, conn := range b.connections {
		conn.Write(sMsg)
	}
}

func (b *Bus) Stop() {
	b.stop()
	b.wg.Wait()
}

func (b *Bus) Subscribe(subscriberId string, conn *communication.Connection) {
	defer b.connectionsMutex.Unlock()
	b.connectionsMutex.Lock()

	b.connections[subscriberId] = conn
}

func (b *Bus) Unsubscribe(subscriberId string) {
	defer b.connectionsMutex.Unlock()
	b.connectionsMutex.Lock()

	if _, ok := b.connections[subscriberId]; ok {
		delete(b.connections, subscriberId)
	}
}

func (b *Bus) Publish(event *Event) {
	b.channel <- event
}
