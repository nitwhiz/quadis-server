package bloccs

import (
	"sync"
	"time"
)

var nextUpdateHandlerID = 0

type UpdateHandler func()

type Game struct {
	Field           *Field
	EventBus        *EventBus
	Running         bool
	updateHandlers  map[int]UpdateHandler
	globalWaitGroup *sync.WaitGroup
}

func NewGame() *Game {
	bus := NewEventBus()

	return &Game{
		Field:           NewField(bus, 10, 20),
		EventBus:        bus,
		Running:         false,
		updateHandlers:  map[int]UpdateHandler{},
		globalWaitGroup: &sync.WaitGroup{},
	}
}

func (g *Game) AddUpdateHandler(u UpdateHandler) int {
	id := nextUpdateHandlerID

	g.updateHandlers[id] = u

	nextUpdateHandlerID++

	return id
}

func (g *Game) Start() {
	g.EventBus.Start()

	g.globalWaitGroup.Add(1)
	g.Running = true

	go func() {
		defer g.globalWaitGroup.Done()

		for {
			if !g.Running {
				return
			}

			g.Update()

			time.Sleep(time.Millisecond * 10)
		}
	}()
}

func (g *Game) Stop() {
	g.Running = false
	g.globalWaitGroup.Wait()
}

func (g *Game) Update() {
	g.Field.Update()

	handlerWaitGroup := &sync.WaitGroup{}

	for _, u := range g.updateHandlers {
		g.globalWaitGroup.Add(1)
		handlerWaitGroup.Add(1)

		go func(u UpdateHandler) {
			defer g.globalWaitGroup.Done()
			defer handlerWaitGroup.Done()

			u()
		}(u)
	}

	handlerWaitGroup.Wait()
}
