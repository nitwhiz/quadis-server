package score

import "sync"

type Score struct {
	Score int
	Lines int
	dirty bool
	mu    *sync.RWMutex
}

func New() *Score {
	return &Score{
		Score: 0,
		Lines: 0,
		dirty: true,
		mu:    &sync.RWMutex{},
	}
}

func (s *Score) RLock() {
	s.mu.RLock()
}

func (s *Score) RUnlock() {
	s.mu.RUnlock()
}

func (s *Score) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Score = 0
	s.Lines = 0
	s.dirty = true
}

func (s *Score) IsDirty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.dirty
}

func (s *Score) SetDirty(dirty bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.dirty = dirty
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

	s.dirty = true
}
