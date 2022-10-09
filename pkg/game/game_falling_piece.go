package game

import (
	"github.com/nitwhiz/quadis-server/pkg/falling_piece"
	"github.com/nitwhiz/quadis-server/pkg/field"
	"github.com/nitwhiz/quadis-server/pkg/piece"
)

func (g *Game) putFallingPiece() int {
	g.fallingPiece.Lock()

	p, pRot, pX, pY := g.fallingPiece.GetPieceAndPosition()

	g.field.PutPiece(p, pRot, pX, pY)

	return g.field.ClearLines()
}

func (g *Game) hardLockFallingPiece() {
	if g.fallingPiece == nil {
		return
	}

	dy := 0

	p, pRot, pX, pY := g.fallingPiece.GetPieceAndPosition()

	for dy < field.Height {
		if !g.field.CanPutPiece(p, pRot, pX, pY+dy) {
			break
		}

		dy++
	}

	g.fallingPiece.SetY(pY + dy - 1)
	g.fallingPiece.Lock()
}

func (g *Game) nextFallingPiece(lastPieceWasHeld bool) {
	if g.nextPiece == nil {
		g.nextPiece = piece.NewLivingPiece(g.rpg.NextElement())
	}

	if g.fallingPiece == nil {
		g.fallingPiece = falling_piece.New(nil)
	}

	g.fallingPiece.Update(g.nextPiece.GetPiece(), g.field.GetCenterX(), 0, 0)

	g.nextPiece.SetPiece(g.rpg.NextElement())

	if !lastPieceWasHeld {
		g.holdingPiece.SetLocked(false)
	}
}

func (g *Game) tryTranslateFallingPiece(dr piece.Rotation, dx int, dy int) {
	if g.fallingPiece == nil || g.fallingPiece.IsLocked() {
		return
	}

	p, pr, px, py := g.fallingPiece.GetPieceAndPosition()

	if g.field.CanPutPiece(p, pr+dr, px+dx, py+dy) {
		g.fallingPiece.SetPosition(pr+dr, px+dx, py+dy)
	}
}

func (g *Game) clearLinesAndNextPiece() (int, bool) {
	clearedLines := g.putFallingPiece()
	gameOver := false

	g.nextFallingPiece(false)

	p, fpRot, fpX, fpY := g.fallingPiece.GetPieceAndPosition()

	if !g.field.CanPutPiece(p, fpRot, fpX, fpY) {
		gameOver = true
	}

	return clearedLines, gameOver
}

func (g *Game) updateFallingPiece(delta int64) (int, bool) {
	clearedLines := 0
	gameOver := false

	if g.fallingPiece == nil {
		g.nextFallingPiece(false)
	}

	if g.fallingPiece.IsLocked() {
		clearedLines, gameOver = g.clearLinesAndNextPiece()
	} else {
		p, pRot, pX, _ := g.fallingPiece.GetPieceAndPosition()
		shouldMove, nextY := g.fallingPiece.GetNextPosition(delta)

		if shouldMove {
			if g.field.CanPutPiece(p, pRot, pX, nextY) {
				g.fallingPiece.SetY(nextY)
			} else {
				clearedLines, gameOver = g.clearLinesAndNextPiece()
			}
		}
	}

	return clearedLines, gameOver
}

func (g *Game) tryHoldFallingPiece() {
	if g.fallingPiece == nil || g.holdingPiece.IsLocked() {
		return
	}

	currentHoldingPiece := g.holdingPiece.GetPiece()

	g.holdingPiece.SetPiece(g.fallingPiece.GetPiece())

	if currentHoldingPiece == nil {
		g.nextFallingPiece(true)
	} else {
		g.fallingPiece.Update(currentHoldingPiece, g.field.GetCenterX(), 0, 0)
	}

	g.holdingPiece.SetLocked(true)
}
