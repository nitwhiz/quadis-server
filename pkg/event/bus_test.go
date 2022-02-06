package event

import (
	"testing"
	"time"
)

func TestBus_BusGeneric(t *testing.T) {
	handler1Calls := 0
	handler2Calls := 0
	handler3Calls := 0
	handler4Calls := 0

	b := NewBus()

	b.AddChannel("test/123")
	b.AddChannel("test/456")

	b.Subscribe("test/123", func(e *Event) {
		handler1Calls++
	}, "test")

	b.Subscribe("test/456", func(e *Event) {
		handler2Calls++
	}, "test")

	b.Subscribe("test/.*", func(e *Event) {
		handler3Calls++
	}, "test")

	b.Publish(New("test/123", "test", nil))
	b.Publish(New("test/456", "test", nil))

	// give channels some time to be received
	time.Sleep(time.Millisecond * 250)

	b.Subscribe("test/.*", func(e *Event) {
		handler4Calls++
	}, "test")

	b.Publish(New("test/456", "test", nil))

	// give channels some time to be received
	time.Sleep(time.Millisecond * 250)

	b.Stop()

	if handler1Calls != 1 {
		t.Fatalf("handler 1 called %d times, expected 1", handler1Calls)
	}

	if handler2Calls != 2 {
		t.Fatalf("handler 2 called %d times, expected 2", handler2Calls)
	}

	if handler3Calls != 3 {
		t.Fatalf("handler 3 called %d times, expected 3", handler3Calls)
	}

	if handler4Calls != 1 {
		t.Fatalf("handler 4 called %d times, expected 1", handler4Calls)
	}
}

func TestBus_BusBroadcast(t *testing.T) {
	handler1Calls := 0
	handler2Calls := 0

	b := NewBus()

	b.AddChannel("test/123")
	b.AddChannel("test/456")

	b.Subscribe("test/123", func(e *Event) {
		handler1Calls++
	}, "test")

	b.Subscribe("test/456", func(e *Event) {
		handler2Calls++
	}, "test")

	b.Publish(New("*", "test", nil))
	b.Publish(New("*", "test", nil))
	b.Publish(New("*", "test", nil))

	// give channels some time to be received
	time.Sleep(time.Millisecond * 250)

	if handler1Calls != 3 {
		t.Fatalf("handler 1 called %d times, expected 3", handler1Calls)
	}

	if handler2Calls != 3 {
		t.Fatalf("handler 2 called %d times, expected 3", handler2Calls)
	}
}

func TestBus_Unsubscribe(t *testing.T) {
	handler1Calls := 0
	handler2Calls := 0

	b := NewBus()

	b.AddChannel("test/123")
	b.AddChannel("test/456")

	b.Subscribe("test/123", func(e *Event) {
		handler1Calls++
	}, "test1")

	b.Subscribe("test/456", func(e *Event) {
		handler2Calls++
	}, "test2")

	b.Publish(New("test/123", "test", nil))
	b.Publish(New("test/456", "test", nil))

	// give channels some time to be received
	time.Sleep(time.Millisecond * 250)

	b.Unsubscribe("test2")

	b.Publish(New("test/456", "test", nil))

	// give channels some time to be received
	time.Sleep(time.Millisecond * 250)

	b.Stop()

	if handler1Calls != 1 {
		t.Fatalf("handler 1 called %d times, expected 1", handler1Calls)
	}

	if handler2Calls != 1 {
		t.Fatalf("handler 2 called %d times, expected 1", handler2Calls)
	}
}
