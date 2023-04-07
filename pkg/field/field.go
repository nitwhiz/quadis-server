package field

import (
	"github.com/nitwhiz/quadis-server/pkg/dirty"
	"github.com/nitwhiz/quadis-server/pkg/piece"
	"github.com/nitwhiz/quadis-server/pkg/rng"
	"strings"
	"sync"
)

type Settings struct {
	Seed   int64
	Width  int
	Height int
}

type Field struct {
	data    []piece.Token
	centerX int
	// todo: this is more or less a second source of truth
	currentBedrock int
	Dirty          *dirty.Dirtiness
	mu             *sync.RWMutex
	random         *rng.Basic
	width          int
	height         int
}

type Payload struct {
	Data string `json:"data"`
}

func New(settings *Settings) *Field {
	return &Field{
		data:    make([]piece.Token, settings.Width*settings.Height),
		centerX: settings.Width/2 - piece.BodyWidth/2,
		Dirty:   dirty.New(),
		mu:      &sync.RWMutex{},
		random:  rng.NewBasic(settings.Seed),
		width:   settings.Width,
		height:  settings.Height,
	}
}

func (f *Field) ToPayload() *Payload {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return &Payload{
		Data: strings.Join(f.Encode64(), " "),
	}
}

func (f *Field) Lock() {
	f.mu.Lock()
}

func (f *Field) Unlock() {
	f.mu.Unlock()
}

func (f *Field) GetCurrentBedrock() int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.currentBedrock
}

func (f *Field) GetHeight() int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.height
}

func (f *Field) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.data = make([]piece.Token, f.width*f.height)

	f.Dirty.Trip()
}

func (f *Field) isInBounds(x int, y int) bool {
	if x < 0 || x >= f.width || y < 0 || y >= f.height {
		return false
	}

	return true
}

func (f *Field) ClearLines() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	cleared := 0

	for y := 0; y < f.height; y++ {
		isFull := true

		for x := 0; x < f.width; x++ {
			d := f.getDataXY(x, y)

			if d == piece.TokenNone || d == piece.TokenBedrock {
				isFull = false
				break
			}
		}

		if isFull {
			cleared++

			for yi := y; yi > 0; yi-- {
				for x := 0; x < f.width; x++ {
					if yi == 0 {
						f.setDataXY(x, yi, 0, true)
					} else {
						f.setDataXY(x, yi, f.getDataXY(x, yi-1), true)
					}
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

func (f *Field) shouldSetDataAt(i int, d piece.Token) bool {
	if i < 0 || i >= f.width*f.height {
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

func (f *Field) setDataAt(i int, d piece.Token, force bool) {
	if force || f.shouldSetDataAt(i, d) {
		f.data[i] = d
		f.Dirty.Trip()
	}
}

func (f *Field) setDataXY(x int, y int, d piece.Token, force bool) {
	f.setDataAt(y*f.width+x, d, force)
}

func (f *Field) getDataAt(i int) piece.Token {
	return f.data[i]
}

func (f *Field) getDataXY(x int, y int) piece.Token {
	return f.getDataAt(y*f.width + x)
}

func (f *Field) PutPiece(p *piece.Piece, r piece.Rotation, x int, y int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for px := 0; px < piece.BodyWidth; px++ {
		for py := 0; py < piece.BodyWidth; py++ {
			d := p.GetDataXY(r, px, py)

			if d != 0 {
				f.setDataXY(x+px, y+py, d, false)
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

	for y := 0; y < f.height; y++ {
		for x := 0; x < f.width; x++ {
			blockData := f.getDataXY(x, y)

			if y != 0 && f.isInBounds(x, y-delta) {
				f.setDataXY(x, y-delta, blockData, true)
			}
		}
	}

	target := f.currentBedrock + delta

	for y := f.currentBedrock; y < target; y++ {
		for x := 0; x < f.width; x++ {
			f.setDataXY(x, f.height-y-1, piece.TokenBedrock, true)
		}
	}

	f.currentBedrock = target
}

func (f *Field) decreaseBedrock(delta int) {
	target := f.currentBedrock - delta

	for f.currentBedrock > 0 && f.currentBedrock > target {
		for y := f.height - 1; y >= 0; y-- {
			for x := 0; x < f.width; x++ {
				blockData := f.getDataXY(x, y)

				if y != f.height-1 {
					f.setDataXY(x, y+1, blockData, true)
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

func (f *Field) getBottomArea(startY int) []piece.Token {
	res := make([]piece.Token, (f.height-startY)*f.width)

	for fy := startY; fy < f.height; fy++ {
		for fx := 0; fx < f.width; fx++ {
			res[(fy-startY)*f.width+fx] = f.getDataXY(fx, fy)
		}
	}

	return res
}

func (f *Field) putData(d []piece.Token) {
	for fy := 0; fy < f.height; fy++ {
		for fx := 0; fx < f.width; fx++ {
			f.setDataXY(fx, fy, d[fy*f.width+fx], true)
		}
	}
}

func (f *Field) ShuffleTokens() {
	f.mu.Lock()
	defer f.mu.Unlock()

	areaStartY := f.height - 1

	for y := f.height - 1; y >= 0; y-- {
		tokenFound := false

		for x := 0; x < f.width; x++ {
			if f.getDataXY(x, y) != piece.TokenNone {
				tokenFound = true
				break
			}
		}

		if !tokenFound {
			break
		}

		areaStartY = y
	}

	if areaStartY == f.height-1 {
		return
	}

	offset := areaStartY * f.width

	// todo: this should not be in here

	f.random.Shuffle(f.width*(f.height-areaStartY)-1, func(i int, j int) {
		if !f.random.Probably(.2) {
			return
		}

		oi := offset + i
		oj := offset + j

		yi := oi / f.width
		yj := oj / f.width

		if yi < f.height-f.currentBedrock && yj < f.height-f.currentBedrock {
			tokI := f.getDataAt(oi)
			tokJ := f.getDataAt(oj)

			f.setDataAt(oi, tokJ, false)
			f.setDataAt(oj, tokI, false)
		}
	})
}
