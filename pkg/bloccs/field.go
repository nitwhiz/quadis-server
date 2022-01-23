package bloccs

import (
	"bloccs-server/pkg/event"
	"fmt"
	"time"
)

// todo: add methods for getting/setting everything in here to prevent data races

type Field struct {
	PlayerID         string            `json:"id"`
	Data             FieldData         `json:"data"`
	Width            int               `json:"width"`
	Height           int               `json:"height"`
	FallingPiece     *FallingPieceData `json:"falling_piece"`
	GameOver         bool              `json:"game_over"`
	Dirty            bool
	rng              *RNG
	currentBedrock   int
	requestedBedrock int
	lastUpdate       *time.Time
	eventBus         *event.Bus
}

func NewField(bus *event.Bus, rng *RNG, w int, h int, playerId string) *Field {
	f := Field{
		PlayerID: playerId,
		Data:     make(FieldData, w*h),
		Width:    w,
		Height:   h,
		FallingPiece: &FallingPieceData{
			X:         0,
			Y:         0,
			Speed:     0,
			FallTimer: 0,
			Dirty:     false,
			HoldLock:  false,
		},
		GameOver:         false,
		Dirty:            true,
		rng:              rng,
		requestedBedrock: 0,
		currentBedrock:   0,
		lastUpdate:       nil,
		eventBus:         bus,
	}

	f.FallingPiece.NextPiece = rng.NextPiece()

	return &f
}

func (f *Field) Update() (bool, bool) {
	if !f.GameOver {
		now := time.Now()

		if f.lastUpdate != nil {
			delta := now.Sub(*f.lastUpdate)

			if keepGoing := f.FallingPiece.Update(f, int(delta)); !keepGoing {
				f.GameOver = true
			}

			if f.FallingPiece.Dirty {
				f.eventBus.Publish(event.New(fmt.Sprintf("update/%s", f.PlayerID), EventUpdateFallingPiece, &event.Payload{
					"falling_piece_data": f.FallingPiece,
					"piece_display":      f.FallingPiece.CurrentPiece.GetData(),
				}))
			}
		}

		f.lastUpdate = &now
	}

	return f.Dirty, f.GameOver
}

func putPieceToSlice(buf []uint8, bufWidth int, p *Piece, x int, y int) {
	for px := 0; px < PieceBufWidth; px++ {
		for py := 0; py < PieceBufWidth; py++ {
			d := p.GetDataXY(px, py)

			if d != 0 {
				id := (py+y)*bufWidth + (px + x)

				buf[id] = d
			}
		}
	}
}

func (f *Field) GetCenterX() int {
	return f.Width/2 - PieceBufWidth/2
}

func (f *Field) PutPiece(p *Piece, x int, y int) {
	putPieceToSlice(f.Data, f.Width, p, x, y)

	f.Dirty = true
}

func (f *Field) isInBounds(x int, y int) bool {
	if x < 0 || x >= f.Width || y < 0 || y >= f.Height {
		return false
	}

	return true
}

func (f *Field) canPutPiece(p *Piece, x int, y int) bool {
	for px := 0; px < PieceBufWidth; px++ {
		for py := 0; py < PieceBufWidth; py++ {
			tx := px + x
			ty := py + y

			if p.GetDataXY(px, py) != 0 && (!f.isInBounds(tx, ty) || f.Data[ty*f.Width+tx] != 0) {
				return false
			}
		}
	}

	return true
}

func (f *Field) ApplyBedrock() {
	// todo: what to do if bedrock crashes into the piece?

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
	}

	for f.currentBedrock < f.requestedBedrock {
		for y := 0; y < f.Height; y++ {
			for x := 0; x < f.Width; x++ {
				blockData := f.GetDataXY(x, y)

				if blockData != 0 && y == 0 {
					f.GameOver = true
					return
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
	}

	f.Dirty = true
}

func (f *Field) IncreaseBedrock(delta int) {
	f.requestedBedrock += delta

	if f.requestedBedrock > f.Height {
		f.requestedBedrock = f.Height
	}

	f.ApplyBedrock()
}

func (f *Field) DecreaseBedrock(delta int) {
	f.requestedBedrock -= delta

	if f.requestedBedrock < 0 {
		f.requestedBedrock = 0
	}

	f.ApplyBedrock()
}

func (f *Field) ClearFullRows() int {
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
		f.Dirty = true
	}

	return cleared
}

func (f *Field) SetDataXY(x int, y int, d uint8) {
	f.Data[y*f.Width+x] = d
}

func (f *Field) GetDataXY(x int, y int) uint8 {
	return f.Data[y*f.Width+x]
}
