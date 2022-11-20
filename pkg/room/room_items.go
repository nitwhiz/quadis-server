package room

import (
	"github.com/nitwhiz/quadis-server/pkg/event"
	"github.com/nitwhiz/quadis-server/pkg/game"
	"github.com/nitwhiz/quadis-server/pkg/item"
	"log"
	"sync"
	"time"
)

type ItemPayload struct {
	Type *string `json:"type"`
}

type ItemDistribution struct {
	room      *Room
	mu        *sync.RWMutex
	gameItems map[string]*item.Item
}

func (r *Room) StartItemDistribution() {
	if r.itemDistribution != nil {
		log.Fatalln("trying to init another item distribution")
	}

	id := ItemDistribution{
		room:      r,
		mu:        &sync.RWMutex{},
		gameItems: map[string]*item.Item{},
	}

	go id.startDistribution()

	r.itemDistribution = &id
}

func (i *ItemDistribution) ActivateItem(sourceGame *game.Game) {
	i.mu.Lock()
	defer i.mu.Unlock()

	gameId := sourceGame.GetId()

	log.Printf("item activated by %s!\n", gameId)

	if gameItem, ok := i.gameItems[gameId]; ok {
		go gameItem.Activate(sourceGame, i.room)
		delete(i.gameItems, gameId)

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
		if _, ok := i.gameItems[gId]; !ok {
			newItem := item.NewTornado()

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
