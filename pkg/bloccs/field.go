package bloccs

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const Bedrock = 'B'

type FieldData []uint8

func (d FieldData) MarshalJSON() ([]byte, error) {
	var result string

	if d == nil {
		result = "null"
	} else {
		result = strings.Join(strings.Fields(fmt.Sprintf("%d", d)), ",")
	}

	return []byte(result), nil
}

type FallingPiece struct {
	Piece     *Piece `json:"piece"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Speed     int    `json:"speed"`
	FallTimer int    `json:"fall_timer"`
}

// info: field has a custom json marshaller

type Field struct {
	data         FieldData
	width        int
	height       int
	fallingPiece *FallingPiece
	lastUpdate   *time.Time
	eventBus     *EventBus
	gameOver     bool
	Dirty        bool
}

func NewField(bus *EventBus, w int, h int) *Field {
	return &Field{
		data:         make(FieldData, w*h),
		width:        w,
		height:       h,
		fallingPiece: nil,
		lastUpdate:   nil,
		eventBus:     bus,
		gameOver:     false,
		Dirty:        true,
	}
}

func (f *Field) Update() {
	if f.gameOver {
		return
	}

	now := time.Now()

	if f.lastUpdate != nil {
		f.Dirty = false

		delta := now.Sub(*f.lastUpdate)

		if f.fallingPiece == nil {
			f.SetFallingPiece(GetRandomPiece())
		} else {
			f.fallingPiece.FallTimer -= int(delta / time.Millisecond)

			if f.fallingPiece.FallTimer <= 0 {
				f.fallingPiece.FallTimer = 1000 / f.fallingPiece.Speed

				if m := f.MoveFallingPiece(0, 1, 0); !m {
					f.LockFallingPiece()
				}
			}
		}
	}

	f.lastUpdate = &now
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
	return f.width/2 - PieceBufWidth/2
}

func (f *Field) PutPiece(p *Piece, x int, y int) {
	putPieceToSlice(f.data, f.width, p, x, y)

	f.Dirty = true
}

func (f *Field) SetFallingPiece(p *Piece) {
	f.Dirty = true

	if p == nil {
		f.fallingPiece = nil

		return
	}

	f.fallingPiece = &FallingPiece{
		Piece:     p,
		X:         f.GetCenterX(),
		Y:         0,
		Speed:     1,
		FallTimer: 1000,
	}
}

func (f *Field) LockFallingPiece() {
	if f.fallingPiece != nil {
		f.PutPiece(f.fallingPiece.Piece, f.fallingPiece.X, f.fallingPiece.Y)
	}

	f.fallingPiece = nil

	cleared := f.ClearFullRows()

	if cleared != 0 {
		f.eventBus.Publish(&Event{
			Type: EventRowsCleared,
			Data: map[string]interface{}{
				"count": cleared,
			},
		})
	}

	f.SetFallingPiece(GetRandomPiece())

	if nm := f.CanMoveFallingPiece(0, 1, 0); !nm {
		f.eventBus.Publish(&Event{
			Type: EventGameOver,
		})

		f.gameOver = true
	}
}

func (f *Field) isInBounds(x int, y int) bool {
	if x < 0 || x >= f.width || y < 0 || y >= f.height {
		return false
	}

	return true
}

func (f *Field) canPutPiece(p *Piece, x int, y int) bool {
	for px := 0; px < PieceBufWidth; px++ {
		for py := 0; py < PieceBufWidth; py++ {
			tx := px + x
			ty := py + y

			if p.GetDataXY(px, py) != 0 && (!f.isInBounds(tx, ty) || f.data[ty*f.width+tx] != 0) {
				return false
			}
		}
	}

	return true
}

func (f *Field) CanMoveFallingPiece(dx int, dy int, dr int) bool {
	if f.fallingPiece == nil {
		return false
	}

	tp := f.fallingPiece.Piece.Clone()

	for ; dr > 0; dr-- {
		tp.Rotate()
	}

	return f.canPutPiece(tp, f.fallingPiece.X+dx, f.fallingPiece.Y+dy)
}

func (f *Field) PunchFallingPiece() {
	if f.fallingPiece == nil {
		return
	}

	for {
		if m := f.MoveFallingPiece(0, 1, 0); !m {
			break
		}
	}

	f.LockFallingPiece()
}

func (f *Field) MoveFallingPiece(dx int, dy int, dr int) bool {
	if f.fallingPiece != nil && f.CanMoveFallingPiece(dx, dy, dr) {
		f.fallingPiece.X += dx
		f.fallingPiece.Y += dy

		for ; dr > 0; dr-- {
			f.fallingPiece.Piece.Rotate()
		}

		f.Dirty = true

		return true
	}

	return false
}

func (f *Field) SetBedrock(h int) {
	for y := 0; y < h; y++ {
		for x := 0; x < f.width; x++ {
			f.SetDataXY(x, f.height-1-y, Bedrock)
		}
	}
}

func (f *Field) ClearFullRows() int {
	cleared := 0

	for y := 0; y < f.height; y++ {
		isFull := true

		for x := 0; x < f.width; x++ {
			d := f.GetDataXY(x, y)

			if d == 0 || d == Bedrock {
				isFull = false
				break
			}
		}

		if isFull {
			cleared++

			for yi := y; yi > 0; yi-- {
				for x := 0; x < f.width; x++ {
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
	f.data[y*f.width+x] = d
}

func (f *Field) GetDataXY(x int, y int) uint8 {
	return f.data[y*f.width+x]
}

func (f *Field) GetData() FieldData {
	var buf FieldData

	buf = append(buf, f.data...)

	if f.fallingPiece != nil {
		putPieceToSlice(buf, f.width, f.fallingPiece.Piece, f.fallingPiece.X, f.fallingPiece.Y)
	}

	return buf
}

func (f *Field) MarshalJSON() ([]byte, error) {
	d := map[string]interface{}{
		"data":   f.GetData(),
		"width":  f.width,
		"height": f.height,
	}

	return json.Marshal(d)
}
