package game

import (
	"bloccs-server/pkg/piece"
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
}

func NewFallingPiece() *FallingPiece {
	return &FallingPiece{
		Piece:     nil,
		X:         0,
		Y:         0,
		Speed:     1,
		FallTimer: 0,
		dirty:     false,
	}
}

func (p *FallingPiece) IsDirty() bool {
	return p.dirty
}

func (p *FallingPiece) SetDirty(dirty bool) {
	p.dirty = dirty
}

func (p *FallingPiece) Rotate() {
	p.Rotation = p.Piece.ClampRotation(p.Rotation + 1)
	p.dirty = true
}

func (p *FallingPiece) GetData() *[]uint8 {
	return p.Piece.GetData(p.Rotation)
}

func (p *FallingPiece) GetDataXY(x int, y int) uint8 {
	return p.Piece.GetDataXY(p.Rotation, x, y)
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
	g.FallingPiece.X = g.Field.GetCenterX()
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

	g.Update()
}
