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
	handlers          map[string][]Handler
	handlersMutex     *sync.Mutex
	running           bool
	waitGroup         *sync.WaitGroup
	exprToRex         map[string]*regexp.Regexp
	channels          map[string]chan *Event
	channelsMutex     *sync.Mutex
	stopChannels      map[string]chan bool
	stopChannelsMutex *sync.Mutex
	mainStopChannel   chan bool
}

func NewBus() *Bus {
	return &Bus{
		handlers:          map[string][]Handler{},
		handlersMutex:     &sync.Mutex{},
		running:           true,
		waitGroup:         &sync.WaitGroup{},
		exprToRex:         map[string]*regexp.Regexp{},
		channels:          map[string]chan *Event{},
		channelsMutex:     &sync.Mutex{},
		stopChannels:      map[string]chan bool{},
		stopChannelsMutex: &sync.Mutex{},
		mainStopChannel:   make(chan bool),
	}
}

func (b *Bus) AddChannel(name string) {
	c := make(chan *Event)

	b.channelsMutex.Lock()
	b.channels[name] = c
	b.channelsMutex.Unlock()

	b.stopChannelsMutex.Lock()
	b.stopChannels[name] = make(chan bool)
	b.stopChannelsMutex.Unlock()

	b.startChannelListener(name, c)
}

func (b *Bus) RemoveChannel(name string) {
	b.channelsMutex.Lock()
	delete(b.channels, name)
	b.channelsMutex.Unlock()

	b.stopChannelsMutex.Lock()
	b.stopChannels[name] <- true
	delete(b.stopChannels, name)
	b.stopChannelsMutex.Unlock()
}

func (b *Bus) startChannelListener(n string, c chan *Event) {
	go func(stopChannel chan bool) {
		defer b.waitGroup.Done()

		b.waitGroup.Add(1)

		for {
			if !b.running {
				break
			}

			select {
			case <-stopChannel:
				return
			case <-b.mainStopChannel:
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
	}(b.stopChannels[n])
}

func (b *Bus) Stop() {
	close(b.mainStopChannel)
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
