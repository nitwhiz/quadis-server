package game

import (
	"context"
	"github.com/google/uuid"
	"github.com/nitwhiz/quadis-server/pkg/communication"
	"github.com/nitwhiz/quadis-server/pkg/event"
	"github.com/nitwhiz/quadis-server/pkg/falling_piece"
	"github.com/nitwhiz/quadis-server/pkg/field"
	"github.com/nitwhiz/quadis-server/pkg/piece"
	"github.com/nitwhiz/quadis-server/pkg/player"
	"github.com/nitwhiz/quadis-server/pkg/rng"
	"github.com/nitwhiz/quadis-server/pkg/score"
	"log"
	"math"
	"sync"
	"time"
)

type Game struct {
	id           string
	player       *player.Player
	fallingPiece *falling_piece.FallingPiece
	nextPiece    *piece.LivingPiece
	holdingPiece *piece.LivingPiece
	field        *field.Field
	bus          *event.Bus
	over         bool
	score        *score.Score
	lastUpdate   *int64
	running      bool
	rpg          *rng.Piece
	ctx          context.Context
	stop         context.CancelFunc
	wg           *sync.WaitGroup
	mu           *sync.RWMutex
	comm         *communication.Connection
}

type Payload struct {
	Id         string `json:"id"`
	PlayerName string `json:"playerName"`
}

func New(bus *event.Bus, c *communication.Connection, player *player.Player) *Game {
	f := field.New()
	s := score.New()

	g := Game{
		id:           uuid.NewString(),
		player:       player,
		fallingPiece: nil,
		nextPiece:    nil,
		holdingPiece: nil,
		field:        f,
		bus:          bus,
		over:         false,
		score:        s,
		lastUpdate:   nil,
		running:      false,
		rpg:          rng.NewPiece(1234),
		wg:           &sync.WaitGroup{},
		mu:           &sync.RWMutex{},
		comm:         c,
	}

	return &g
}

func (g *Game) reset() {
	g.fallingPiece = nil
	g.nextPiece = nil
	g.holdingPiece = piece.NewLivingPiece(nil)
	g.over = false
	g.lastUpdate = nil
	g.running = false

	g.field.Reset()
	g.score.Reset()

	g.rpg.NextBag()

	ctx, cancel := context.WithCancel(context.Background())

	g.ctx = ctx
	g.stop = cancel
}

func (g *Game) ToPayload() *Payload {
	defer g.mu.RUnlock()
	g.mu.RLock()

	return &Payload{
		Id:         g.id,
		PlayerName: g.player.GetName(),
	}
}

func (g *Game) GetId() string {
	defer g.mu.RUnlock()
	g.mu.RLock()

	return g.id
}

func (g *Game) doUpdate(delta int64) {
	oldBedrockLevel := g.field.GetCurrentBedrock()
	clearedLines, gameOver := g.updateFallingPiece(delta)

	if clearedLines > 0 {
		g.score.AddLines(clearedLines)

		newBedrockLevel := g.field.GetCurrentBedrock()

		if newBedrockLevel == 0 {
			distributableBedrock := int(math.Max(float64(clearedLines-oldBedrockLevel), 0))
			log.Printf("can distribute %d bedrock\n", distributableBedrock)
			// todo: distribute bedrock
		}
	}

	if g.field.Dirty.Clear() {
		g.bus.Publish(&event.Event{
			Type:    event.TypeFieldUpdate,
			Origin:  event.OriginGame(g.id),
			Payload: g.field.ToPayload(),
		})
	}

	if g.fallingPiece.Dirty.Clear() {
		g.bus.Publish(&event.Event{
			Type:    event.TypeFallingPieceUpdate,
			Origin:  event.OriginGame(g.id),
			Payload: g.fallingPiece.ToPayload(),
		})
	}

	if g.nextPiece.Dirty.Clear() {
		g.bus.Publish(&event.Event{
			Type:    event.TypeNextPieceUpdate,
			Origin:  event.OriginGame(g.id),
			Payload: g.nextPiece.ToPayload(),
		})
	}

	if g.holdingPiece.Dirty.Clear() {
		g.bus.Publish(&event.Event{
			Type:    event.TypeHoldingPieceUpdate,
			Origin:  event.OriginGame(g.id),
			Payload: g.holdingPiece.ToPayload(),
		})
	}

	if g.score.Dirty.Clear() {
		g.bus.Publish(&event.Event{
			Type:    event.TypeScoreUpdate,
			Origin:  event.OriginGame(g.id),
			Payload: g.score.ToPayload(),
		})
	}

	if gameOver {
		g.over = true

		g.bus.Publish(&event.Event{
			Type:   event.TypeGameOver,
			Origin: event.OriginGame(g.id),
		})

		g.Stop()
	}
}

func (g *Game) Update() {
	defer g.mu.Unlock()
	g.mu.Lock()

	if g.over {
		return
	}

	now := time.Now().UnixMilli()

	if g.lastUpdate != nil {
		g.doUpdate(now - *g.lastUpdate)
	}

	g.lastUpdate = &now
}

func (g *Game) Start() {
	defer g.mu.Unlock()
	g.mu.Lock()

	g.reset()

	g.startUpdater()
	g.startCommandReader()

	g.running = true
}

func (g *Game) Stop() {
	defer g.mu.RUnlock()
	g.mu.RLock()

	g.stop()
	g.wg.Wait()
}

func (g *Game) startUpdater() {
	go func() {
		defer g.wg.Done()
		g.wg.Add(1)

		ticker := time.NewTicker(time.Millisecond * 10) // 100 fps
		defer ticker.Stop()

		defer func() {
			g.running = false
		}()

		for {
			select {
			case <-g.ctx.Done():
				break
			case <-ticker.C:
				g.Update()
				break
			}
		}
	}()
}

func (g *Game) startCommandReader() {
	go func() {
		defer g.wg.Done()
		g.wg.Add(1)

		ticker := time.NewTicker(time.Millisecond * 250)
		defer ticker.Stop()

		for {
			select {
			case <-g.ctx.Done():
				break
			case cmd := <-g.comm.GetInputChannel():
				g.HandleCommand(Command(cmd))
				break
			case <-ticker.C:
				break
			}
		}
	}()
}
