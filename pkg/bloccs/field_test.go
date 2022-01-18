package bloccs

import (
	"bloccs-server/pkg/event"
	"fmt"
	"testing"
)

func TestField_PutPiece(t *testing.T) {
	bus := event.NewBus()

	field := NewField(bus, 6, 6, "test")

	piece := NewPiece(&PieceT)

	field.PutPiece(piece, 2, 2)

	expectedData := []uint8{
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 'T', 'T', 'T', 0,
		0, 0, 0, 'T', 0, 0,
		0, 0, 0, 0, 0, 0,
	}

	fmt.Println(field.Data)
	fmt.Println(expectedData)

	for x := 0; x < 6; x++ {
		for y := 0; y < 6; y++ {
			i := y*6 + x

			if field.Data[i] != expectedData[i] {
				t.Fatalf("rotationStates not equal at %d;%d", x, y)
			}
		}
	}

	bus.Stop()

	// todo: test event bus output
}
