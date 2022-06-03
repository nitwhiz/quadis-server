package event

import (
	"sync"
)

const All = "*"

type Handler func(event *Event)

type Bus struct {
	nextHandlerId int
	handlers      map[Type]map[int]Handler
	handlersMutex *sync.RWMutex
	waitGroup     *sync.WaitGroup
	channel       chan *Event
	stopChannel   chan bool
}

func NewBus() *Bus {
	return &Bus{
		nextHandlerId: 0,
		handlers:      map[Type]map[int]Handler{},
		handlersMutex: &sync.RWMutex{},
		waitGroup:     &sync.WaitGroup{},
		channel:       make(chan *Event),
		stopChannel:   make(chan bool),
	}
}

func (b *Bus) Stop() {
	b.stopChannel <- true
	close(b.stopChannel)

	b.waitGroup.Wait()
}

func (b *Bus) Subscribe(eventType Type, handler Handler) int {
	b.handlersMutex.Lock()
	defer b.handlersMutex.Unlock()

	if _, ok := b.handlers[eventType]; !ok {
		b.handlers[eventType] = map[int]Handler{}
	}

	handlerId := b.nextHandlerId

	b.handlers[eventType][handlerId] = handler

	b.nextHandlerId += 1

	return handlerId
}

func (b *Bus) Unsubscribe(handlerId int) {
	b.handlersMutex.Lock()
	defer b.handlersMutex.Unlock()

	for eventType := range b.handlers {
		if _, ok := b.handlers[eventType][handlerId]; ok {
			delete(b.handlers[eventType], handlerId)
			break
		}
	}
}

func (b *Bus) Publish(event *Event) {
	b.channel <- event
}

func (b *Bus) handleEvent(event *Event) {
	defer b.handlersMutex.RUnlock()
	b.handlersMutex.RLock()

	hs, ok := b.handlers[event.Type]

	if ok {
		for _, h := range hs {
			h(event)
		}
	}

	hs, ok = b.handlers[All]

	if ok {
		for _, h := range hs {
			h(event)
		}
	}
}

func (b *Bus) Start() {
	go func() {
		defer b.waitGroup.Done()
		b.waitGroup.Add(1)

		for {
			select {
			case <-b.stopChannel:
				return
			case event := <-b.channel:
				b.handleEvent(event)
			}
		}
	}()
}
