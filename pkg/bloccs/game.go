package bloccs

import (
	"bloccs-server/pkg/event"
	"fmt"
	"sync"
	"time"
)

type UpdateHandler func()

type Game struct {
	ID              string
	Field           *Field
	EventBus        *event.Bus
	Over            bool
	stopChannel     chan bool
	globalWaitGroup *sync.WaitGroup
}

func NewGame(bus *event.Bus, id string) *Game {
	game := &Game{
		ID:              id,
		Field:           NewField(bus, 10, 20, id),
		EventBus:        bus,
		Over:            false,
		stopChannel:     make(chan bool),
		globalWaitGroup: &sync.WaitGroup{},
	}

	bus.AddChannel(fmt.Sprintf("game_update/%s", id))

	return game
}

func (g *Game) Start() {
	g.globalWaitGroup.Add(1)

	go func() {
		defer g.globalWaitGroup.Done()

		for {
			select {
			case <-g.stopChannel:
				return
			case <-time.After(time.Millisecond):
				break
			}

			g.Update()

			// 60 tps = 16 ms per frame = 15 ms idle + 1 ms stop wait
			time.Sleep(time.Millisecond * 15)
		}
	}()
}

func (g *Game) Stop() {
	close(g.stopChannel)
	g.globalWaitGroup.Wait()
}

// todo: should actually be in field

func (g *Game) PublishFieldUpdate() {
	g.EventBus.Publish(event.New(fmt.Sprintf("game_update/%s", g.ID), EventGameFieldUpdate, &event.Payload{
		"field": g.Field,
	}))
}

func (g *Game) PublishFallingPieceUpdate() {
	g.EventBus.Publish(event.New(fmt.Sprintf("game_update/%s", g.Field.ID), EventUpdateFallingPiece, &event.Payload{
		"falling_piece_data": g.Field.FallingPiece,
		"piece_display":      g.Field.FallingPiece.CurrentPiece.GetData(),
	}))
}

func (g *Game) Update() {
	if g.Over {
		return
	}

	dirty, gameOver := g.Field.Update()

	if dirty {
		g.PublishFieldUpdate()
	}

	if gameOver {
		g.EventBus.Publish(event.New(fmt.Sprintf("game_update/%s", g.ID), EventGameOver, nil))
	}

	g.Over = gameOver
}

// todo: refactor command no not query g.Field twice

// Command returns true if the command was understood
func (g *Game) Command(cmd string) bool {
	switch cmd {
	case "L":
		g.Field.FallingPiece.Move(g.Field, -1, 0, 0)
		return true
	case "R":
		g.Field.FallingPiece.Move(g.Field, 1, 0, 0)
		return true
	case "D":
		g.Field.FallingPiece.Move(g.Field, 0, 1, 0)
		return true
	case "P":
		g.Field.FallingPiece.Punch(g.Field)
		return true
	case "X":
		g.Field.FallingPiece.Move(g.Field, 0, 0, 1)
		return true
	default:
		return false
	}
}
