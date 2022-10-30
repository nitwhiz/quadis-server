package room

import (
	"context"
	"github.com/nitwhiz/quadis-server/pkg/event"
	"github.com/nitwhiz/quadis-server/pkg/game"
	"github.com/nitwhiz/quadis-server/pkg/rng"
	"sync"
	"time"
)

type BedrockDistribution struct {
	Channel         chan *game.Bedrock
	randomGameIdBag *rng.String
	targetMap       map[string]string
	bus             *event.Bus
	ctx             context.Context
	room            *Room
	mu              *sync.RWMutex
	random          *rng.Basic
}

func NewBedrockDistribution(room *Room, seed int64) *BedrockDistribution {
	gameIdBagGenerator := func() []string {
		defer room.gamesMutex.RUnlock()
		room.gamesMutex.RLock()

		var gameIds []string

		for _, g := range room.games {
			if !g.IsOver() {
				gameIds = append(gameIds, g.GetId())
			}
		}

		return gameIds
	}

	d := BedrockDistribution{
		Channel:         make(chan *game.Bedrock, 64),
		randomGameIdBag: rng.NewString(seed, gameIdBagGenerator),
		targetMap:       map[string]string{},
		bus:             room.bus,
		ctx:             room.ctx,
		room:            room,
		mu:              &sync.RWMutex{},
		random:          rng.NewBasic(seed),
	}

	return &d
}

func (d *BedrockDistribution) startDistribution() {
	for {
		select {
		case <-d.ctx.Done():
			return
		case b := <-d.Channel:
			d.mu.RLock()

			if targetGameId, ok := d.targetMap[b.SourceId]; ok {
				if targetGameId != b.SourceId {
					d.room.gamesMutex.RLock()

					if targetGame, ok := d.room.games[targetGameId]; ok {
						targetGame.SendBedrock(b.Amount)
					}

					d.room.gamesMutex.RUnlock()
				}
			}

			d.mu.RUnlock()

			// avoid lock up
			time.Sleep(time.Microsecond * 250)
			break
		}
	}
}

func (d *BedrockDistribution) startRandomizer() {
	for {
		select {
		case <-d.ctx.Done():
			return
		case <-time.After(time.Second * 3):
			d.Randomize()
		}
	}
}

func (d *BedrockDistribution) Start() {
	go d.startDistribution()
	go d.startRandomizer()
}

func (d *BedrockDistribution) Randomize() {
	defer d.mu.Unlock()
	d.mu.Lock()

	if d.room.GetRunningGamesCount() > 2 {
		d.defaultRandomizer()
	} else {
		d.deathMatchRandomizer()
	}

	d.bus.Publish(&event.Event{
		Type:   event.TypeBedrockTargetsUpdate,
		Origin: event.OriginRoom(d.room.GetId()),
		Payload: &BedrockTargetsPayload{
			Targets: d.targetMap,
		},
	})
}
