package game

import (
	"bloccs-server/pkg/event"
	"bloccs-server/pkg/field"
	"bloccs-server/pkg/piece"
	"bloccs-server/pkg/rng"
	"bloccs-server/pkg/score"
	"context"
	"fmt"
	"github.com/google/uuid"
	"hash/fnv"
	"sync"
	"time"
)

type Game struct {
	ID              string
	FallingPiece    *FallingPiece
	HoldPiece       *piece.Piece
	holdLock        bool
	holdDirty       bool
	NextPiece       *piece.Piece
	nextDirty       bool
	Field           *field.Field
	EventBus        *event.Bus
	IsOver          bool
	Score           *score.Score
	rpg             *rng.RPG
	lastUpdate      *time.Time
	globalWaitGroup *sync.WaitGroup
	mu              *sync.RWMutex
	cancelTickLoop  context.CancelFunc
}

func New(bus *event.Bus, rngSeed string) *Game {
	h := fnv.New32a()

	_, _ = h.Write([]byte(rngSeed))

	rpg := rng.NewRPG(int64(h.Sum32()))

	game := &Game{
		ID:              uuid.NewString(),
		FallingPiece:    NewFallingPiece(),
		HoldPiece:       nil,
		holdLock:        false,
		holdDirty:       false,
		NextPiece:       nil,
		nextDirty:       false,
		Field:           field.New(bus, 10, 20),
		EventBus:        bus,
		IsOver:          false,
		Score:           score.New(),
		rpg:             rpg,
		lastUpdate:      nil,
		globalWaitGroup: &sync.WaitGroup{},
		mu:              &sync.RWMutex{},
		cancelTickLoop:  nil,
	}

	game.nextFallingPiece()

	bus.AddChannel(fmt.Sprintf("update/%s", game.ID))

	return game
}

func (g *Game) GetScore() (int, int) {
	g.Score.RLock()
	defer g.Score.RUnlock()

	return g.Score.Score, g.Score.Lines
}

func (g *Game) Start() {
	g.globalWaitGroup.Add(1)

	g.rpg.NextBag()

	now := time.Now()

	g.lastUpdate = &now

	ctx, cancel := context.WithCancel(context.Background())

	g.cancelTickLoop = cancel

	go func() {
		defer g.globalWaitGroup.Done()

		// 100 fps
		ticker := time.NewTicker(time.Millisecond * 10)

		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				g.Update()
			}
		}
	}()
}

func (g *Game) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.HoldPiece = nil
	g.holdLock = false
	g.holdDirty = false
	g.NextPiece = nil
	g.nextDirty = false
	g.IsOver = false
	g.lastUpdate = nil
	g.cancelTickLoop = nil

	g.Score.Reset()
	g.Field.Reset()
}

func (g *Game) Stop() {
	if g.cancelTickLoop != nil {
		g.cancelTickLoop()
	}

	g.globalWaitGroup.Wait()

	g.EventBus.RemoveChannel(fmt.Sprintf("update/%s", g.ID))
}

func (g *Game) Update() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.IsOver {
		return
	}

	now := time.Now()

	if g.lastUpdate != nil {
		clearedRows, gameOver := g.updateFallingPiece(int(time.Now().Sub(*g.lastUpdate)))

		g.Score.AddLines(clearedRows)
		g.Field.DecreaseBedrock(clearedRows)

		if g.Field.IsDirty() {
			g.publishFieldUpdate()
			g.Field.SetDirty(false)
		}

		if g.FallingPiece.IsDirty() {
			g.publishFallingPieceUpdate()
			g.FallingPiece.SetDirty(false)
		}

		if g.nextDirty {
			g.publishNextPieceUpdate()
			g.nextDirty = false
		}

		if g.holdDirty {
			g.publishHoldPieceUpdate()
			g.holdDirty = false
		}

		if g.Score.IsDirty() {
			g.publishScoreUpdate()
			g.Score.SetDirty(false)
		}

		if gameOver {
			g.IsOver = true
			g.publishGameOver()
		}
	}

	g.lastUpdate = &now
}
