package bloccs

const PieceBufWidth = 4

type Piece struct {
	Name           uint8 `json:"name"`
	Rotation       int   `json:"rotation"`
	rotationStates *[]FieldData
}

func NewPiece(p *Piece) *Piece {
	res := *p
	return &res
}

func (p *Piece) Rotate() {
	p.Rotation++

	if p.Rotation >= len(*p.rotationStates) {
		p.Rotation = 0
	}
}

func (p *Piece) GetData() *FieldData {
	return &((*p.rotationStates)[p.Rotation])
}

func (p *Piece) GetDataXY(x int, y int) uint8 {
	return (*p.rotationStates)[p.Rotation][y*PieceBufWidth+x]
}

var PieceI = Piece{
	Name: 'I',
	rotationStates: &[]FieldData{
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
	Name: 'O',
	rotationStates: &[]FieldData{
		{
			0, 0, 0, 0,
			0, 'O', 'O', 0,
			0, 'O', 'O', 0,
			0, 0, 0, 0,
		},
	},
}

var PieceL = Piece{
	Name: 'L',
	rotationStates: &[]FieldData{
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
	Name: 'J',
	rotationStates: &[]FieldData{
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
	Name: 'S',
	rotationStates: &[]FieldData{
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
	Name: 'T',
	rotationStates: &[]FieldData{
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
	Name: 'Z',
	rotationStates: &[]FieldData{
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

var Pieces = []*Piece{
	&PieceT,
	&PieceL,
	&PieceI,
	&PieceJ,
	&PieceO,
	&PieceS,
	&PieceZ,
}
