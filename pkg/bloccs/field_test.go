package bloccs

import (
	"fmt"
	"testing"
)

func TestField_PutPiece(t *testing.T) {
	field := NewField(6, 6)

	piece := PieceT.Clone()

	field.PutPiece(piece, 2, 2)

	expectedData := []uint8{
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0,
		0, 0, 'T', 'T', 'T', 0,
		0, 0, 0, 'T', 0, 0,
		0, 0, 0, 0, 0, 0,
	}

	fmt.Println(field.data)
	fmt.Println(expectedData)

	for x := 0; x < 6; x++ {
		for y := 0; y < 6; y++ {
			i := y*6 + x

			if field.data[i] != expectedData[i] {
				t.Fatalf("rotationStates not equal at %d;%d", x, y)
			}
		}
	}
}
