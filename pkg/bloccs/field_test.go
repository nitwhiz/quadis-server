package bloccs

import (
	"bloccs-server/pkg/event"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func printField(data PieceData, width int, height int) {
	fmt.Print("┌───┬" + strings.Repeat("─", width*3+6) + "─┐\n")

	for y := 0; y < height; y++ {
		fmt.Printf("│ %d │", y)

		for x := 0; x < width; x++ {
			i := y*6 + x
			c := data[i]
			s := ""

			if c == 0 {
				s = "."
			} else {
				s = string(c)
			}

			if c < 33 || c > 126 {
				s = strconv.Itoa(int(c))
			}

			fmt.Printf(" %3s", s)
		}

		fmt.Print(" │\n")
	}

	fmt.Print("│   └" + strings.Repeat("─", width*3+7) + "┤\n│    ")

	for x := 0; x < width; x++ {
		fmt.Printf(" %3d", x)
	}

	fmt.Print(" │\n└──" + strings.Repeat("─", width*3+8) + "─┘\n")
}

func TestField_PutPiece(t *testing.T) {
	var T = tPiece.Name

	bus := event.NewBus()

	bus.Start()

	fieldWidth := 6
	fieldHeight := 6

	field := NewField(bus, "test", fieldWidth, fieldHeight)

	p := NewPiece(&tPiece)

	field.PutPiece(p, 0, 2, 2)

	expectedData := PieceData{
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, T, T, T, 0,
		0, 0, 0, T, 0, 0,
		0, 0, 0, 0, 0, 0,
	}

	for x := 0; x < fieldWidth; x++ {
		for y := 0; y < fieldHeight; y++ {
			i := y*fieldWidth + x

			if field.Data[i] != expectedData[i] {
				fmt.Println("field:")
				printField(field.Data, fieldWidth, fieldHeight)

				fmt.Println("expected field:")
				printField(expectedData, fieldWidth, fieldHeight)

				t.Fatalf("fields not equal at %d;%d", x, y)
			}
		}
	}

	bus.Stop()
}

func TestField_PutPieceEmitsEvent(t *testing.T) {
	bus := event.NewBus()

	bus.Start()

	eventCount := 0
	expectedEventCount := 1

	field := NewField(bus, "test", 6, 6)

	bus.Subscribe(EventFieldUpdate, func(event *event.Event) {
		if f, _ := event.Source.(*Field); f != field {
			t.Fatalf("event source is not field")
		}

		eventCount += 1
	})

	p := NewPiece(&tPiece)

	field.PutPiece(p, 0, 2, 2)

	bus.Stop()

	if eventCount != expectedEventCount {
		t.Fatalf("missing events. exepected %d, got %d", expectedEventCount, eventCount)
	}
}

func TestField_PutPieceAtBottom(t *testing.T) {
	var T = tPiece.Name

	bus := event.NewBus()

	bus.Start()

	fieldWidth := 6
	fieldHeight := 6

	field := NewField(bus, "test", fieldWidth, fieldHeight)

	p := NewPiece(&tPiece)

	field.PutPiece(p, 0, 2, 3)

	expectedData := PieceData{
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, T, T, T, 0,
		0, 0, 0, T, 0, 0,
	}

	for x := 0; x < fieldWidth; x++ {
		for y := 0; y < fieldHeight; y++ {
			i := y*fieldWidth + x

			if field.Data[i] != expectedData[i] {
				fmt.Println("field:")
				printField(field.Data, fieldWidth, fieldHeight)

				fmt.Println("expected field:")
				printField(expectedData, fieldWidth, fieldHeight)

				t.Fatalf("fields not equal at %d;%d", x, y)
			}
		}
	}

	bus.Stop()
}
