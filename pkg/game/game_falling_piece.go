package game

import (
	"github.com/nitwhiz/quadis-server/pkg/falling_piece"
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

	g.fallingPiece.LockMovement()
	defer g.fallingPiece.UnlockMovement()

	dy := 0

	p, pRot, pX, pY := g.fallingPiece.GetPieceAndPosition()

	for dy < g.GetField().GetHeight() {
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
		g.nextPiece = piece.NewLivingPiece(g.pieceGenerator.NextElement())
	}

	if g.fallingPiece == nil {
		g.fallingPiece = falling_piece.New(nil)
	}

	g.fallingPiece.SetPiece(g.nextPiece.GetPiece(), g.field.GetCenterX(), 0, 0)

	if g.overridePiece == nil {
		g.nextPiece.SetPiece(g.pieceGenerator.NextElement())
	} else {
		g.nextPiece.SetPiece(g.overridePiece)
	}

	if !lastPieceWasHeld {
		g.holdingPiece.SetLocked(false)
	}
}

func (g *Game) tryTranslateFallingPiece(dr piece.Rotation, dx int, dy int) {
	g.fallingPiece.LockMovement()
	defer g.fallingPiece.UnlockMovement()

	if g.fallingPiece == nil || g.fallingPiece.IsLocked() {
		return
	}

	p, pr, px, py := g.fallingPiece.GetPieceAndPosition()

	if dr != 0 && g.fallingPiece.IsRotationLocked() {
		dr = 0
	}

	tr := p.ClampRotation(pr + dr)

	if g.field.CanPutPiece(p, tr, px+dx, py+dy) {
		g.fallingPiece.SetPosition(tr, px+dx, py+dy)
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
	if g.fallingPiece == nil {
		g.nextFallingPiece(false)
	}

	if !g.fallingPiece.TryLockMovement() {
		return 0, false
	}

	defer g.fallingPiece.UnlockMovement()

	clearedLines := 0
	gameOver := false

	p, pRot, pX, pY := g.fallingPiece.GetPieceAndPosition()

	forcedUp := false

	for !g.field.CanPutPiece(p, pRot, pX, pY) {
		pY--
		forcedUp = true

		if pY <= 0 {
			pY = 0

			break
		}
	}

	if forcedUp {
		g.fallingPiece.SetY(pY)
		g.fallingPiece.Lock()
	}

	if g.fallingPiece.IsLocked() {
		clearedLines, gameOver = g.clearLinesAndNextPiece()
	} else {
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
	g.fallingPiece.LockMovement()
	defer g.fallingPiece.UnlockMovement()

	if g.fallingPiece == nil || g.holdingPiece.IsLocked() {
		return
	}

	currentHoldingPiece := g.holdingPiece.GetPiece()

	g.holdingPiece.SetPiece(g.fallingPiece.GetPiece())

	if currentHoldingPiece == nil {
		g.nextFallingPiece(true)
	} else {
		g.fallingPiece.SetPiece(currentHoldingPiece, g.field.GetCenterX(), 0, 0)
	}

	g.holdingPiece.SetLocked(true)
}
