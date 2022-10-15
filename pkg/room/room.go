package room

import (
	"context"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nitwhiz/quadis-server/pkg/communication"
	"github.com/nitwhiz/quadis-server/pkg/event"
	"github.com/nitwhiz/quadis-server/pkg/game"
	"github.com/nitwhiz/quadis-server/pkg/player"
	"sync"
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
}

type Payload struct {
	Id    string          `json:"id"`
	Games []*game.Payload `json:"games"`
}

type BedrockTargetsPayload struct {
	Targets map[string]string `json:"targets"`
}

type MessagePayload struct {
	Id         string            `json:"id"`
	Parameters map[string]string `json:"parameters"`
}

func New() *Room {
	ctx, cancel := context.WithCancel(context.Background())

	b := event.NewBus(ctx)

	r := Room{
		id:         uuid.NewString(),
		games:      map[string]*game.Game{},
		gamesMutex: &sync.RWMutex{},
		bus:        b,
		wg:         &sync.WaitGroup{},
		mu:         &sync.RWMutex{},
		ctx:        ctx,
		stop:       cancel,
	}

	// I don't like this cyclic dependency
	r.bedrockDistribution = NewBedrockDistribution(&r, 1234)

	r.bedrockDistribution.Start()

	return &r
}

func (r *Room) ToPayload() *Payload {
	defer r.mu.RUnlock()
	r.mu.RLock()

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
	defer r.mu.RUnlock()
	r.mu.RLock()

	return r.id
}

func (r *Room) RemoveGame(id string) {
	r.gamesMutex.Lock()

	if g, ok := r.games[id]; ok {
		r.bus.Unsubscribe(id)

		g.Stop()
		delete(r.games, id)

		r.bus.Publish(&event.Event{
			Type:    event.TypeLeave,
			Origin:  event.OriginRoom(r.GetId()),
			Payload: g.ToPayload(),
		})

		r.gamesMutex.Unlock()

		r.bedrockDistribution.Randomize()
	} else {
		r.gamesMutex.Unlock()
	}
}

func (r *Room) CreateGame(ws *websocket.Conn) error {
	gameId := uuid.NewString()

	c := communication.NewConnection(&communication.Settings{
		WS:            ws,
		ParentContext: r.ctx,
		PreStopCallback: func() {
			r.RemoveGame(gameId)
		},
	})

	hrm, err := r.HandshakeGreeting(c)

	if err != nil {
		return err
	}

	p := player.New(hrm.PlayerName)
	g := game.New(&game.Settings{
		Id:             gameId,
		EventBus:       r.bus,
		Connection:     c,
		Player:         p,
		BedrockChannel: r.bedrockDistribution.Channel,
		ParentContext:  r.ctx,
		Seed:           1234,
	})

	r.mu.Lock()
	r.games[gameId] = g
	r.mu.Unlock()

	err = r.HandshakeAck(c, g, false)

	r.bus.Subscribe(gameId, c)

	r.bus.Publish(&event.Event{
		Type:    event.TypeJoin,
		Origin:  event.OriginRoom(r.GetId()),
		Payload: g.ToPayload(),
	})

	r.bedrockDistribution.Randomize()

	if err != nil {
		return err
	}

	return nil
}

func (r *Room) Start() {
	defer r.gamesMutex.RUnlock()
	r.gamesMutex.RLock()

	r.bus.Publish(&event.Event{
		Type:   event.TypeStart,
		Origin: event.OriginRoom(r.id),
	})

	for _, g := range r.games {
		g.Stop()
		g.Start()
	}
}

func (r *Room) Stop() {
	defer r.gamesMutex.RUnlock()
	r.gamesMutex.RLock()

	defer r.mu.Unlock()
	r.mu.Lock()

	for _, g := range r.games {
		g.Stop()
	}

	r.stop()
	r.wg.Wait()
}
