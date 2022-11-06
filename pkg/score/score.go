package score

import (
	"github.com/nitwhiz/quadis-server/pkg/dirty"
	"sync"
)

type Score struct {
	score int
	lines int
	Dirty *dirty.Dirtiness
	mu    *sync.RWMutex
}

type Payload struct {
	Score int `json:"score"`
	Lines int `json:"lines"`
}

func New() *Score {
	return &Score{
		score: 0,
		lines: 0,
		Dirty: dirty.New(),
		mu:    &sync.RWMutex{},
	}
}

func (s *Score) ToPayload() *Payload {
	defer s.mu.RUnlock()
	s.mu.RLock()

	return &Payload{
		Score: s.score,
		Lines: s.lines,
	}
}

func (s *Score) Reset() {
	defer s.mu.Unlock()
	s.mu.Lock()

	s.score = 0
	s.lines = 0

	s.Dirty.Trip()
}

func (s *Score) AddLines(l int) {
	defer s.mu.Unlock()
	s.mu.Lock()

	if l == 0 {
		return
	}

	switch l {
	case 1:
		s.score += 60
		break
	case 2:
		s.score += 150
		break
	case 3:
		s.score += 420
		break
	case 4:
		s.score += 2500
		break
	default:
		break
	}

	s.lines += l

	s.Dirty.Trip()
}
