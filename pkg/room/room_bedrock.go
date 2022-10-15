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
}

func NewBedrockDistribution(room *Room) *BedrockDistribution {
	gameIdBagGenerator := func() []string {
		defer room.gamesMutex.RUnlock()
		room.gamesMutex.RLock()

		var gameIds []string

		for _, g := range room.games {
			gameIds = append(gameIds, g.GetId())
		}

		return gameIds
	}

	d := BedrockDistribution{
		Channel:         make(chan *game.Bedrock, 64),
		randomGameIdBag: rng.NewString(1234, gameIdBagGenerator),
		targetMap:       map[string]string{},
		bus:             room.bus,
		ctx:             room.ctx,
		room:            room,
		mu:              &sync.RWMutex{},
	}

	return &d
}

func (d *BedrockDistribution) startDistribution() {
	for {
		// avoid lock up
		time.Sleep(time.Millisecond * 1)

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

	// todo: add probability to choose no player as target
	// todo: never choose game over player as target

	d.randomGameIdBag.NextBag()

	bSize := d.randomGameIdBag.GetSize()

	for i := 0; i < bSize*2; i++ {
		d.targetMap[d.randomGameIdBag.NextElement()] = d.randomGameIdBag.NextElement()
	}

	d.bus.Publish(&event.Event{
		Type:   event.TypeBedrockTargetsUpdate,
		Origin: event.OriginRoom(d.room.GetId()),
		Payload: &BedrockTargetsPayload{
			Targets: d.targetMap,
		},
	})
}
