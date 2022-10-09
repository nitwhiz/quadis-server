package event

import (
	"context"
	"sync"
	"time"
)

const eventTypeAll = "*"

type Handler func(event *Event)

type Bus struct {
	handlers      map[string][]Handler
	handlersMutex *sync.RWMutex
	wg            *sync.WaitGroup
	channel       chan *Event
	ctx           context.Context
	stop          context.CancelFunc
}

func NewBus() *Bus {
	ctx, cancel := context.WithCancel(context.Background())

	b := Bus{
		handlers:      map[string][]Handler{},
		handlersMutex: &sync.RWMutex{},
		wg:            &sync.WaitGroup{},
		channel:       make(chan *Event, 256),
		ctx:           ctx,
		stop:          cancel,
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
	b.handlersMutex.RLock()
	defer b.handlersMutex.RUnlock()

	for hType, handlers := range b.handlers {
		if event.Type == hType || hType == eventTypeAll {
			for _, handler := range handlers {
				handler(event)
			}
		}
	}

}

func (b *Bus) Stop() {
	b.stop()
	b.wg.Wait()
}

func (b *Bus) SubscribeAll(handler Handler) {
	// todo: unsubscribe based on some id
	b.Subscribe(eventTypeAll, handler)
}

func (b *Bus) Subscribe(eventType string, handler Handler) {
	b.handlersMutex.Lock()
	defer b.handlersMutex.Unlock()

	if _, ok := b.handlers[eventType]; !ok {
		b.handlers[eventType] = []Handler{}
	}

	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *Bus) Publish(event *Event) {
	b.channel <- event
}
