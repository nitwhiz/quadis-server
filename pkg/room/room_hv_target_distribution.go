package room

import (
	"github.com/nitwhiz/quadis-server/pkg/event"
	"github.com/nitwhiz/quadis-server/pkg/rng"
	"sync"
	"time"
)

type TargetsPayload struct {
	Targets map[string]string `json:"targets"`
}

type TargetsDistribution struct {
	room            *Room
	randomGameIdBag *rng.String
	targetMap       map[string]string
	bus             *event.Bus
	mu              *sync.RWMutex
	random          *rng.Basic
}

func (r *Room) StartTargetDistribution() {
	gameIdBagGenerator := func() []string {
		var gameIds []string

		for gId, g := range r.GetGames() {
			if !g.IsOver() {
				gameIds = append(gameIds, gId)
			}
		}

		return gameIds
	}

	seed := r.randomSeed.NextInt64()

	tCtx := TargetsDistribution{
		room:            r,
		randomGameIdBag: rng.NewString(seed, gameIdBagGenerator),
		targetMap:       map[string]string{},
		bus:             r.GetEventBus(),
		mu:              &sync.RWMutex{},
		random:          rng.NewBasic(seed),
	}

	r.NewHypervisor(&HypervisorConfig{
		StartType: HypervisorStartTypeLazy,
		Interval:  time.Second * 5,
	}, func() {
		tCtx.Randomize()
	}).Start()

	r.targets = &tCtx
}

func (t *TargetsDistribution) GetTargetGameId(sourceGameId string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if targetGameId, ok := t.targetMap[sourceGameId]; ok {
		return targetGameId
	}

	return ""
}

func (t *TargetsDistribution) Randomize() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.room.GetRunningGamesCount() > 2 {
		t.defaultRandomizer()
	} else {
		t.deathMatchRandomizer()
	}

	t.bus.Publish(&event.Event{
		Type:   event.TypeTargetsUpdate,
		Origin: event.OriginRoom(t.room.GetId()),
		Payload: &TargetsPayload{
			Targets: t.targetMap,
		},
	})
}
