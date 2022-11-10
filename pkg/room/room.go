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
	stop                context.CancelFunc
	bedrockDistribution *BedrockDistribution
	gamesStarted        bool
	gameOverCount       int
	createdAt           *time.Time
	randomSeed          *rng.Basic
	rules               *Rules
}

type Payload struct {
	Id    string          `json:"id"`
	Games []*game.Payload `json:"games"`
}

type BedrockTargetsPayload struct {
	Targets map[string]string `json:"targets"`
}

func New() *Room {
	ctx, cancel := context.WithCancel(context.Background())

	b := event.NewBus(ctx)

	now := time.Now()

	r := Room{
		id:            uuid.NewString(),
		games:         map[string]*game.Game{},
		gamesMutex:    &sync.RWMutex{},
		bus:           b,
		wg:            &sync.WaitGroup{},
		mu:            &sync.RWMutex{},
		ctx:           ctx,
		stop:          cancel,
		gamesStarted:  false,
		createdAt:     &now,
		gameOverCount: 0,
		randomSeed:    rng.NewBasic(now.UnixMicro()),
		rules: &Rules{
			BedrockEnabled: true,
		},
	}

	if r.rules.BedrockEnabled {
		// todo: remove cyclic dependency

		r.bedrockDistribution = NewBedrockDistribution(&r, r.randomSeed.NextInt64())
		r.bedrockDistribution.Start()
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

func (r *Room) StopGames() {
	defer r.gamesMutex.RUnlock()
	r.gamesMutex.RLock()

	for _, g := range r.games {
		g.ToggleOver()
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.gamesStarted = false
}

func (r *Room) Stop() {
	r.StopGames()

	r.mu.Lock()
	defer r.mu.Unlock()

	r.stop()
	r.wg.Wait()
}
