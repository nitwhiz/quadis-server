package room

import (
	"github.com/nitwhiz/quadis-server/pkg/event"
	"github.com/nitwhiz/quadis-server/pkg/game"
	"github.com/nitwhiz/quadis-server/pkg/item"
	"github.com/nitwhiz/quadis-server/pkg/metrics"
	"github.com/nitwhiz/quadis-server/pkg/rng"
	"log"
	"sync"
	"time"
)

type ItemPayload struct {
	Type *string `json:"type"`
}

type ItemDistribution struct {
	room          *Room
	mu            *sync.RWMutex
	gameItems     map[string]*item.Item
	itemGenerator *item.Generator
	random        *rng.Basic
}

func (r *Room) StartItemDistribution(seed int64) {
	if r.itemDistribution != nil {
		log.Fatalln("trying to init another item distribution")
	}

	id := ItemDistribution{
		room:          r,
		mu:            &sync.RWMutex{},
		gameItems:     map[string]*item.Item{},
		itemGenerator: item.NewGenerator(seed),
		random:        rng.NewBasic(seed),
	}

	go id.startDistribution()

	r.itemDistribution = &id
}

func (r *Room) UpdateItemAffection(gameId string, itemType string) {
	r.itemDistribution.UpdateItemAffection(gameId, itemType)
}

func (i *ItemDistribution) UpdateItemAffection(gameId string, itemType string) {
	i.room.bus.Publish(&event.Event{
		Type:   event.TypeItemAffectionUpdate,
		Origin: event.OriginGame(gameId),
		Payload: &ItemPayload{
			Type: &itemType,
		},
	})
}

func (i *ItemDistribution) ActivateItem(sourceGame *game.Game) {
	i.mu.Lock()
	defer i.mu.Unlock()

	gameId := sourceGame.GetId()

	if gameItem, ok := i.gameItems[gameId]; ok && gameItem != nil {
		go gameItem.Activate(sourceGame, i.room)

		metrics.IncreaseItemActivationsTotal(gameItem.Type)

		i.gameItems[gameId] = nil

		i.room.bus.Publish(&event.Event{
			Type:   event.TypeItemUpdate,
			Origin: event.OriginGame(gameId),
			Payload: &ItemPayload{
				Type: nil,
			},
		})
	}
}

func (i *ItemDistribution) randomize() {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.room.gamesMutex.RLock()
	defer i.room.gamesMutex.RUnlock()

	for gId := range i.room.games {
		if gItem, ok := i.gameItems[gId]; !ok || gItem == nil {
			newItem := i.itemGenerator.NextElement()

			if i.random.Probably(.75) {
				i.gameItems[gId] = newItem

				i.room.bus.Publish(&event.Event{
					Type:   event.TypeItemUpdate,
					Origin: event.OriginGame(gId),
					Payload: &ItemPayload{
						Type: &newItem.Type,
					},
				})
			}
		}
	}
}

func (i *ItemDistribution) startDistribution() {
	i.room.wg.Add(1)
	defer i.room.wg.Done()

	for {
		select {
		case <-i.room.ctx.Done():
			return
		case <-time.After(time.Second * 10):
			i.randomize()
		}
	}
}
