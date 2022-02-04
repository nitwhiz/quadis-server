package event

import (
	"fmt"
	"strings"
)

// Hello is a special event sent to the client to initiate a handshake
const Hello = "hello"

// HelloAck is a special event sent to the client to conclude the previously initiated handshake
const HelloAck = "hello_ack"

const GameOver = "game_over"
const GameStart = "game_start"

const UpdateScore = "update_score"
const UpdateField = "update_field"
const UpdateFallingPiece = "update_falling_piece"
const UpdateNextPiece = "update_next_piece"
const UpdateHoldPiece = "update_hold_piece"

const PlayerJoin = "player_join"
const PlayerLeave = "player_leave"

type PlayerPayload struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	CreateAt int64  `json:"create_at"`
}

type RoomPayload struct {
	ID      string          `json:"id"`
	Players []PlayerPayload `json:"players"`
}

type HelloAckPayload struct {
	You  PlayerPayload `json:"you"`
	Room RoomPayload   `json:"room"`
}

type UpdateScorePayload struct {
	Score int `json:"score"`
	Lines int `json:"lines"`
}

type PieceData []uint8

func (d PieceData) MarshalJSON() ([]byte, error) {
	var result string

	if d == nil {
		result = "null"
	} else {
		result = strings.Join(strings.Fields(fmt.Sprintf("%d", d)), ",")
	}

	return []byte(result), nil
}

type UpdateFieldPayload struct {
	Width  int       `json:"width"`
	Height int       `json:"height"`
	Data   PieceData `json:"data"`
}

type UpdateFallingPiecePayload struct {
	PieceName uint8 `json:"piece_name"`
	Rotation  int   `json:"rotation"`
	X         int   `json:"x"`
	Y         int   `json:"y"`
}

type UpdateNextPiecePayload struct {
	PieceName uint8 `json:"piece_name"`
}

type UpdateHoldPiecePayload struct {
	PieceName uint8 `json:"piece_name"`
}

type PlayerJoinPayload = PlayerPayload

type PlayerLeavePayload = PlayerPayload
