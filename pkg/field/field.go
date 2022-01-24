package field

import (
	"bloccs-server/pkg/event"
	"bloccs-server/pkg/piece"
)

type Field struct {
	Data             []uint8
	Width            int
	Height           int
	dirty            bool
	currentBedrock   int
	requestedBedrock int
	eventBus         *event.Bus
}

func New(bus *event.Bus, w int, h int) *Field {
	f := Field{
		Data:             make([]uint8, w*h),
		Width:            w,
		Height:           h,
		dirty:            true,
		requestedBedrock: 0,
		currentBedrock:   0,
		eventBus:         bus,
	}

	return &f
}

func (f *Field) IsDirty() bool {
	return f.dirty
}

func (f *Field) SetDirty(dirty bool) {
	f.dirty = dirty
}

func putPieceToSlice(buf []uint8, bufWidth int, p *piece.Piece, r int, x int, y int) {
	for px := 0; px < piece.DataWidth; px++ {
		for py := 0; py < piece.DataWidth; py++ {
			d := p.GetDataXY(r, px, py)

			if d != 0 {
				id := (py+y)*bufWidth + (px + x)

				buf[id] = d
			}
		}
	}
}

func (f *Field) GetCenterX() int {
	return f.Width/2 - piece.DataWidth/2
}

func (f *Field) PutPiece(p *piece.Piece, r int, x int, y int) {
	putPieceToSlice(f.Data, f.Width, p, r, x, y)

	f.dirty = true
}

func (f *Field) isInBounds(x int, y int) bool {
	if x < 0 || x >= f.Width || y < 0 || y >= f.Height {
		return false
	}

	return true
}

func (f *Field) CanPutPiece(p *piece.Piece, r int, x int, y int) bool {
	for px := 0; px < piece.DataWidth; px++ {
		for py := 0; py < piece.DataWidth; py++ {
			tx := px + x
			ty := py + y

			if p.GetDataXY(r, px, py) != 0 && (!f.isInBounds(tx, ty) || f.Data[ty*f.Width+tx] != 0) {
				return false
			}
		}
	}

	return true
}

// ApplyBedrock returns if there is a bedrock induced game over
func (f *Field) ApplyBedrock() bool {
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

		f.dirty = true
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
			f.SetDataXY(x, f.Height-f.currentBedrock, piece.Bedrock)
		}

		f.dirty = true
	}

	return false
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

			if d == 0 || d == piece.Bedrock {
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
		f.dirty = true
	}

	return cleared
}

func (f *Field) SetDataXY(x int, y int, d uint8) {
	f.Data[y*f.Width+x] = d
}

func (f *Field) GetDataXY(x int, y int) uint8 {
	return f.Data[y*f.Width+x]
}
