package player

import "sync"

type Player struct {
	name string
	mu   *sync.RWMutex
}

func New(name string) *Player {
	return &Player{
		name: name,
		mu:   &sync.RWMutex{},
	}
}

func (p *Player) GetName() string {
	defer p.mu.RUnlock()
	p.mu.RLock()

	return p.name
}
