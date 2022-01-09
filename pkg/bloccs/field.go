package bloccs

import (
	"bloccs-server/pkg/event"
	"fmt"
	"time"
)

type Field struct {
	ID            string            `json:"id"`
	Data          FieldData         `json:"data"`
	Width         int               `json:"width"`
	Height        int               `json:"height"`
	FallingPiece  *FallingPieceData `json:"falling_piece"`
	GameOver      bool              `json:"game_over"`
	Dirty         bool
	bedrockHeight int
	lastUpdate    *time.Time
	eventBus      *event.Bus
}

func NewField(bus *event.Bus, w int, h int, id string) *Field {
	return &Field{
		ID:     id,
		Data:   make(FieldData, w*h),
		Width:  w,
		Height: h,
		FallingPiece: &FallingPieceData{
			NextPiece:    GetRandomPiece(),
			CurrentPiece: nil,
			X:            0,
			Y:            0,
			Speed:        0,
			FallTimer:    0,
			Dirty:        false,
		},
		GameOver:      false,
		Dirty:         true,
		bedrockHeight: 0,
		lastUpdate:    nil,
		eventBus:      bus,
	}
}

func (f *Field) Update() (bool, bool) {
	if !f.GameOver {
		now := time.Now()

		if f.lastUpdate != nil {
			f.Dirty = false

			delta := now.Sub(*f.lastUpdate)

			if keepGoing := f.FallingPiece.Update(f, int(delta)); !keepGoing {
				f.GameOver = true
			}

			if f.FallingPiece.Dirty {
				f.eventBus.Publish(event.New(fmt.Sprintf("game_update/%s", f.ID), EventUpdateFallingPiece, &event.Payload{
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

func (f *Field) SetBedrock(h int) {
	for y := 0; y < h; y++ {
		for x := 0; x < f.Width; x++ {
			f.SetDataXY(x, f.Height-1-y, Bedrock)
		}
	}

	f.bedrockHeight = h
}

func (f *Field) IncreaseBedrock(delta int) {
	f.bedrockHeight += delta

	if f.bedrockHeight > f.Height {
		f.bedrockHeight = f.Height
	}

	f.SetBedrock(f.bedrockHeight)
}

func (f *Field) DecreaseBedrock(delta int) {
	f.bedrockHeight -= delta

	if f.bedrockHeight < 0 {
		f.bedrockHeight = 0
	}

	f.SetBedrock(f.bedrockHeight)
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
