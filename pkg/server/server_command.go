package server

import (
	"encoding/json"
	"errors"
	"github.com/nitwhiz/quadis-server/pkg/room"
	"strings"
)

// this is just dev stuff and does not need to be pretty. it should _just work_.

type CommandType string

const CommandTypeSetField = CommandType("set_field")

type RoomConsoleCommand struct {
	CommandType CommandType       `json:"cmdType"`
	Payload     map[string]string `json:"payload"`
}

func (s *Server) handleRoomConsoleCommand(room *room.Room, requestBody []byte) error {
	var rcc RoomConsoleCommand

	if err := json.Unmarshal(requestBody, &rcc); err != nil {
		return err
	}

	switch rcc.CommandType {
	case CommandTypeSetField:
		gameId, hasGameId := rcc.Payload["gameId"]
		words, hasWords := rcc.Payload["words"]

		if !hasGameId || !hasWords {
			return errors.New("missing game id or words")
		}

		g := room.GetGame(gameId)

		if err := g.GetField().Decode64(strings.Split(words, " ")); err != nil {
			return err
		}
		break
	default:
		break
	}

	return nil
}
