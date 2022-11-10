package room

import (
	"context"
	"github.com/google/uuid"
	"github.com/nitwhiz/quadis-server/pkg/event"
	"github.com/nitwhiz/quadis-server/pkg/game"
	"github.com/nitwhiz/quadis-server/pkg/rng"
	"sync"
	"time"
)

type Room struct {
	id                  string
	games               map[string]*game.Game
	gamesMutex          *sync.RWMutex
	bus                 *event.Bus
	wg                  *sync.WaitGroup
	mu                  *sync.RWMutex
	ctx                 context.Context
	shutdown            context.CancelFunc
	gamesStarted        bool
	gameOverCount       int
	createdAt           time.Time
	randomSeed          *rng.Basic
	rules               *Rules
	targets             *TargetsDistribution
	bedrockDistribution *BedrockDistribution
	itemDistribution    *ItemDistribution
}

type Payload struct {
	Id    string          `json:"id"`
	Games []*game.Payload `json:"games"`
}

func New() *Room {
	ctx, shutdown := context.WithCancel(context.Background())

	b := event.NewBus(ctx)

	now := time.Now()

	r := Room{
		id:               uuid.NewString(),
		games:            map[string]*game.Game{},
		gamesMutex:       &sync.RWMutex{},
		itemDistribution: nil,
		bus:              b,
		wg:               &sync.WaitGroup{},
		mu:               &sync.RWMutex{},
		ctx:              ctx,
		shutdown:         shutdown,
		gamesStarted:     false,
		createdAt:        time.Now(),
		gameOverCount:    0,
		randomSeed:       rng.NewBasic(now.UnixMicro()),
		rules: &Rules{
			BedrockEnabled: true,
			ItemsEnabled:   true,
		},
	}

	r.StartCurfewBouncer()
	r.StartTargetDistribution()

	if r.rules.BedrockEnabled {
		r.StartBedrockDistribution()
	}

	if r.rules.ItemsEnabled {
		r.StartItemDistribution()
	}

	return &r
}

func (r *Room) ToPayload() *Payload {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var gps []*game.Payload

	for _, g := range r.games {
		gps = append(gps, g.ToPayload())
	}

	return &Payload{
		Id:    r.id,
		Games: gps,
	}
}

func (r *Room) GetWaitGroup() *sync.WaitGroup {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.wg
}

func (r *Room) GetEventBus() *event.Bus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.bus
}

func (r *Room) ShutdownDone() <-chan struct{} {
	return r.ctx.Done()
}

func (r *Room) GetGames() map[string]*game.Game {
	r.gamesMutex.RLock()
	defer r.gamesMutex.RUnlock()

	res := map[string]*game.Game{}

	for gId, g := range r.games {
		res[gId] = g
	}

	return res
}

func (r *Room) GetLastActivity() time.Time {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.gamesMutex.RLock()
	defer r.gamesMutex.RUnlock()

	if len(r.games) == 0 {
		zeroTime, _ := time.Parse(time.RFC822, "06 Dec 95 22:00 CET")
		return zeroTime
	}

	var latestActivity time.Time

	for _, g := range r.games {
		lastGameActivity := g.GetLastActivity()

		if latestActivity.Before(lastGameActivity) {
			latestActivity = lastGameActivity
		}
	}

	return latestActivity
}

func (r *Room) GetId() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.id
}

func (r *Room) Start() {
	r.gamesMutex.RLock()
	defer r.gamesMutex.RUnlock()

	r.bus.Publish(&event.Event{
		Type:   event.TypeStart,
		Origin: event.OriginRoom(r.id),
	})

	seed := r.randomSeed.NextInt64()

	for _, g := range r.games {
		g.Start(seed)
	}

	r.gamesStarted = true
}

func (r *Room) StopGames(shutdown bool) {
	r.gamesMutex.RLock()
	defer r.gamesMutex.RUnlock()

	for _, g := range r.games {
		g.ToggleOver(shutdown)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.gamesStarted = false
}

func (r *Room) Shutdown() {
	r.StopGames(true)

	r.mu.Lock()
	defer r.mu.Unlock()

	r.shutdown()
	r.wg.Wait()
}
