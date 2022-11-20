package field

import (
	"github.com/nitwhiz/quadis-server/pkg/piece"
	"testing"
)

type encode64TestData struct {
	FieldWidth    int
	FieldHeight   int
	FieldData     []piece.Token
	ExpectedWords []string
}

// these tests are made for tokens values <= 8
func getEncode64TestData() []encode64TestData {
	return []encode64TestData{
		// empty field
		{
			FieldWidth:  3,
			FieldHeight: 4,
			FieldData: []piece.Token{
				0, 0, 0,
				0, 0, 0,
				0, 0, 0,
				0, 0, 0,
			},
			ExpectedWords: []string{"0000000000000000"},
		},
		// first token
		{
			FieldWidth:  3,
			FieldHeight: 4,
			FieldData: []piece.Token{
				1, 0, 0,
				0, 0, 0,
				0, 0, 0,
				0, 0, 0,
			},
			ExpectedWords: []string{"0000100000000000"},
		},
		// last token
		{
			FieldWidth:  3,
			FieldHeight: 4,
			FieldData: []piece.Token{
				0, 0, 0,
				0, 0, 0,
				0, 0, 0,
				0, 0, 1,
			},
			ExpectedWords: []string{"0000000000000001"},
		},
		// all tokens
		{
			FieldWidth:  3,
			FieldHeight: 4,
			FieldData: []piece.Token{
				1, 2, 3,
				4, 5, 6,
				7, 8, 1,
				2, 3, 4,
			},
			ExpectedWords: []string{"0000123456781234"},
		},
		// all tokens 2
		{
			FieldWidth:  3,
			FieldHeight: 4,
			FieldData: []piece.Token{
				6, 6, 6,
				1, 6, 8,
				3, 8, 8,
				1, 3, 8,
			},
			ExpectedWords: []string{"0000666168388138"},
		},
		// bigger field
		{
			FieldWidth:  3,
			FieldHeight: 14,
			FieldData: []piece.Token{
				1, 2, 3,
				4, 5, 6,
				7, 8, 1,
				2, 3, 4,
				5, 6, 7,
				8, 1, 2,
				3, 4, 5,
				6, 7, 8,
				1, 2, 3,
				4, 5, 6,
				7, 8, 1,
				2, 3, 4,
				5, 6, 7,
				8, 1, 2,
			},
			ExpectedWords: []string{"0000001234567812", "3456781234567812", "3456781234567812"},
		},
	}
}

func TestEncode64(t *testing.T) {
	for _, d := range getEncode64TestData() {
		f := New(&Settings{
			Seed:   0,
			Width:  d.FieldWidth,
			Height: d.FieldHeight,
		})

		f.putData(d.FieldData)

		words := f.Encode64()

		if len(words) != len(d.ExpectedWords) {
			t.Fatalf("expected %d words, got %d\n", len(d.ExpectedWords), len(words))
		}

		for i, w := range words {
			if w != d.ExpectedWords[i] {
				t.Fatalf("expected word #%d ('%s') to match '%s'\n", i, w, d.ExpectedWords[i])
			}
		}
	}
}

type decode64TestData struct {
	FieldWidth        int
	FieldHeight       int
	Words             []string
	ExpectedFieldData []piece.Token
}

