package score

type Score struct {
	Score int
	Lines int
	dirty bool
}

func New() *Score {
	return &Score{
		Score: 0,
		Lines: 0,
		dirty: true,
	}
}

func (s *Score) IsDirty() bool {
	return s.dirty
}

func (s *Score) SetDirty(dirty bool) {
	s.dirty = dirty
}

func (s *Score) AddLines(l int) {
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
