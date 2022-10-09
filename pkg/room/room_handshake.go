package room

import (
	"encoding/json"
	"errors"
	"github.com/nitwhiz/quadis-server/pkg/communication"
	"github.com/nitwhiz/quadis-server/pkg/event"
	"github.com/nitwhiz/quadis-server/pkg/game"
)

type HelloAckPayload struct {
	Room           *Payload      `json:"room"`
	ControlledGame *game.Payload `json:"controlledGame"`
	Host           bool          `json:"host"`
}

type HelloResponseMessage struct {
	PlayerName string `json:"playerName"`
}

func (hmr *HelloResponseMessage) Validate() bool {
	return true
}

func (r *Room) HandshakeGreeting(c *communication.Connection) (*HelloResponseMessage, error) {
	msg, err := (&event.Event{
		Type:   event.TypeHello,
		Origin: event.OriginRoom(r.GetId()),
	}).Serialize()

	if err != nil {
		return nil, err
	}

	c.Write(msg)

	resp := c.Read()

	if resp == "" {
		return nil, errors.New("empty hello response")
	}

	var hrm HelloResponseMessage

	err = json.Unmarshal([]byte(resp), &hrm)

	if err != nil {
		return nil, errors.New("malformed hello response")
	}

	if !hrm.Validate() {
		return nil, errors.New("invalid hello response")
	}

	return &hrm, nil
}

func (r *Room) HandshakeAck(c *communication.Connection, g *game.Game, host bool) error {
	msg, err := (&event.Event{
		Type:   event.TypeHelloAck,
		Origin: event.OriginRoom(r.GetId()),
		Payload: &HelloAckPayload{
			Room:           r.ToPayload(),
			ControlledGame: g.ToPayload(),
			Host:           host,
		},
	}).Serialize()

	if err != nil {
		return err
	}

	c.Write(msg)

	return nil
}
