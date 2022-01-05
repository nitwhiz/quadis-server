package bloccs

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const Bedrock = 'X'

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

type fallingPiece struct {
	piece     *Piece
	x         int
	y         int
	speed     int
	fallTimer int
}

type Field struct {
	data         FieldData
	width        int
	height       int
	fallingPiece *fallingPiece
	lastUpdate   *time.Time
	eventBus     *EventBus
	gameOver     bool
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
	}
}

func (f *Field) Update() {
	if f.gameOver {
		return
	}

	now := time.Now()

	if f.lastUpdate != nil {
		delta := now.Sub(*f.lastUpdate)

		if f.fallingPiece == nil {
			f.SetFallingPiece(GetRandomPiece())
		} else {
			f.fallingPiece.fallTimer -= int(delta / time.Millisecond)

			if f.fallingPiece.fallTimer <= 0 {
				f.fallingPiece.fallTimer = 1000 / f.fallingPiece.speed

				if m := f.MoveFallingPiece(0, 1, 0); !m {
					f.LockFallingPiece()

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
}

func (f *Field) SetFallingPiece(p *Piece) {
	if p == nil {
		f.fallingPiece = nil

		return
	}

	f.fallingPiece = &fallingPiece{
		piece:     p,
		x:         f.GetCenterX(),
		y:         0,
		speed:     1,
		fallTimer: 1000,
	}
}

func (f *Field) LockFallingPiece() {
	if f.fallingPiece != nil {
		f.PutPiece(f.fallingPiece.piece, f.fallingPiece.x, f.fallingPiece.y)
	}

	f.fallingPiece = nil
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

	tp := f.fallingPiece.piece.Clone()

	for ; dr > 0; dr-- {
		tp.Rotate()
	}

	return f.canPutPiece(tp, f.fallingPiece.x+dx, f.fallingPiece.y+dy)
}

func (f *Field) MoveFallingPiece(dx int, dy int, dr int) bool {
	if f.fallingPiece != nil && f.CanMoveFallingPiece(dx, dy, dr) {
		f.fallingPiece.x += dx
		f.fallingPiece.y += dy

		for ; dr > 0; dr-- {
			f.fallingPiece.piece.Rotate()
		}

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
		putPieceToSlice(buf, f.width, f.fallingPiece.piece, f.fallingPiece.x, f.fallingPiece.y)
	}

	return buf
}

func (f *Field) MarshalJSON() ([]byte, error) {
	// todo: cache this?
	d := map[string]interface{}{
		"data":   f.GetData(),
		"width":  f.width,
		"height": f.height,
	}

	return json.Marshal(d)
}
