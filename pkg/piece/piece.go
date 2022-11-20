package piece

import (
	"fmt"
	"strings"
)

type Token uint8
type Body []Token
type Tokens = Body
type Rotation int

const BodyWidth = 4

const TokenNone = Token(0)

const TokenI = Token(1)
const TokenO = Token(2)
const TokenL = Token(3)
const TokenJ = Token(4)
const TokenS = Token(5)
const TokenT = Token(6)
const TokenZ = Token(7)

const TokenBedrock = Token(8)

const MaxToken = TokenBedrock

type Piece struct {
	Token        Token
	rotatedFaces *[]Body
}

type Payload struct {
	Token Token `json:"token"`
}

func (d Tokens) MarshalJSON() ([]byte, error) {
	var result string

	if d == nil {
		result = "null"
	} else {
		result = strings.Join(strings.Fields(fmt.Sprintf("%d", d)), ",")
	}

	return []byte(result), nil
}

func (p *Piece) ClampRotation(rot Rotation) Rotation {
	return rot % Rotation(len(*p.rotatedFaces))
}

func (p *Piece) GetData(rot Rotation) *Body {
	return &((*p.rotatedFaces)[p.ClampRotation(rot)])
}

func (p *Piece) GetDataXY(rot Rotation, x int, y int) Token {
	return (*p.rotatedFaces)[p.ClampRotation(rot)][y*BodyWidth+x]
}

var I = Piece{
	Token: TokenI,
	rotatedFaces: &[]Body{
		{
			0, 0, 0, 0,
			0, 0, 0, 0,
			TokenI, TokenI, TokenI, TokenI,
			0, 0, 0, 0,
		},
		{
			0, 0, TokenI, 0,
			0, 0, TokenI, 0,
			0, 0, TokenI, 0,
			0, 0, TokenI, 0,
		},
	},
}

var O = Piece{
	Token: TokenO,
	rotatedFaces: &[]Body{
		{
			0, 0, 0, 0,
			0, TokenO, TokenO, 0,
			0, TokenO, TokenO, 0,
			0, 0, 0, 0,
		},
	},
}

var L = Piece{
	Token: TokenL,
	rotatedFaces: &[]Body{
		{
			0, 0, 0, 0,
			TokenL, TokenL, TokenL, 0,
			0, 0, TokenL, 0,
			0, 0, 0, 0,
		},
		{
			0, TokenL, 0, 0,
			0, TokenL, 0, 0,
			TokenL, TokenL, 0, 0,
			0, 0, 0, 0,
		},
		{
			TokenL, 0, 0, 0,
			TokenL, TokenL, TokenL, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
		},
		{
			0, TokenL, TokenL, 0,
			0, TokenL, 0, 0,
			0, TokenL, 0, 0,
			0, 0, 0, 0,
		},
	},
}

var J = Piece{
	Token: TokenJ,
	rotatedFaces: &[]Body{
		{
			0, 0, 0, 0,
			TokenJ, TokenJ, TokenJ, 0,
			TokenJ, 0, 0, 0,
			0, 0, 0, 0,
		},
		{
			TokenJ, TokenJ, 0, 0,
			0, TokenJ, 0, 0,
			0, TokenJ, 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 0, TokenJ, 0,
			TokenJ, TokenJ, TokenJ, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
		},
		{
			0, TokenJ, 0, 0,
			0, TokenJ, 0, 0,
			0, TokenJ, TokenJ, 0,
			0, 0, 0, 0,
		},
	},
}

var S = Piece{
	Token: TokenS,
	rotatedFaces: &[]Body{
		{
			0, 0, 0, 0,
			0, TokenS, TokenS, 0,
			TokenS, TokenS, 0, 0,
			0, 0, 0, 0,
		},
		{
			0, TokenS, 0, 0,
			0, TokenS, TokenS, 0,
			0, 0, TokenS, 0,
			0, 0, 0, 0,
		},
		{
			0, 0, 0, 0,
			0, TokenS, TokenS, 0,
			TokenS, TokenS, 0, 0,
			0, 0, 0, 0,
		},
		{
			0, TokenS, 0, 0,
			0, TokenS, TokenS, 0,
			0, 0, TokenS, 0,
			0, 0, 0, 0,
		},
	},
}

var T = Piece{
	Token: TokenT,
	rotatedFaces: &[]Body{
		{
			0, 0, 0, 0,
			TokenT, TokenT, TokenT, 0,
			0, TokenT, 0, 0,
			0, 0, 0, 0,
		},
		{
			0, TokenT, 0, 0,
			TokenT, TokenT, 0, 0,
			0, TokenT, 0, 0,
			0, 0, 0, 0,
		},
		{
			0, TokenT, 0, 0,
			TokenT, TokenT, TokenT, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
		},
		{
			0, TokenT, 0, 0,
			0, TokenT, TokenT, 0,
			0, TokenT, 0, 0,
			0, 0, 0, 0,
		},
	},
}

var Z = Piece{
	Token: TokenZ,
	rotatedFaces: &[]Body{
		{
			0, 0, 0, 0,
			TokenZ, TokenZ, 0, 0,
			0, TokenZ, TokenZ, 0,
			0, 0, 0, 0,
		},
		{
			0, 0, TokenZ, 0,
			0, TokenZ, TokenZ, 0,
			0, TokenZ, 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 0, 0, 0,
			TokenZ, TokenZ, 0, 0,
			0, TokenZ, TokenZ, 0,
			0, 0, 0, 0,
		},
		{
			0, 0, TokenZ, 0,
			0, TokenZ, TokenZ, 0,
			0, TokenZ, 0, 0,
			0, 0, 0, 0,
		},
	},
}

var All = []*Piece{
	&T,
	&L,
	&I,
	&J,
	&O,
	&S,
	&Z,
}
