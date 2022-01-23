package bloccs

import (
	"bloccs-server/pkg/event"
	"fmt"
	"hash/fnv"
	"sync"
	"time"
)

type UpdateHandler func()

type Game struct {
	PlayerID string
	Field    *Field
	EventBus *event.Bus
	Over     bool
	// todo: refactor to score struct
	Score           int
	Lines           int
	stopChannel     chan bool
	globalWaitGroup *sync.WaitGroup
}

func NewGame(bus *event.Bus, roomId string, playerId string) *Game {
	h := fnv.New32a()

	_, _ = h.Write([]byte(roomId))

	game := &Game{
		PlayerID:        playerId,
		Field:           NewField(bus, NewRNG(int64(h.Sum32())), 10, 20, playerId),
		EventBus:        bus,
		Over:            false,
		Score:           0,
		Lines:           0,
		stopChannel:     make(chan bool),
		globalWaitGroup: &sync.WaitGroup{},
	}

	bus.AddChannel(fmt.Sprintf("update/%s", playerId))

	return game
}

func (g *Game) Start() {
	g.globalWaitGroup.Add(1)

	go func() {
		defer g.globalWaitGroup.Done()

		// 100 fps
		ticker := time.NewTicker(time.Millisecond * 10)

		defer ticker.Stop()

		for {
			select {
			case <-g.stopChannel:
				return
			case <-ticker.C:
				g.Update()
			}
		}
	}()
}

func (g *Game) Stop() {
	close(g.stopChannel)
	g.globalWaitGroup.Wait()

	g.EventBus.RemoveChannel(fmt.Sprintf("update/%s", g.PlayerID))
}

// todo: should actually be in field

func (g *Game) PublishFieldUpdate() {
	g.EventBus.Publish(event.New(fmt.Sprintf("update/%s", g.PlayerID), EventGameUpdate, &event.Payload{
		"field": g.Field,
		// todo: this is too late?
		"score": g.Score,
		"lines": g.Lines,
	}))
}

func (g *Game) PublishFallingPieceUpdate() {
	g.EventBus.Publish(event.New(fmt.Sprintf("update/%s", g.Field.PlayerID), EventUpdateFallingPiece, &event.Payload{
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

		// todo: g.Field should not be manipulated from here; maintain dirty flags from outside?; diffing?
		g.Field.Dirty = false
	}

	if gameOver {
		g.EventBus.Publish(event.New(fmt.Sprintf("update/%s", g.PlayerID), EventGameOver, nil))
	}

	g.Over = gameOver
}

// Command returns true if the command was understood
func (g *Game) Command(cmd string) bool {
	// todo: refactor command no not query g.Field twice

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
		g.Field.FallingPiece.HardDrop(g.Field)
		return true
	case "X":
		g.Field.FallingPiece.Move(g.Field, 0, 0, 1)
		return true
	case "H":
		g.Field.FallingPiece.HoldCurrentPiece(g.Field)
		return true
	default:
		return false
	}
}
