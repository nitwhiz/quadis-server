package field

import (
	"fmt"
	"github.com/nitwhiz/bloccs-server/pkg/event"
	"github.com/nitwhiz/bloccs-server/pkg/piece"
	"testing"
)

func TestField_PutPiece(t *testing.T) {
	bus := event.NewBus()

	field := New(bus, 6, 6)

	p := piece.New(&piece.T)

	field.PutPiece(p, 0, 2, 2)

	expectedData := []uint8{
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 'T', 'T', 'T', 0,
		0, 0, 0, 'T', 0, 0,
		0, 0, 0, 0, 0, 0,
	}

	for x := 0; x < 6; x++ {
		for y := 0; y < 6; y++ {
			i := y*6 + x

			if field.Data[i] != expectedData[i] {
				fmt.Println(field.Data)
				fmt.Println(expectedData)

				t.Fatalf("rotationStates not equal at %d;%d", x, y)
			}
		}
	}

	bus.Stop()
}

func TestField_PutPieceAtBottom(t *testing.T) {
	bus := event.NewBus()

	field := New(bus, 6, 6)

	p := piece.New(&piece.T)

	field.PutPiece(p, 0, 2, 3)

	expectedData := []uint8{
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 'T', 'T', 'T', 0,
		0, 0, 0, 'T', 0, 0,
	}

	for x := 0; x < 6; x++ {
		for y := 0; y < 6; y++ {
			i := y*6 + x

			if field.Data[i] != expectedData[i] {
				fmt.Println(field.Data)
				fmt.Println(expectedData)

				t.Fatalf("rotationStates not equal at %d;%d", x, y)
			}
		}
	}

	bus.Stop()
}
