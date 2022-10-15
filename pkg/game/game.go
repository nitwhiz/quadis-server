package game

import (
	"context"
	"github.com/nitwhiz/quadis-server/pkg/communication"
	"github.com/nitwhiz/quadis-server/pkg/event"
	"github.com/nitwhiz/quadis-server/pkg/falling_piece"
	"github.com/nitwhiz/quadis-server/pkg/field"
	"github.com/nitwhiz/quadis-server/pkg/piece"
	"github.com/nitwhiz/quadis-server/pkg/player"
	"github.com/nitwhiz/quadis-server/pkg/rng"
	"github.com/nitwhiz/quadis-server/pkg/score"
	"math"
	"sync"
	"time"
)

type Settings struct {
	Id             string
	EventBus       *event.Bus
	Connection     *communication.Connection
	Player         *player.Player
	BedrockChannel chan *Bedrock
	ParentContext  context.Context
	Seed           int64
}

type Game struct {
	id             string
	player         *player.Player
	fallingPiece   *falling_piece.FallingPiece
	nextPiece      *piece.LivingPiece
	holdingPiece   *piece.LivingPiece
	field          *field.Field
	bus            *event.Bus
	over           bool
	score          *score.Score
	lastUpdate     *int64
	rpg            *rng.Piece
	ctx            context.Context
	stop           context.CancelFunc
	wg             *sync.WaitGroup
	mu             *sync.RWMutex
	comm           *communication.Connection
	bedrockChannel chan *Bedrock
	parentContext  context.Context
}

type Payload struct {
	Id         string `json:"id"`
	PlayerName string `json:"playerName"`
}

func New(settings *Settings) *Game {
	f := field.New()
	s := score.New()

	g := Game{
		id:             settings.Id,
		player:         settings.Player,
		fallingPiece:   nil,
		nextPiece:      nil,
		holdingPiece:   nil,
		field:          f,
		bus:            settings.EventBus,
		over:           true,
		score:          s,
		lastUpdate:     nil,
		rpg:            rng.NewPiece(settings.Seed),
		wg:             &sync.WaitGroup{},
		mu:             &sync.RWMutex{},
		comm:           settings.Connection,
		bedrockChannel: settings.BedrockChannel,
		parentContext:  settings.ParentContext,
	}

	return &g
}

func (g *Game) IsOver() bool {
	defer g.mu.RUnlock()
	g.mu.RLock()

	return g.over
}

func (g *Game) reset() {
	g.fallingPiece = nil
	g.nextPiece = nil
	g.holdingPiece = piece.NewLivingPiece(nil)
	g.over = true
	g.lastUpdate = nil

	g.field.Reset()
	g.score.Reset()

	g.rpg.NextBag()

	ctx, cancel := context.WithCancel(g.parentContext)

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

			g.bedrockChannel <- &Bedrock{
				Amount:   distributableBedrock,
				SourceId: g.id,
			}
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

		go g.Stop()
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

	g.over = false

	g.startUpdater()
	g.startCommandReader()
}

func (g *Game) Stop() {
	defer g.mu.RUnlock()
	g.mu.RLock()

	if g.stop != nil {
		g.stop()
		g.stop = nil
	}

	g.over = true

	g.wg.Wait()
}

func (g *Game) startUpdater() {
	go func() {
		defer g.wg.Done()
		g.wg.Add(1)

		for {
			select {
			case <-g.ctx.Done():
				return
			case <-time.After(time.Millisecond * 10): // 100 fps
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

		for {
			select {
			case <-g.ctx.Done():
				return
			case cmd := <-g.comm.Input:
				g.HandleCommand(Command(cmd))
				break
			case <-time.After(time.Millisecond * 250):
				break
			}
		}
	}()
}

func (g *Game) SendBedrock(amount int) {
	defer g.mu.RUnlock()
	g.mu.RLock()

	g.field.IncreaseBedrock(amount)
}
