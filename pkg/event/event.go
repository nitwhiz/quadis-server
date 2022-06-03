package event

import (
	"encoding/json"
)

type Source interface {
	GetId() string
}

type Type string

type Payload interface{}

type Event struct {
	Source  Source  `json:"source"`
	Type    Type    `json:"type"`
	Payload Payload `json:"payload"`
}

func New(eventType Type, source Source, payload Payload) *Event {
	return &Event{
		Source:  source,
		Type:    eventType,
		Payload: payload,
	}
}

func (e *Event) GetAsBytes() ([]byte, error) {
	bs, err := json.Marshal(e)

	if err != nil {
		return nil, err
	}

	return bs, nil
}
