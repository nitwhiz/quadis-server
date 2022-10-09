package field

import (
	"github.com/nitwhiz/quadis-server/pkg/dirty"
	"github.com/nitwhiz/quadis-server/pkg/piece"
	"sync"
)

const Width = 10
const Height = 20

type Field struct {
	data           []piece.Token
	centerX        int
	currentBedrock int
	Dirty          *dirty.Dirtiness
	mu             *sync.RWMutex
}

type Payload struct {
	Data piece.Tokens `json:"data"`
}

func New() *Field {
	return &Field{
		data:    make([]piece.Token, Width*Height),
		centerX: Width/2 - piece.BodyWidth/2,
		Dirty:   dirty.New(),
		mu:      &sync.RWMutex{},
	}
}

func (f *Field) ToPayload() *Payload {
	defer f.mu.RUnlock()
	f.mu.RLock()

	return &Payload{
		Data: f.data,
	}
}

func (f *Field) GetCurrentBedrock() int {
	defer f.mu.RUnlock()
	f.mu.RLock()

	return f.currentBedrock
}

func (f *Field) Reset() {
	defer f.mu.Unlock()
	f.mu.Lock()

	f.data = make([]piece.Token, Width*Height)

	f.Dirty.Toggle()
}

func (f *Field) isInBounds(x int, y int) bool {
	if x < 0 || x >= Width || y < 0 || y >= Height {
		return false
	}

	return true
}

func (f *Field) ClearLines() int {
	defer f.mu.Unlock()
	f.mu.Lock()

	cleared := 0

	for y := 0; y < Height; y++ {
		isFull := true

		for x := 0; x < Width; x++ {
			d := f.getDataXY(x, y)

			if d == 0 || d == piece.TokenBedrock {
				isFull = false
				break
			}
		}

		if isFull {
			cleared++

			for yi := y; yi > 0; yi-- {
				for x := 0; x < Width; x++ {
					if yi == 0 {
						f.setDataXY(x, yi, 0)
					}

					f.setDataXY(x, yi, f.getDataXY(x, yi-1))
				}
			}

			y--
		}
	}

	if cleared > 0 {
		f.decreaseBedrock(cleared)
	}

	return cleared
}

func (f *Field) setDataXY(x int, y int, d piece.Token) {
	i := y*Width + x

	if f.data[i] != d {
		f.data[i] = d
		f.Dirty.Toggle()
	}
}

func (f *Field) getDataXY(x int, y int) piece.Token {
	return f.data[y*Width+x]
}

func (f *Field) PutPiece(p *piece.Piece, r piece.Rotation, x int, y int) {
	defer f.mu.Unlock()
	f.mu.Lock()

	for px := 0; px < piece.BodyWidth; px++ {
		for py := 0; py < piece.BodyWidth; py++ {
			d := p.GetDataXY(r, px, py)

			if d != 0 {
				f.setDataXY(x+px, y+py, d)
			}
		}
	}
}

func (f *Field) CanPutPiece(p *piece.Piece, r piece.Rotation, x int, y int) bool {
	defer f.mu.RUnlock()
	f.mu.RLock()

	for px := 0; px < piece.BodyWidth; px++ {
		for py := 0; py < piece.BodyWidth; py++ {
			tx := px + x
			ty := py + y

			if p.GetDataXY(r, px, py) != 0 && (!f.isInBounds(tx, ty) || f.getDataXY(tx, ty) != 0) {
				return false
			}
		}
	}

	return true
}

// IncreaseBedrock increases the bedrock level as far as possible, even moving pieces out over the top border
func (f *Field) IncreaseBedrock(delta int) {
	defer f.mu.Unlock()
	f.mu.Lock()

	target := f.currentBedrock + delta

	for f.currentBedrock < target {
		for y := 0; y < Height; y++ {
			for x := 0; x < Width; x++ {
				blockData := f.getDataXY(x, y)

				if y != 0 && f.isInBounds(x, y-1) {
					f.setDataXY(x, y-1, blockData)
				}
			}
		}

		f.currentBedrock++

		for x := 0; x < Width; x++ {
			f.setDataXY(x, Height-f.currentBedrock, piece.TokenBedrock)
		}
	}
}

func (f *Field) decreaseBedrock(delta int) {
	target := f.currentBedrock - delta

	for f.currentBedrock > 0 && f.currentBedrock > target {
		for y := Height - 1; y >= 0; y-- {
			for x := 0; x < Width; x++ {
				blockData := f.getDataXY(x, y)

				if y != Height-1 {
					f.setDataXY(x, y+1, blockData)
				}
			}
		}

		f.currentBedrock--
	}
}

func (f *Field) GetCenterX() int {
	defer f.mu.RUnlock()
	f.mu.RLock()

	return f.centerX
}
