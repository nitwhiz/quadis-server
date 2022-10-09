package room

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nitwhiz/quadis-server/pkg/communication"
	"github.com/nitwhiz/quadis-server/pkg/event"
	"github.com/nitwhiz/quadis-server/pkg/game"
	"github.com/nitwhiz/quadis-server/pkg/player"
	"log"
	"sync"
)

type Room struct {
	id         string
	games      map[string]*game.Game
	gamesMutex *sync.RWMutex
	bus        *event.Bus
	wg         *sync.WaitGroup
	mu         *sync.RWMutex
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
	b := event.NewBus()

	return &Room{
		id:         uuid.NewString(),
		games:      map[string]*game.Game{},
		gamesMutex: &sync.RWMutex{},
		bus:        b,
		wg:         &sync.WaitGroup{},
		mu:         &sync.RWMutex{},
	}
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

func (r *Room) CreatePlayer(ws *websocket.Conn) error {
	c := communication.NewConnection(ws)

	hrm, err := r.HandshakeGreeting(c)

	if err != nil {
		return err
	}

	log.Printf("%+v\n", hrm)

	p := player.New(hrm.PlayerName)
	g := game.New(r.bus, c, p)

	r.mu.Lock()
	r.games[g.GetId()] = g
	r.mu.Unlock()

	r.bus.SubscribeAll(func(event *event.Event) {
		// todo: serialization can happen once for broadcasts
		msg, _ := event.Serialize()
		c.Write(msg)
	})

	r.bus.Publish(&event.Event{
		Type:    event.TypeJoin,
		Origin:  event.OriginRoom(r.GetId()),
		Payload: g.ToPayload(),
	})

	err = r.HandshakeAck(c, g, false)

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
		g.Start()
	}
}
