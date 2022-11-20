package game

import (
	"context"
	"github.com/nitwhiz/quadis-server/pkg/communication"
	"github.com/nitwhiz/quadis-server/pkg/event"
	"github.com/nitwhiz/quadis-server/pkg/falling_piece"
	"github.com/nitwhiz/quadis-server/pkg/field"
	"github.com/nitwhiz/quadis-server/pkg/piece"
	"github.com/nitwhiz/quadis-server/pkg/player"
	"github.com/nitwhiz/quadis-server/pkg/score"
	"math"
	"sync"
	"time"
)

type OverCallback func()

type ActivateItemCallback func(g *Game)

type Settings struct {
	Id                   string
	EventBus             *event.Bus
	Connection           *communication.Connection
	Player               *player.Player
	BedrockChannel       chan *Bedrock
	ParentContext        context.Context
	OverCallback         OverCallback
	ActivateItemCallback ActivateItemCallback
	Seed                 int64
	IsHost               bool
}

type Game struct {
	id                   string
	player               *player.Player
	fallingPiece         *falling_piece.FallingPiece
	nextPiece            *piece.LivingPiece
	holdingPiece         *piece.LivingPiece
	field                *field.Field
	bus                  *event.Bus
	over                 bool
	score                *score.Score
	lastUpdate           *int64
	pieceGenerator       *piece.Generator
	ctx                  context.Context
	stop                 context.CancelFunc
	wg                   *sync.WaitGroup
	mu                   *sync.RWMutex
	con                  *communication.Connection
	bedrockChannel       chan *Bedrock
	overCallback         OverCallback
	activateItemCallback ActivateItemCallback
	lastActivity         time.Time
	host                 bool
	overridePiece        *piece.Piece
}

type Payload struct {
	Id         string `json:"id"`
	PlayerName string `json:"playerName"`
}

func New(settings *Settings) *Game {
	f := field.New(&field.Settings{
		Seed:   settings.Seed,
		Width:  10,
		Height: 20,
	})
	s := score.New()

	ctx, cancel := context.WithCancel(settings.ParentContext)

	g := Game{
		id:                   settings.Id,
		player:               settings.Player,
		fallingPiece:         nil,
		nextPiece:            nil,
		holdingPiece:         nil,
		field:                f,
		bus:                  settings.EventBus,
		over:                 true,
		score:                s,
		lastUpdate:           nil,
		pieceGenerator:       nil,
		wg:                   &sync.WaitGroup{},
		mu:                   &sync.RWMutex{},
		con:                  settings.Connection,
		bedrockChannel:       settings.BedrockChannel,
		ctx:                  ctx,
		stop:                 cancel,
		overCallback:         settings.OverCallback,
		activateItemCallback: settings.ActivateItemCallback,
		lastActivity:         time.Now(),
		host:                 settings.IsHost,
	}

	go g.startCommandReader()
	go g.startUpdater()

	return &g
}

func (g *Game) IsHost() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.host
}

func (g *Game) activateItem() {
	g.activateItemCallback(g)
}

func (g *Game) GetField() *field.Field {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.field
}

func (g *Game) GetScore() *score.Score {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.score
}

func (g *Game) GetPlayer() *player.Player {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.player
}

func (g *Game) IsOver() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.over
}

func (g *Game) GetFallingPiece() *falling_piece.FallingPiece {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.fallingPiece
}

func (g *Game) SetOverridePiece(piece *piece.Piece) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.overridePiece = piece

	if piece != nil {
		g.fallingPiece.SetPiece(piece, g.field.GetCenterX(), 0, 0)
		g.nextPiece.SetPiece(piece)
	} else {
		g.nextPiece.SetPiece(g.pieceGenerator.NextElement())
	}
}

func (g *Game) init(seed int64) {
	g.pieceGenerator = piece.NewGenerator(seed)

	g.fallingPiece = nil
	g.nextPiece = nil
	g.holdingPiece = piece.NewLivingPiece(nil)
	g.over = true
	g.lastUpdate = nil

	g.field.Reset()
	g.score.Reset()

	g.pieceGenerator.NextBag()

	g.lastActivity = time.Now()
}

func (g *Game) ToPayload() *Payload {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return &Payload{
		Id:         g.id,
		PlayerName: g.player.GetName(),
	}
}

func (g *Game) GetId() string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.id
}

func (g *Game) doUpdate(delta int64) {
	oldBedrockLevel := g.field.GetCurrentBedrock()
	clearedLines, gameOver := g.updateFallingPiece(delta)

	if clearedLines > 0 {
		g.score.AddLines(clearedLines)

		newBedrockLevel := g.field.GetCurrentBedrock()

		if newBedrockLevel == 0 && g.bedrockChannel != nil {
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
		g.bus.Publish(&event.Event{
			Type:   event.TypeGameOver,
			Origin: event.OriginGame(g.id),
		})

		go g.ToggleOver(false)
	}
}

func (g *Game) Update() {
	if g.IsOver() {
		return
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli()

	if g.lastUpdate != nil {
		g.doUpdate(now - *g.lastUpdate)
	}

	g.lastUpdate = &now
}

func (g *Game) Start(seed int64) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.init(seed)

	g.over = false
}

func (g *Game) GetLastActivity() time.Time {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.lastActivity
}

// ToggleOver sets over to true
func (g *Game) ToggleOver(shutdown bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.over != true && !shutdown {
		go g.overCallback()
	}

	g.over = true
}

func (g *Game) Stop() {
	if g.stop != nil {
		g.stop()
	}

	g.wg.Wait()
}

func (g *Game) startUpdater() {
	g.wg.Add(1)
	defer g.wg.Done()

	for {
		select {
		case <-g.ctx.Done():
			return
		case <-time.After(time.Millisecond * 10): // 100 fps
			g.Update()
			break
		}
	}
}

func (g *Game) startCommandReader() {
	g.wg.Add(1)
	defer g.wg.Done()

	for {
		select {
		case <-g.ctx.Done():
			return
		case cmd := <-g.con.GetInputChannel():
			g.HandleCommand(Command(cmd))
			break
		case <-time.After(time.Millisecond * 250):
			break
		}
	}
}

func (g *Game) AddBedrock(amount int) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	g.field.IncreaseBedrock(amount)
}
