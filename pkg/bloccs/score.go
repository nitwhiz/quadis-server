package bloccs

import (
	"bloccs-server/pkg/event"
	"sync"
)

const EventScoreUpdate = "score_update"

type Score struct {
	GameId   string `json:"gameId"`
	Score    int    `json:"score"`
	Lines    int    `json:"lines"`
	mu       *sync.RWMutex
	eventBus *event.Bus
}

func NewScore(gameId string) *Score {
	return &Score{
		GameId: gameId,
		Score:  0,
		Lines:  0,
		mu:     &sync.RWMutex{},
	}
}

func (s *Score) GetId() string {
	return s.GameId
}

func (s *Score) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Score = 0
	s.Lines = 0
}

func (s *Score) AddLines(l int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if l == 0 {
		return
	}

	switch l {
	case 1:
		s.Score += 60
		break
	case 2:
		s.Score += 150
		break
	case 3:
		s.Score += 420
		break
	case 4:
		s.Score += 2500
		break
	default:
		break
	}

	s.Lines += l

	s.eventBus.Publish(event.New(EventScoreUpdate, s, nil))
}
