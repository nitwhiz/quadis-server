package score

import (
	"github.com/nitwhiz/quadis-server/pkg/dirty"
	"github.com/nitwhiz/quadis-server/pkg/metrics"
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
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &Payload{
		Score: s.score,
		Lines: s.lines,
	}
}

func (s *Score) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.score = 0
	s.lines = 0

	s.Dirty.Trip()
}

func getScoreByLineCount(l int) int {
	switch l {
	case 1:
		return 60
	case 2:
		return 150
	case 3:
		return 420
	case 4:
		return 2500
	default:
		return 0
	}
}

func (s *Score) AddLines(l int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if l == 0 {
		return
	}

	n := getScoreByLineCount(l)

	s.score += n
	s.lines += l

	s.Dirty.Trip()

	metrics.ScoreTotal.Add(float64(n))
	metrics.LinesClearedTotal.Add(float64(l))
}
