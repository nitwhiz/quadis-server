package room

import (
	"github.com/nitwhiz/quadis-server/pkg/communication"
	"github.com/nitwhiz/quadis-server/pkg/event"
)

const MsgIdJoin = "player_join"
const MsgIdLeave = "player_leave"

const MsgParamPlayerName = "player_name"

func (r *Room) WriteMessage(c *communication.Connection, msgId string, parameters map[string]string) {
	msg, _ := (&event.Event{
		Type:   event.TypeMessage,
		Origin: event.OriginRoom(r.GetId()),
		Payload: &MessagePayload{
			Id:         msgId,
			Parameters: parameters,
		},
	}).Serialize()

	c.Write(msg)
}
