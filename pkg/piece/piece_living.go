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
	l.mu.RLock()
	defer l.mu.RUnlock()

	return &Payload{
		Token: l.piece.Token,
	}
}

func (l *LivingPiece) GetPiece() *Piece {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.piece
}

func (l *LivingPiece) SetPiece(piece *Piece) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.piece != piece {
		l.piece = piece
		l.Dirty.Trip()
	}
}

func (l *LivingPiece) IsLocked() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.locked
}

func (l *LivingPiece) SetLocked(v bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.locked != v {
		l.locked = v
		l.Dirty.Trip()
	}
}
