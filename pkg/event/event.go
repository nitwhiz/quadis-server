package event

import (
	"encoding/json"
)

type Payload map[string]interface{}

// todo: playload types

type Event struct {
	Channel string   `json:"channel"`
	Type    string   `json:"type"`
	Payload *Payload `json:"payload"`
}

func New(channelName string, eventType string, payload *Payload) *Event {
	return &Event{
		Channel: channelName,
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
