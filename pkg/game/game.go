package game

import (
	"context"
	"github.com/google/uuid"
	"github.com/nitwhiz/bloccs-server/pkg/event"
	"github.com/nitwhiz/bloccs-server/pkg/field"
	"github.com/nitwhiz/bloccs-server/pkg/piece"
	"github.com/nitwhiz/bloccs-server/pkg/rng"
	"github.com/nitwhiz/bloccs-server/pkg/score"
	"hash/fnv"
	"math"
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
	ClearedRowsBus  *event.Bus
	running         bool
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
		ClearedRowsBus:  event.NewBus(),
		running:         false,
	}

	game.ClearedRowsBus.AddChannel("none")

	game.nextFallingPiece()

	return game
}

func (g *Game) GetScore() (int, int) {
	g.Score.RLock()
	defer g.Score.RUnlock()

	return g.Score.Score, g.Score.Lines
}

func (g *Game) Start(gameOverHandler func()) {
	g.globalWaitGroup.Add(1)

	g.rpg.NextBag()

	now := time.Now()

	g.lastUpdate = &now

	ctx, cancel := context.WithCancel(context.Background())

	g.cancelTickLoop = cancel

	g.running = true

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
				g.Update(gameOverHandler)
			}
		}
	}()
}

func (g *Game) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()

	// todo: new rpg on base of previous seed

	g.FallingPiece = NewFallingPiece()
	g.HoldPiece = nil
	g.holdLock = false
	g.holdDirty = false
	g.NextPiece = nil
	g.nextDirty = false
	g.IsOver = false
	g.lastUpdate = nil
	g.cancelTickLoop = nil
	g.running = false

	// todo: wtf
	g.nextFallingPiece()
	g.nextFallingPiece()

	g.Score.Reset()
	g.Field.Reset()
}

func (g *Game) Stop() {
	if !g.running {
		return
	}

	g.running = false

	if g.cancelTickLoop != nil {
		g.cancelTickLoop()
		g.cancelTickLoop = nil
	}

	g.globalWaitGroup.Wait()

	//g.EventBus.RemoveChannel(fmt.Sprintf("update/%s", g.ID))
}

func (g *Game) Update(gameOverHandler func()) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.IsOver {
		return
	}

	now := time.Now()

	if g.lastUpdate != nil {
		oldBedrockLevel := g.Field.GetBedrock()
		clearedRows, gameOver := g.updateFallingPiece(int(time.Now().Sub(*g.lastUpdate)))

		if clearedRows > 0 {
			g.Score.AddLines(clearedRows)
			newBedrockLevel := g.Field.DecreaseBedrock(clearedRows)

			if newBedrockLevel == 0 {
				g.publishRowsCleared(clearedRows, int(math.Max(float64(clearedRows-oldBedrockLevel), 0)))
			} else {
				g.publishRowsCleared(clearedRows, 0)
			}
		}

		if g.Field.IsDirty() {
			g.publishFieldUpdate()
			g.Field.SetDirty(false)
		}

		if g.FallingPiece.IsDirty() {
			g.publishFallingPieceUpdate()
			g.FallingPiece.SetDirty(false)

			if !g.CanPutFallingPiece() {
				gameOver = true
			}
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
			go gameOverHandler()
		}
	}

	g.lastUpdate = &now
}

func (g *Game) CanPutFallingPiece() bool {
	if g.FallingPiece == nil {
		return false
	}

	return g.Field.CanPutPiece(g.FallingPiece.Piece, g.FallingPiece.Rotation, g.FallingPiece.X, g.FallingPiece.Y)
}
