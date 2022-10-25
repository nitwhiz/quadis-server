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
	window           *Window
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

	b.window = NewWindow(ctx, b.windowClosedCallback)

	go b.startListener()

	return &b
}

func (b *Bus) windowClosedCallback(events []*Event) {
	winEvent := &Event{
		Type:   TypeWindow,
		Origin: OriginSystem(),
		Payload: &WindowPayload{
			Events: events,
		},
	}

	sMsg, err := winEvent.Serialize()

	if err != nil {
		log.Printf("serialization error: %s, ignoring.\n", err)
		return
	}

	defer b.connectionsMutex.RUnlock()
	b.connectionsMutex.RLock()

	for _, conn := range b.connections {
		conn.Write(sMsg)
	}
}

func (b *Bus) startListener() {
	defer b.wg.Done()
	b.wg.Add(1)

	for {
		select {
		case <-b.ctx.Done():
			return
		case event := <-b.channel:
			b.window.Add(event)

			break
		case <-time.After(time.Millisecond * 10):
			break
		}
	}
}

func (b *Bus) handleEvent(event *Event) {
	b.window.Add(event)
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