func getDecode64TestData() []decode64TestData {
	return []decode64TestData{
		// empty field
		{
			FieldWidth:  3,
			FieldHeight: 4,
			ExpectedFieldData: []piece.Token{
				0, 0, 0,
				0, 0, 0,
				0, 0, 0,
				0, 0, 0,
			},
			Words: []string{"0000000000000000"},
		},
		// first token
		{
			FieldWidth:  3,
			FieldHeight: 4,
			ExpectedFieldData: []piece.Token{
				1, 0, 0,
				0, 0, 0,
				0, 0, 0,
				0, 0, 0,
			},
			Words: []string{"0000100000000000"},
		},
		// last token
		{
			FieldWidth:  3,
			FieldHeight: 4,
			ExpectedFieldData: []piece.Token{
				0, 0, 0,
				0, 0, 0,
				0, 0, 0,
				0, 0, 1,
			},
			Words: []string{"0000000000000001"},
		},
		// all tokens
		{
			FieldWidth:  3,
			FieldHeight: 4,
			ExpectedFieldData: []piece.Token{
				1, 2, 3,
				4, 5, 6,
				7, 8, 1,
				2, 3, 4,
			},
			Words: []string{"0000123456781234"},
		},
		// all tokens 2
		{
			FieldWidth:  3,
			FieldHeight: 4,
			ExpectedFieldData: []piece.Token{
				6, 6, 6,
				1, 6, 8,
				3, 8, 8,
				1, 3, 8,
			},
			Words: []string{"0000666168388138"},
		},
		// bigger field
		{
			FieldWidth:  3,
			FieldHeight: 14,
			ExpectedFieldData: []piece.Token{
				1, 2, 3,
				4, 5, 6,
				7, 8, 1,
				2, 3, 4,
				5, 6, 7,
				8, 1, 2,
				3, 4, 5,
				6, 7, 8,
				1, 2, 3,
				4, 5, 6,
				7, 8, 1,
				2, 3, 4,
				5, 6, 7,
				8, 1, 2,
			},
			Words: []string{"0000001234567812", "3456781234567812", "3456781234567812"},
		},
		// ignore preceding data
		{
			FieldWidth:  3,
			FieldHeight: 4,
			ExpectedFieldData: []piece.Token{
				6, 6, 6,
				1, 6, 8,
				3, 8, 8,
				1, 3, 8,
			},
			Words: []string{"0000000000000000", "0000666168388138"},
		},
	}
}

func TestDecode64Field(t *testing.T) {
	for _, d := range getDecode64TestData() {
		f := New(&Settings{
			Seed:   0,
			Width:  d.FieldWidth,
			Height: d.FieldHeight,
		})

		if err := f.Decode64(d.Words); err != nil {
			t.Fatal(err)
		}

		for x := 0; x < d.FieldWidth; x++ {
			for y := 0; y < d.FieldHeight; y++ {
				i := y*d.FieldWidth + x

				if f.getDataXY(x, y) != d.ExpectedFieldData[i] {
					t.Errorf("field data not equal at %d,%d", x, y)
					t.Error("field    :", f.data)
					t.Error("expected :", d.ExpectedFieldData)

					t.Fatal()
				}
			}
		}
	}
}

type decode64BedrockTestData struct {
	FieldWidth      int
	FieldHeight     int
	Words           []string
	ExpectedBedrock int
}

func getDecode64BedrockTestData() []decode64BedrockTestData {
	return []decode64BedrockTestData{
		{
			FieldWidth:      3,
			FieldHeight:     4,
			Words:           []string{"0000000000000000"},
			ExpectedBedrock: 0,
		},
		{
			FieldWidth:      3,
			FieldHeight:     4,
			Words:           []string{"0000000000000888"},
			ExpectedBedrock: 1,
		},
		{
			FieldWidth:      3,
			FieldHeight:     4,
			Words:           []string{"0000000000888888"},
			ExpectedBedrock: 2,
		},
		{
			FieldWidth:      3,
			FieldHeight:     4,
			Words:           []string{"0000000888888888"},
			ExpectedBedrock: 3,
		},
		{
			FieldWidth:      3,
			FieldHeight:     4,
			Words:           []string{"0000888888888888"},
			ExpectedBedrock: 4,
		},
		{
			FieldWidth:      3,
			FieldHeight:     4,
			Words:           []string{"0000413214888888"},
			ExpectedBedrock: 2,
		},
	}
}

func TestDecode64Bedrock(t *testing.T) {
	for _, d := range getDecode64BedrockTestData() {
		f := New(&Settings{
			Seed:   0,
			Width:  d.FieldWidth,
			Height: d.FieldHeight,
		})

		if err := f.Decode64(d.Words); err != nil {
			t.Fatal(err)
		}

		if f.currentBedrock != d.ExpectedBedrock {
			t.Fatalf("expected %d bedrock, got %d", d.ExpectedBedrock, f.currentBedrock)
		}
	}
}
