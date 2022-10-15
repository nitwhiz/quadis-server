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

func NewBedrockDistribution(room *Room) *BedrockDistribution {
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
		randomGameIdBag: rng.NewString(1234, gameIdBagGenerator),
		targetMap:       map[string]string{},
		bus:             room.bus,
		ctx:             room.ctx,
		room:            room,
		mu:              &sync.RWMutex{},
		random:          rng.NewBasic(1234),
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

	d.targetMap = map[string]string{}

	d.randomGameIdBag.NextBag()

	bSize := d.randomGameIdBag.GetSize()

	if bSize <= 1 {
		return
	}

	for i := 0; i < bSize; i++ {
		randomId := d.randomGameIdBag.NextElement()

		d.targetMap[randomId] = randomId
	}

	// just to be extra sure: new bag
	d.randomGameIdBag.NextBag()

	for id := range d.targetMap {
		target := d.randomGameIdBag.NextElement()

		if d.random.Probably(.25) {
			continue
		}

		d.targetMap[id] = target
	}

	d.bus.Publish(&event.Event{
		Type:   event.TypeBedrockTargetsUpdate,
		Origin: event.OriginRoom(d.room.GetId()),
		Payload: &BedrockTargetsPayload{
			Targets: d.targetMap,
		},
	})
}
