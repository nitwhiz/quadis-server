package piece

const DataWidth = 4

const Bedrock = 'B'

type Piece struct {
	Name         uint8
	rotatedFaces *[][]uint8
}

func New(p *Piece) *Piece {
	res := *p
	return &res
}

func (p *Piece) ClampRotation(rot int) int {
	return rot % len(*p.rotatedFaces)
}

func (p *Piece) GetData(rot int) *[]uint8 {
	return &((*p.rotatedFaces)[p.ClampRotation(rot)])
}

func (p *Piece) GetDataXY(rot int, x int, y int) uint8 {
	return (*p.rotatedFaces)[p.ClampRotation(rot)][y*DataWidth+x]
}

var I = Piece{
	Name: 'I',
	rotatedFaces: &[][]uint8{
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

var O = Piece{
	Name: 'O',
	rotatedFaces: &[][]uint8{
		{
			0, 0, 0, 0,
			0, 'O', 'O', 0,
			0, 'O', 'O', 0,
			0, 0, 0, 0,
		},
	},
}

var L = Piece{
	Name: 'L',
	rotatedFaces: &[][]uint8{
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

var J = Piece{
	Name: 'J',
	rotatedFaces: &[][]uint8{
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

var S = Piece{
	Name: 'S',
	rotatedFaces: &[][]uint8{
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

var T = Piece{
	Name: 'T',
	rotatedFaces: &[][]uint8{
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

var Z = Piece{
	Name: 'Z',
	rotatedFaces: &[][]uint8{
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

var All = []*Piece{
	&T,
	&L,
	&I,
	&J,
	&O,
	&S,
	&Z,
}
