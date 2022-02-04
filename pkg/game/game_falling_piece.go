package game

import (
	"bloccs-server/pkg/piece"
	"sync"
	"time"
)

type FallingPiece struct {
	Piece     *piece.Piece
	X         int
	Y         int
	Rotation  int
	Speed     int
	FallTimer int
	dirty     bool
	mu        *sync.RWMutex
}

func NewFallingPiece() *FallingPiece {
	return &FallingPiece{
		Piece:     nil,
		X:         0,
		Y:         0,
		Speed:     1,
		FallTimer: 0,
		dirty:     false,
		mu:        &sync.RWMutex{},
	}
}

func (p *FallingPiece) IsDirty() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.dirty
}

func (p *FallingPiece) SetDirty(dirty bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.dirty = dirty
}

func (g *Game) canMoveFallingPiece(dr int, dx int, dy int) bool {
	if g.FallingPiece.Piece != nil && g.Field.CanPutPiece(
		g.FallingPiece.Piece,
		g.FallingPiece.Rotation+dr,
		g.FallingPiece.X+dx,
		g.FallingPiece.Y+dy,
	) {
		return true
	}

	return false
}

func (g *Game) moveFallingPiece(dr int, dx int, dy int) {
	if g.canMoveFallingPiece(dr, dx, dy) {
		g.FallingPiece.Rotation += dr
		g.FallingPiece.X += dx
		g.FallingPiece.Y += dy

		g.FallingPiece.dirty = true
	}
}

func (g *Game) initFallingPiece() {
	g.FallingPiece.X = g.Field.CenterX
	g.FallingPiece.Y = 0
	g.FallingPiece.Rotation = 0

	g.FallingPiece.Speed = 1
	g.FallingPiece.FallTimer = 1000

	g.FallingPiece.dirty = true
}

func (g *Game) setFallingPiece(p *piece.Piece) {
	g.FallingPiece.Piece = p

	g.initFallingPiece()
}

func (g *Game) nextFallingPiece() {
	g.setFallingPiece(g.NextPiece)

	g.NextPiece = g.rpg.NextPiece()

	g.FallingPiece.dirty = true
	g.nextDirty = true
}

func (g *Game) lockFallingPiece() int {
	if g.FallingPiece.Piece != nil {
		g.Field.PutPiece(
			g.FallingPiece.Piece,
			g.FallingPiece.Rotation,
			g.FallingPiece.X,
			g.FallingPiece.Y,
		)
	}

	cleared := g.Field.ClearFullRows()

	g.nextFallingPiece()

	g.holdLock = false

	return cleared
}

func (g *Game) updateFallingPiece(delta int) (int, bool) {
	clearedLines := 0
	gameOver := false

	if g.FallingPiece.Piece == nil {
		g.nextFallingPiece()
		g.holdLock = false
	} else {
		g.FallingPiece.FallTimer -= delta / int(time.Millisecond)

		if g.FallingPiece.FallTimer <= 0 {
			g.FallingPiece.FallTimer = 1000 / g.FallingPiece.Speed

			if g.canMoveFallingPiece(0, 0, 1) {
				g.moveFallingPiece(0, 0, 1)
			} else {
				clearedLines = g.lockFallingPiece()

				if !g.canMoveFallingPiece(0, 0, 1) &&
					!g.canMoveFallingPiece(0, 1, 0) &&
					!g.canMoveFallingPiece(0, -1, 0) {
					gameOver = true
				}
			}
		}
	}

	return clearedLines, gameOver
}

func (g *Game) holdFallingPiece() {
	if g.FallingPiece.Piece == nil {
		return
	}

	if g.holdLock {
		return
	}

	g.holdLock = true

	if g.HoldPiece != nil {
		g.FallingPiece.Piece, g.HoldPiece = g.HoldPiece, g.FallingPiece.Piece

		g.initFallingPiece()
	} else {
		g.HoldPiece = g.FallingPiece.Piece

		g.nextFallingPiece()
	}

	g.holdDirty = true
}

func (g *Game) hardLockFallingPiece() {
	if g.FallingPiece.Piece == nil {
		return
	}

	dy := 0

	for dy < g.Field.Height {
		if !g.canMoveFallingPiece(0, 0, dy) {
			break
		}

		dy++
	}

	g.FallingPiece.Y += dy - 1
	g.FallingPiece.FallTimer = 0
}
