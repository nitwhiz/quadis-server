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
	f.mu.RLock()
	defer f.mu.RUnlock()

	return &Payload{
		Data: f.data,
	}
}

func (f *Field) GetCurrentBedrock() int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.currentBedrock
}

func (f *Field) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.data = make([]piece.Token, Width*Height)

	f.Dirty.Trip()
}

func (f *Field) isInBounds(x int, y int) bool {
	if x < 0 || x >= Width || y < 0 || y >= Height {
		return false
	}

	return true
}

func (f *Field) ClearLines() int {
	f.mu.Lock()
	defer f.mu.Unlock()

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

func (f *Field) shouldSetDataXY(i int, d piece.Token) bool {
	if i < 0 && i >= Width*Height {
		return false
	}

	if d != piece.TokenNone && f.data[i] == piece.TokenBedrock {
		return false
	}

	if f.data[i] == d {
		return false
	}

	return true
}

func (f *Field) setDataXY(x int, y int, d piece.Token) {
	i := y*Width + x

	if f.shouldSetDataXY(i, d) {
		f.data[i] = d
		f.Dirty.Trip()
	}
}

func (f *Field) getDataXY(x int, y int) piece.Token {
	return f.data[y*Width+x]
}

func (f *Field) PutPiece(p *piece.Piece, r piece.Rotation, x int, y int) {
	f.mu.Lock()
	defer f.mu.Unlock()

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
	f.mu.RLock()
	defer f.mu.RUnlock()

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
	f.mu.Lock()
	defer f.mu.Unlock()

	for y := 0; y < Height; y++ {
		for x := 0; x < Width; x++ {
			blockData := f.getDataXY(x, y)

			if y != 0 && f.isInBounds(x, y-delta) {
				f.setDataXY(x, y-delta, blockData)
			}
		}
	}

	target := f.currentBedrock + delta

	for y := f.currentBedrock; y < target; y++ {
		for x := 0; x < Width; x++ {
			f.setDataXY(x, Height-y-1, piece.TokenBedrock)
		}
	}

	f.currentBedrock = target
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
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.centerX
}
