package bloccs

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const EventRowsCleared = "player_game_rows_cleared"
const EventUpdate = "player_game_update"
const EventPlayerJoin = "player_join"
const EventPlayerLeave = "player_leave"
const EventRoomInfo = "room_info"
const EventGameOver = "player_game_over"

var nextEventID = 0

type Event struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

type EventHandler func(event *Event)

type EventBus struct {
	queue      []*Event
	queueMutex *sync.Mutex
	handlers   map[int]EventHandler
	running    bool
	waitGroup  *sync.WaitGroup
}

func NewEventBus() *EventBus {
	return &EventBus{
		queue:      []*Event{},
		handlers:   map[int]EventHandler{},
		queueMutex: &sync.Mutex{},
		running:    false,
		waitGroup:  &sync.WaitGroup{},
	}
}

func (e *EventBus) Start() {
	e.waitGroup.Add(1)
	e.running = true

	go func() {
		defer e.waitGroup.Done()

		for {
			if !e.running {
				break
			}

			e.Tick()

			time.Sleep(time.Millisecond * 10)
		}
	}()
}

func (e *EventBus) Stop() {
	e.running = false
	e.waitGroup.Wait()
}

func (e *EventBus) AddHandler(handler EventHandler) int {
	id := nextEventID

	e.handlers[id] = handler

	nextEventID++

	return id
}

func (e *EventBus) RemoveHandler(id int) error {
	if _, ok := e.handlers[id]; ok {
		delete(e.handlers, id)
	}

	return errors.New(fmt.Sprintf("no event handler with id %d found", id))
}

func (e *EventBus) Publish(event *Event) {
	e.queueMutex.Lock()
	e.queue = append(e.queue, event)
	e.queueMutex.Unlock()
}

func (e *EventBus) Tick() {
	e.queueMutex.Lock()

	for _, event := range e.queue {
		// todo(enhancement): drop event if it's too old

		for _, handler := range e.handlers {
			e.waitGroup.Add(1)

			go func(h EventHandler) {
				defer e.waitGroup.Done()

				h(event)
			}(handler)
		}
	}

	e.queue = make([]*Event, 0)

	e.queueMutex.Unlock()
}
