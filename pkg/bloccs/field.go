package bloccs

import (
	"bloccs-server/pkg/event"
	"sync"
)

const EventFieldUpdate = "field_update"

type Field struct {
	GameId           string    `json:"gameId"`
	Data             PieceData `json:"data"`
	Width            int       `json:"width"`
	Height           int       `json:"height"`
	CenterX          int       `json:"-"`
	currentBedrock   int
	requestedBedrock int
	eventBus         *event.Bus
	mu               *sync.RWMutex
}

func NewField(bus *event.Bus, id string, w int, h int) *Field {
	f := Field{
		GameId:           id,
		Data:             make(PieceData, w*h),
		Width:            w,
		Height:           h,
		CenterX:          w/2 - pieceDataWidth/2,
		requestedBedrock: 0,
		currentBedrock:   0,
		eventBus:         bus,
		mu:               &sync.RWMutex{},
	}

	return &f
}

func (f *Field) GetId() string {
	return f.GameId
}

func (f *Field) publishUpdate() {
	f.eventBus.Publish(event.New(EventFieldUpdate, f, nil))
}

func (f *Field) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.Data = make(PieceData, f.Width*f.Height)

	f.publishUpdate()

	f.requestedBedrock = 0
	f.currentBedrock = 0
}

func putPieceToSlice(buf PieceData, bufWidth int, p *Piece, r int, x int, y int) {
	for px := 0; px < pieceDataWidth; px++ {
		for py := 0; py < pieceDataWidth; py++ {
			d := p.GetDataXY(r, px, py)

			if d != 0 {
				id := (py+y)*bufWidth + (px + x)

				// todo(DATA RACE): marshalling field.data happens async to this
				buf[id] = d
			}
		}
	}
}

func (f *Field) PutPiece(p *Piece, r int, x int, y int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	putPieceToSlice(f.Data, f.Width, p, r, x, y)

	f.publishUpdate()
}

func (f *Field) isInBounds(x int, y int) bool {
	if x < 0 || x >= f.Width || y < 0 || y >= f.Height {
		return false
	}

	return true
}

func (f *Field) CanPutPiece(p *Piece, r int, x int, y int) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for px := 0; px < pieceDataWidth; px++ {
		for py := 0; py < pieceDataWidth; py++ {
			tx := px + x
			ty := py + y

			if p.GetDataXY(r, px, py) != 0 && (!f.isInBounds(tx, ty) || f.Data[ty*f.Width+tx] != 0) {
				return false
			}
		}
	}

	return true
}

// applyBedrock returns if there is a bedrock induced game over
func (f *Field) applyBedrock() bool {
	for f.currentBedrock > f.requestedBedrock {
		for y := f.Height - 1; y >= 0; y-- {
			for x := 0; x < f.Width; x++ {
				blockData := f.GetDataXY(x, y)

				if y != f.Height-1 {
					f.SetDataXY(x, y+1, blockData)
				}
			}
		}

		f.currentBedrock--

		f.publishUpdate()
	}

	for f.currentBedrock < f.requestedBedrock {
		for y := 0; y < f.Height; y++ {
			for x := 0; x < f.Width; x++ {
				blockData := f.GetDataXY(x, y)

				// todo: check if bedrock crashes into current falling piece

				if blockData != 0 && y == 0 {
					return true
				}

				if y != 0 {
					f.SetDataXY(x, y-1, blockData)
				}
			}
		}

		f.currentBedrock++

		for x := 0; x < f.Width; x++ {
			f.SetDataXY(x, f.Height-f.currentBedrock, Bedrock)
		}

		f.publishUpdate()
	}

	return false
}

func (f *Field) IncreaseBedrock(delta int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.requestedBedrock += delta

	if f.requestedBedrock > f.Height {
		f.requestedBedrock = f.Height
	}

	f.applyBedrock()
}

func (f *Field) DecreaseBedrock(delta int) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.requestedBedrock -= delta

	if f.requestedBedrock < 0 {
		f.requestedBedrock = 0
	}

	f.applyBedrock()
}

func (f *Field) ClearFullRows() int {
	f.mu.Lock()
	defer f.mu.Unlock()

	cleared := 0

	for y := 0; y < f.Height; y++ {
		isFull := true

		for x := 0; x < f.Width; x++ {
			d := f.GetDataXY(x, y)

			if d == 0 || d == Bedrock {
				isFull = false
				break
			}
		}

		if isFull {
			cleared++

			for yi := y; yi > 0; yi-- {
				for x := 0; x < f.Width; x++ {
					if yi == 0 {
						f.SetDataXY(x, yi, 0)
					}

					f.SetDataXY(x, yi, f.GetDataXY(x, yi-1))
				}
			}

			y--
		}
	}

	if cleared > 0 {
		f.publishUpdate()
	}

	return cleared
}

func (f *Field) SetDataXY(x int, y int, d uint8) {
	f.Data[y*f.Width+x] = d
}

func (f *Field) GetDataXY(x int, y int) uint8 {
	return f.Data[y*f.Width+x]
}
