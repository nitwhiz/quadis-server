package event

import (
	"log"
	"regexp"
	"sync"
	"time"
)

const ChanBroadcast = "*"
const ChanExprAll = ".*"

type Handler func(event *Event)

type Bus struct {
	handlers      map[string][]Handler
	handlersMutex *sync.Mutex
	running       bool
	waitGroup     *sync.WaitGroup
	exprToRex     map[string]*regexp.Regexp
	channels      map[string]chan *Event
	stopChannel   chan bool
}

func NewBus() *Bus {
	return &Bus{
		handlers:      map[string][]Handler{},
		handlersMutex: &sync.Mutex{},
		running:       true,
		waitGroup:     &sync.WaitGroup{},
		exprToRex:     map[string]*regexp.Regexp{},
		channels:      map[string]chan *Event{},
		stopChannel:   make(chan bool),
	}
}

func (b *Bus) AddChannel(name string) {
	c := make(chan *Event)

	b.channels[name] = c

	b.startChannelListener(name, c)
}

func (b *Bus) startChannelListener(n string, c chan *Event) {
	go func() {
		defer b.waitGroup.Done()
		b.waitGroup.Add(1)

		for {
			if !b.running {
				break
			}

			select {
			case <-b.stopChannel:
				return
			case event := <-c:
				go func() {
					defer b.waitGroup.Done()
					b.waitGroup.Add(1)

					b.handlersMutex.Lock()

					for expr, handlers := range b.handlers {
						rex := b.exprToRex[expr]

						if rex.MatchString(n) {
							for _, h := range handlers {
								h(event)
							}
						}
					}

					b.handlersMutex.Unlock()
				}()

				break
			case <-time.After(time.Millisecond * 10):
				break
			}
		}
	}()
}

func (b *Bus) Stop() {
	close(b.stopChannel)
	b.waitGroup.Wait()
}

func (b *Bus) Subscribe(channelExpr string, handler Handler) {
	if _, ok := b.exprToRex[channelExpr]; !ok {
		rex, _ := regexp.Compile(channelExpr)

		b.exprToRex[channelExpr] = rex
	}

	b.handlersMutex.Lock()

	if _, ok := b.handlers[channelExpr]; !ok {
		b.handlers[channelExpr] = []Handler{}
	}

	b.handlers[channelExpr] = append(b.handlers[channelExpr], handler)

	b.handlersMutex.Unlock()
}

func (b *Bus) Publish(event *Event) {
	if event.Channel == ChanBroadcast {
		for n, c := range b.channels {
			exactEvent := &Event{
				Channel: n,
				Type:    event.Type,
				Payload: event.Payload,
			}

			c <- exactEvent
		}
	} else if _, ok := b.channels[event.Channel]; ok {
		b.channels[event.Channel] <- event
	} else {
		log.Printf("channel '%s' not found!", event.Channel)
	}

}
