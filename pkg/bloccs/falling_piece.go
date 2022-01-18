package bloccs

import (
	"bloccs-server/pkg/event"
	"fmt"
	"time"
)

type FallingPieceData struct {
	HoldingPiece *Piece `json:"holding_piece"`
	NextPiece    *Piece `json:"next_piece"`
	CurrentPiece *Piece `json:"current_piece"`
	X            int    `json:"x"`
	Y            int    `json:"y"`
	Speed        int    `json:"speed"`
	FallTimer    int    `json:"fall_timer"`
	Dirty        bool   `json:"-"`
}

func (p *FallingPieceData) HoldCurrentPiece(f *Field) {
	// todo: disallow holding pieces 2 times in a row
	// todo: refactor

	if p.HoldingPiece != nil {
		p.CurrentPiece, p.HoldingPiece = p.HoldingPiece, p.CurrentPiece

		p.X = f.GetCenterX()
		p.Y = 0

		// todo: configurable
		p.Speed = 1

		p.FallTimer = 1000

		p.Dirty = true
	} else {
		p.HoldingPiece = p.CurrentPiece
		p.Next(f)
	}
}

func (p *FallingPieceData) Update(f *Field, delta int) bool {
	if p.CurrentPiece == nil {
		p.Next(f)
	} else {
		p.Dirty = false

		p.FallTimer -= delta / int(time.Millisecond)

		if p.FallTimer <= 0 {
			p.FallTimer = 1000 / p.Speed

			if m := p.Move(f, 0, 1, 0); !m {
				if keepGoing := p.Lock(f); !keepGoing {
					return false
				}
			}
		}
	}

	return true
}

func (p *FallingPieceData) Next(f *Field) {
	p.CurrentPiece = p.NextPiece

	p.X = f.GetCenterX()
	p.Y = 0

	// todo: configurable
	p.Speed = 1

	p.FallTimer = 1000

	p.Dirty = true

	p.NextPiece = f.rng.NextPiece()

	f.ApplyBedrock()

	f.eventBus.Publish(event.New(fmt.Sprintf("update/%s", f.PlayerID), EventUpdateFallingPiece, &event.Payload{
		"falling_piece_data": f.FallingPiece,
		"piece_display":      f.FallingPiece.CurrentPiece.GetData(),
	}))
}

func (p *FallingPieceData) CanMove(f *Field, dx int, dy int, dr int) bool {
	if p == nil {
		return false
	}

	testPiece := NewPiece(p.CurrentPiece)

	for ; dr > 0; dr-- {
		testPiece.Rotate()
	}

	return f.canPutPiece(testPiece, p.X+dx, p.Y+dy)
}

func (p *FallingPieceData) Move(f *Field, dx int, dy int, dr int) bool {
	if p != nil && p.CanMove(f, dx, dy, dr) {
		p.X += dx
		p.Y += dy

		for ; dr > 0; dr-- {
			p.CurrentPiece.Rotate()
		}

		p.Dirty = true

		return true
	}

	return false
}

func (p *FallingPieceData) HardDrop(f *Field) bool {
	if p == nil {
		return false
	}

	for {
		if m := p.Move(f, 0, 1, 0); !m {
			break
		}
	}

	return p.Lock(f)
}

// Lock returns false on gameover
func (p *FallingPieceData) Lock(f *Field) bool {
	if p != nil {
		f.PutPiece(p.CurrentPiece, p.X, p.Y)
	}

	cleared := f.ClearFullRows()

	if cleared != 0 {
		f.eventBus.Publish(event.New(fmt.Sprintf("update/%s", f.PlayerID), EventRowsCleared, &event.Payload{
			"count": cleared,
		}))
	}

	p.Next(f)

	if nm := p.CanMove(f, 0, 1, 0); !nm {
		f.eventBus.Publish(event.New(fmt.Sprintf("update/%s", f.PlayerID), EventGameOver, &event.Payload{}))

		return false
	}

	return true
}
