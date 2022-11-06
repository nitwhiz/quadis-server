package piece

import (
	"github.com/nitwhiz/quadis-server/pkg/dirty"
	"sync"
)

type LivingPiece struct {
	piece  *Piece
	Dirty  *dirty.Dirtiness
	locked bool
	mu     *sync.RWMutex
}

func NewLivingPiece(piece *Piece) *LivingPiece {
	return &LivingPiece{
		piece:  piece,
		Dirty:  dirty.New(),
		locked: false,
		mu:     &sync.RWMutex{},
	}
}

func (l *LivingPiece) ToPayload() *Payload {
	defer l.mu.RUnlock()
	l.mu.RLock()

	return &Payload{
		Token: l.piece.Token,
	}
}

func (l *LivingPiece) GetPiece() *Piece {
	defer l.mu.RUnlock()
	l.mu.RLock()

	return l.piece
}

func (l *LivingPiece) SetPiece(piece *Piece) {
	defer l.mu.Unlock()
	l.mu.Lock()

	if l.piece != piece {
		l.piece = piece
		l.Dirty.Trip()
	}
}

func (l *LivingPiece) IsLocked() bool {
	defer l.mu.RUnlock()
	l.mu.RLock()

	return l.locked
}

func (l *LivingPiece) SetLocked(v bool) {
	defer l.mu.Unlock()
	l.mu.Lock()

	if l.locked != v {
		l.locked = v
		l.Dirty.Trip()
	}
}
