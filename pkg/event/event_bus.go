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

	winEvent.PublishedAt = time.Now().UnixMilli()

	b.connectionsMutex.RLock()
	defer b.connectionsMutex.RUnlock()

	now := time.Now().UnixMilli()

	for _, e := range events {
		e.SentAt = now
	}

	winEvent.SentAt = now

	sMsg, err := winEvent.Serialize()

	if err != nil {
		log.Printf("serialization error: %s, ignoring.\n", err)
		return
	}

	for _, conn := range b.connections {
		conn.Write(sMsg)
	}
}

func (b *Bus) startListener() {
	b.wg.Add(1)
	defer b.wg.Done()

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
	b.connectionsMutex.Lock()
	defer b.connectionsMutex.Unlock()

	b.connections[subscriberId] = conn
}

func (b *Bus) Unsubscribe(subscriberId string) {
	b.connectionsMutex.Lock()
	defer b.connectionsMutex.Unlock()

	if _, ok := b.connections[subscriberId]; ok {
		delete(b.connections, subscriberId)
	}
}

func (b *Bus) Publish(event *Event) {
	event.PublishedAt = time.Now().UnixMilli()
	b.channel <- event
}
