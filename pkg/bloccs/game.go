package bloccs

import (
	"sync"
	"time"
)

type UpdateHandler func()

type Game struct {
	Field          *Field
	EventBus       *EventBus
	Running        bool
	UpdateHandlers []UpdateHandler
	waitGroup      *sync.WaitGroup
}

func NewGame() *Game {
	bus := NewEventBus()

	return &Game{
		Field:          NewField(bus, 10, 20),
		EventBus:       bus,
		Running:        false,
		UpdateHandlers: []UpdateHandler{},
		waitGroup:      &sync.WaitGroup{},
	}
}

func (g *Game) AddUpdateHandler(u UpdateHandler) {
	g.UpdateHandlers = append(g.UpdateHandlers, u)
}

func (g *Game) Start() {
	g.EventBus.Start()

	g.waitGroup.Add(1)
	g.Running = true

	go func() {
		defer g.waitGroup.Done()

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
	g.waitGroup.Wait()
}

func (g *Game) Update() {
	g.Field.Update()

	for _, u := range g.UpdateHandlers {
		g.waitGroup.Add(1)

		go func(u UpdateHandler) {
			defer g.waitGroup.Done()

			u()
		}(u)
	}
}
