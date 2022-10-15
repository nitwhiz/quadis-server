package event

import (
	"context"
	"sync"
	"time"
)

const eventTypeAll = "*"

type Handler func(event *Event)

type Bus struct {
	handlers      map[string]map[string][]Handler
	handlersMutex *sync.RWMutex
	wg            *sync.WaitGroup
	channel       chan *Event
	ctx           context.Context
	stop          context.CancelFunc
}

func NewBus(parentContext context.Context) *Bus {
	ctx, cancel := context.WithCancel(parentContext)

	b := Bus{
		handlers:      map[string]map[string][]Handler{},
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
	defer b.handlersMutex.RUnlock()
	b.handlersMutex.RLock()

	for hType, handlers := range b.handlers {
		if event.Type == hType || hType == eventTypeAll {
			for _, subscriberHandlers := range handlers {
				for _, handler := range subscriberHandlers {
					handler(event)
				}
			}
		}
	}

}

func (b *Bus) Stop() {
	b.stop()
	b.wg.Wait()
}

func (b *Bus) SubscribeAll(handler Handler, subscriberId string) {
	b.Subscribe(eventTypeAll, subscriberId, handler)
}

func (b *Bus) UnsubscribeAll(subscriberId string) {
	b.Unsubscribe(eventTypeAll, subscriberId)
}

func (b *Bus) Subscribe(eventType string, subscriberId string, handler Handler) {
	defer b.handlersMutex.Unlock()
	b.handlersMutex.Lock()

	if _, ok := b.handlers[eventType]; !ok {
		b.handlers[eventType] = map[string][]Handler{}
	}

	if _, ok := b.handlers[eventType][subscriberId]; !ok {
		b.handlers[eventType][subscriberId] = []Handler{}
	}

	b.handlers[eventType][subscriberId] = append(b.handlers[eventType][subscriberId], handler)
}

func (b *Bus) Unsubscribe(eventType string, subscriberId string) {
	defer b.handlersMutex.Unlock()
	b.handlersMutex.Lock()

	if _, ok := b.handlers[eventType]; ok {
		if _, ok := b.handlers[eventType][subscriberId]; ok {
			delete(b.handlers[eventType], subscriberId)
		}
	}
}

func (b *Bus) Publish(event *Event) {
	b.channel <- event
}
