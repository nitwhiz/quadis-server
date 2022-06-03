package event

import (
	"testing"
)

type MyEventSource struct {
	Id    string
	Extra int
}

func (s *MyEventSource) GetId() string {
	return s.Id
}

func TestEvent_New(t *testing.T) {
	src := MyEventSource{Id: "1337", Extra: 42}

	evnt := New("testType", &src, nil)

	if evnt.Source != &src {
		t.Fatalf("source is not the same")
	}

	if (evnt.Source).(*MyEventSource).Extra != 42 {
		t.Fatalf("incorrect event source data")
	}
}
