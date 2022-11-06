package falling_piece

import (
	"github.com/nitwhiz/quadis-server/pkg/dirty"
	"github.com/nitwhiz/quadis-server/pkg/piece"
	"sync"
)

type FallingPiece struct {
	piece     *piece.Piece
	x         int
	y         int
	rotation  piece.Rotation
	speed     int64
	fallTimer int64
	locked    bool
	Dirty     *dirty.Dirtiness
	mu        *sync.RWMutex
}

type Payload struct {
	Piece    piece.Payload  `json:"piece"`
	Rotation piece.Rotation `json:"rotation"`
	X        int            `json:"x"`
	Y        int            `json:"y"`
}

func New(piece *piece.Piece) *FallingPiece {
	return &FallingPiece{
		piece:     piece,
		x:         0,
		y:         0,
		rotation:  0,
		speed:     1,
		fallTimer: 1000,
		locked:    false,
		Dirty:     dirty.New(),
		mu:        &sync.RWMutex{},
	}
}

func (p *FallingPiece) ToPayload() *Payload {
	defer p.mu.RUnlock()
	p.mu.RLock()

	return &Payload{
		Piece: piece.Payload{
			Token: p.piece.Token,
		},
		Rotation: p.rotation,
		X:        p.x,
		Y:        p.y,
	}
}

func (p *FallingPiece) Lock() {
	defer p.mu.Unlock()
	p.mu.Lock()

	p.fallTimer = 0
	p.locked = true
}

func (p *FallingPiece) IsLocked() bool {
	defer p.mu.RUnlock()
	p.mu.RLock()

	return p.locked
}

func (p *FallingPiece) GetPieceAndPosition() (*piece.Piece, piece.Rotation, int, int) {
	defer p.mu.RUnlock()
	p.mu.RLock()

	return p.piece, p.rotation, p.x, p.y
}

func (p *FallingPiece) GetNextPosition(delta int64) (bool, int) {
	defer p.mu.RUnlock()
	p.mu.RLock()

	if p.piece != nil && !p.locked {
		p.fallTimer -= delta

		if p.fallTimer <= 0 {
			p.fallTimer = 1000 / p.speed

			return true, p.y + 1
		}
	}

	return false, p.y
}

func (p *FallingPiece) GetPiece() *piece.Piece {
	defer p.mu.RUnlock()
	p.mu.RLock()

	return p.piece
}

func (p *FallingPiece) Update(nextPiece *piece.Piece, x int, y int, r piece.Rotation) {
	defer p.mu.Unlock()
	p.mu.Lock()

	p.piece = nextPiece

	p.x = x
	p.y = y
	p.rotation = r

	p.speed = 1
	p.fallTimer = 1000 / p.speed

	p.locked = false

	p.Dirty.Trip()
}

func (p *FallingPiece) SetY(y int) {
	defer p.mu.Unlock()
	p.mu.Lock()

	p.y = y

	p.Dirty.Trip()
}

func (p *FallingPiece) SetPosition(r piece.Rotation, x int, y int) {
	defer p.mu.Unlock()
	p.mu.Lock()

	p.rotation = r
	p.x = x
	p.y = y

	p.Dirty.Trip()
}
