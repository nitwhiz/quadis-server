package bloccs

import (
	"bloccs-server/pkg/event"
	"bloccs-server/pkg/rng"
	"github.com/google/uuid"
	"hash/fnv"
	"sync"
	"time"
)

const EventGameStart = "game_start"
const EventGameOver = "game_over"

type GameSettings struct {
	// FallingPieceSpeed falling piece speed in blocks per second
	FallingPieceSpeed float64
	Seed              string
	FieldWidth        int
	FieldHeight       int
}

type Game struct {
	Id             string `json:"id"`
	commandChannel chan string
	fallingPiece   *FallingPiece
	// fallingPieceSpeed falling piece speed in blocks per second
	fallingPieceSpeed float64
	holdPiece         *Piece
	holdLock          bool
	nextPiece         *Piece
	field             *Field
	eventBus          *event.Bus
	isOver            bool
	score             *Score
	rbg               *rng.RPG[Piece]
	lastUpdate        *time.Time
	globalWaitGroup   *sync.WaitGroup
	mu                *sync.Mutex
	stopChannel       chan bool
}

func (g *Game) GetId() string {
	return g.Id
}

func NewGame(bus *event.Bus, settings *GameSettings) *Game {
	h := fnv.New32a()

	_, _ = h.Write([]byte(settings.Seed))

	rpg := rng.NewRBG(int64(h.Sum32()), func() []*Piece {
		var b []*Piece

		for _, p := range AllPieces {
			b = append(b, NewPiece(p))
		}

		return b
	})

	gameId := uuid.NewString()

	game := &Game{
		Id:                gameId,
		commandChannel:    make(chan string),
		fallingPiece:      NewFallingPiece(bus, gameId),
		fallingPieceSpeed: settings.FallingPieceSpeed,
		holdPiece:         nil,
		holdLock:          false,
		nextPiece:         nil,
		field:             NewField(bus, gameId, settings.FieldWidth, settings.FieldHeight),
		eventBus:          bus,
		isOver:            false,
		score:             NewScore(gameId),
		rbg:               rpg,
		lastUpdate:        nil,
		globalWaitGroup:   &sync.WaitGroup{},
		mu:                &sync.Mutex{},
		stopChannel:       make(chan bool),
	}

	game.nextFallingPiece()

	return game
}

func (g *Game) GetCommandChannel() chan string {
	return g.commandChannel
}

func (g *Game) Start() {
	g.rbg.NextBag()

	now := time.Now()
	g.lastUpdate = &now

	go func() {
		defer g.globalWaitGroup.Done()
		g.globalWaitGroup.Add(1)

		// 100 tps
		ticker := time.NewTicker(time.Millisecond * 10)

		defer ticker.Stop()

		for {
			select {
			case <-g.stopChannel:
				return
			case cmd := <-g.commandChannel:
				// todo(performance): only update (and/or publish) if necessary
				g.Command(cmd)
				g.Update()
			case <-ticker.C:
				// todo(performance): only update (and/or publish) if necessary
				g.Update()
			}
		}
	}()

	g.eventBus.Publish(event.New(EventGameStart, g, nil))
}

func (g *Game) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.holdPiece = nil
	g.holdLock = false
	g.nextPiece = nil
	g.isOver = false
	g.lastUpdate = nil

	g.score.Reset()
	g.field.Reset()
}

func (g *Game) Stop() {
	g.stopChannel <- true
	close(g.stopChannel)

	g.globalWaitGroup.Wait()
}

func (g *Game) Update() {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.isOver {
		return
	}

	now := time.Now()

	if g.lastUpdate != nil {
		clearedRows, gameOver := g.updateFallingPiece(int(time.Now().Sub(*g.lastUpdate)))

		g.score.AddLines(clearedRows)
		g.field.DecreaseBedrock(clearedRows)

		if gameOver {
			g.isOver = true
			g.eventBus.Publish(event.New(EventGameOver, g, nil))
		}
	}

	g.lastUpdate = &now
}
