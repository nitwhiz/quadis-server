package bloccs

import (
	"bloccs-server/pkg/event"
	"testing"
	"time"
)

func TestGame_StartStop(t *testing.T) {
	bus := event.NewBus()

	bus.Start()

	game := NewGame(bus, &GameSettings{
		FallingPieceSpeed: 1,
		Seed:              "D34D-B33F",
	})

	game.Start()

	game.Stop()
}

func TestGame_FallingPieceMovement(t *testing.T) {
	bus := event.NewBus()

	bus.Start()

	fallSpeed := 70.0

	settings := &GameSettings{
		FallingPieceSpeed: fallSpeed,
		Seed:              "D34D-B33F",
		FieldWidth:        6,
		FieldHeight:       6,
	}

	game := NewGame(bus, settings)

	game.Start()

	pieceStepTime := time.Millisecond * time.Duration(1000.0/fallSpeed)

	// wait 5 more milliseconds, just for good measure
	time.Sleep(pieceStepTime*2 + time.Millisecond*5)

	pos1 := game.fallingPiece.Y

	time.Sleep(pieceStepTime + time.Millisecond*5)

	pos2 := game.fallingPiece.Y

	game.Stop()

	if pos1 != 1 || pos2 != 2 {
		t.Fatalf("falling piece not moved far enough")
	}
}
