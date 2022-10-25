package event

import (
	"context"
	"sync"
	"time"
)

type WindowClosedCallback func([]*Event)

type Window struct {
	events    []*Event
	callback  WindowClosedCallback
	mu        *sync.RWMutex
	isWaiting bool
	ctx       context.Context
}

func NewWindow(parentContext context.Context, cb WindowClosedCallback) *Window {
	return &Window{
		events:    []*Event{},
		callback:  cb,
		mu:        &sync.RWMutex{},
		isWaiting: false,
		ctx:       parentContext,
	}
}

func (w *Window) Add(e *Event) {
	defer w.mu.Unlock()
	w.mu.Lock()

	w.events = append(w.events, e)

	if !w.isWaiting {
		w.isWaiting = true

		go w.waitCallback()
	}
}

func (w *Window) runCallback() {
	defer w.mu.Unlock()
	w.mu.Lock()

	var events []*Event

	for _, e := range w.events {
		events = append(events, e)
	}

	w.events = []*Event{}

	w.callback(events)
}

func (w *Window) waitCallback() {
	defer w.runCallback()
	defer func() {
		w.isWaiting = false
	}()

	w.isWaiting = true

	for {
		select {
		case <-w.ctx.Done():
		case <-time.After(time.Millisecond * 10):
			return
		}
	}
}
