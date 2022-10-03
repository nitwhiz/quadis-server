package server

import (
	"github.com/nitwhiz/bloccs-server/pkg/event"
	"github.com/nitwhiz/bloccs-server/pkg/game"
	"log"
)

func (r *Room) passEvent(p *Player, e *event.Event) error {
	// todo: throttle events from other players; only send specific events instantly
	// todo: this data-races if event body is something not locked - may be irrelevant though

	bs, err := e.GetAsBytes()

	if err != nil {
		log.Printf("cannot get bytes of message")
		return err
	}

	if err := p.SendMessage(bs); err != nil {
		return err
	}

	return nil
}

func (r *Room) AddPlayer(p *Player) {
	r.playersMutex.Lock()
	defer r.playersMutex.Unlock()

	if _, ok := r.Players[p.ID]; ok {
		return
	}

	if len(r.Players) >= 7 {
		return
	}

	r.Players[p.ID] = p

	r.eventBus.Subscribe(event.ChannelRoom, func(e *event.Event) {
		if err := r.passEvent(p, e); err != nil {
			log.Println("error passing event")
			r.RemovePlayer(p)
		}
	}, p.ID)

	r.eventBus.Subscribe("update/.*", func(e *event.Event) {
		if e.Type == event.GameOver {
			// todo
		}

		if e.Type == event.RowsCleared {
			if rcp, ok := e.Payload.(*event.RowsClearedPayload); ok {
				if rcp.BedrockCount == 0 {
					if targetGameID, ok := r.bedrockTargetMap[rcp.GameId]; ok {
						if tp, ok := r.Players[targetGameID]; ok && p.game.ID != rcp.GameId {
							// todo: refactor this mess

							tp.game.Field.IncreaseBedrock(rcp.RowsCount)

							squished := false

							for !tp.game.CanPutFallingPiece() {
								squished = true

								if tp.game.FallingPiece.Y <= 0 {
									break
								}

								tp.game.FallingPiece.Y--
							}

							if squished {
								tp.game.Command(game.CommandHardLock)
							}
						}
					}
				}
			}
		}

		if err := r.passEvent(p, e); err != nil {
			log.Println("error passing event")
			r.RemovePlayer(p)
		}
	}, p.ID)

	r.eventBus.Publish(event.New(event.ChannelRoom, event.PlayerJoin, &event.PlayerJoinPayload{
		ID:       p.ID,
		Name:     p.Name,
		CreateAt: p.CreateAt,
	}))

	r.randomPlayerIdGenerator.NextBag()
}

func (r *Room) RemovePlayer(p *Player) {
	r.playersMutex.Lock()
	defer r.playersMutex.Unlock()

	if _, ok := r.Players[p.ID]; !ok {
		return
	}

	r.eventBus.Unsubscribe(p.ID)

	delete(r.Players, p.ID)

	r.eventBus.Publish(event.New(event.ChannelRoom, event.PlayerLeave, &event.PlayerLeavePayload{
		ID:       p.ID,
		Name:     p.Name,
		CreateAt: p.CreateAt,
	}))
}
