package bloccs

import (
	"math/rand"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixMilli()))

const PieceBufWidth = 4

type Piece struct {
	rotationStates *[][]uint8
	rotation       int
}

func GetRandomPiece() *Piece {
	return Pieces[rng.Intn(len(Pieces))].Clone()
}

func (p *Piece) Clone() *Piece {
	res := *p
	return &res
}

func (p *Piece) Rotate() {
	p.rotation++

	if p.rotation >= len(*p.rotationStates) {
		p.rotation = 0
	}
}

func (p *Piece) GetData() *[]uint8 {
	return &((*p.rotationStates)[p.rotation])
}

func (p *Piece) GetDataXY(x int, y int) uint8 {
	return (*p.rotationStates)[p.rotation][y*PieceBufWidth+x]
}

var PieceI = Piece{
	rotationStates: &[][]uint8{
		{
			0, 0, 0, 0,
			0, 0, 0, 0,
			'I', 'I', 'I', 'I',
			0, 0, 0, 0,
		},
		{
			0, 0, 'I', 0,
			0, 0, 'I', 0,
			0, 0, 'I', 0,
			0, 0, 'I', 0,
		},
	},
}

var PieceO = Piece{
	rotationStates: &[][]uint8{
		{
			0, 0, 0, 0,
			0, 'O', 'O', 0,
			0, 'O', 'O', 0,
			0, 0, 0, 0,
		},
	},
}

var PieceL = Piece{
	rotationStates: &[][]uint8{
		{
			0, 0, 0, 0,
			'L', 'L', 'L', 0,
			0, 0, 'L', 0,
			0, 0, 0, 0,
		},
		{
			0, 'L', 0, 0,
			0, 'L', 0, 0,
			'L', 'L', 0, 0,
			0, 0, 0, 0,
		},
		{
			'L', 0, 0, 0,
			'L', 'L', 'L', 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 'L', 'L', 0,
			0, 'L', 0, 0,
			0, 'L', 0, 0,
			0, 0, 0, 0,
		},
	},
}

var PieceJ = Piece{
	rotationStates: &[][]uint8{
		{
			0, 0, 0, 0,
			'J', 'J', 'J', 0,
			'J', 0, 0, 0,
			0, 0, 0, 0,
		},
		{
			'J', 'J', 0, 0,
			0, 'J', 0, 0,
			0, 'J', 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 0, 'J', 0,
			'J', 'J', 'J', 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 'J', 0, 0,
			0, 'J', 0, 0,
			0, 'J', 'J', 0,
			0, 0, 0, 0,
		},
	},
}

var PieceS = Piece{
	rotationStates: &[][]uint8{
		{
			0, 0, 0, 0,
			0, 'S', 'S', 0,
			'S', 'S', 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 'S', 0, 0,
			0, 'S', 'S', 0,
			0, 0, 'S', 0,
			0, 0, 0, 0,
		},
		{
			0, 0, 0, 0,
			0, 'S', 'S', 0,
			'S', 'S', 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 'S', 0, 0,
			0, 'S', 'S', 0,
			0, 0, 'S', 0,
			0, 0, 0, 0,
		},
	},
}

var PieceT = Piece{
	rotationStates: &[][]uint8{
		{
			0, 0, 0, 0,
			'T', 'T', 'T', 0,
			0, 'T', 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 'T', 0, 0,
			'T', 'T', 0, 0,
			0, 'T', 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 'T', 0, 0,
			'T', 'T', 'T', 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 'T', 0, 0,
			0, 'T', 'T', 0,
			0, 'T', 0, 0,
			0, 0, 0, 0,
		},
	},
}

var PieceZ = Piece{
	rotationStates: &[][]uint8{
		{
			0, 0, 0, 0,
			'Z', 'Z', 0, 0,
			0, 'Z', 'Z', 0,
			0, 0, 0, 0,
		},
		{
			0, 0, 'Z', 0,
			0, 'Z', 'Z', 0,
			0, 'Z', 0, 0,
			0, 0, 0, 0,
		},
		{
			0, 0, 0, 0,
			'Z', 'Z', 0, 0,
			0, 'Z', 'Z', 0,
			0, 0, 0, 0,
		},
		{
			0, 0, 'Z', 0,
			0, 'Z', 'Z', 0,
			0, 'Z', 0, 0,
			0, 0, 0, 0,
		},
	},
}

var Pieces = []Piece{
	PieceT,
	PieceL,
	PieceI,
	PieceJ,
	PieceO,
	PieceS,
	PieceZ,
}
