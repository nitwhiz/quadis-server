package game

import (
	"bloccs-server/pkg/event"
	"fmt"
)

func (g *Game) publish(e string, p event.Payload) {
	g.EventBus.Publish(event.New(fmt.Sprintf("update/%s", g.ID), e, p))
}

func (g *Game) publishFieldUpdate() {
	g.Field.RLock()
	defer g.Field.RUnlock()

	g.publish(event.UpdateField, &event.UpdateFieldPayload{
		Width:  g.Field.Width,
		Height: g.Field.Height,
		Data:   g.Field.Data,
	})
}

func (g *Game) publishFallingPieceUpdate() {
	g.publish(event.UpdateFallingPiece, &event.UpdateFallingPiecePayload{
		PieceName: g.FallingPiece.Piece.Name,
		Rotation:  g.FallingPiece.Rotation,
		X:         g.FallingPiece.X,
		Y:         g.FallingPiece.Y,
	})
}

func (g *Game) publishNextPieceUpdate() {
	g.publish(event.UpdateNextPiece, &event.UpdateNextPiecePayload{
		PieceName: g.NextPiece.Name,
	})
}

func (g *Game) publishHoldPieceUpdate() {
	g.publish(event.UpdateHoldPiece, &event.UpdateHoldPiecePayload{
		PieceName: g.HoldPiece.Name,
	})
}

func (g *Game) publishScoreUpdate() {
	g.publish(event.UpdateScore, &event.UpdateScorePayload{
		Score: g.Score.Score,
		Lines: g.Score.Lines,
	})
}
