package event

import (
	"encoding/json"
)

type Origin struct {
	Id   string `json:"id"`
	Type string `json:"type"`
}

type Event struct {
	Type    string  `json:"type"`
	Origin  *Origin `json:"origin"`
	Payload any     `json:"payload"`
}

type WindowPayload struct {
	Events []*Event `json:"events"`
}

func (e *Event) Serialize() (string, error) {
	bs, err := json.Marshal(e)

	if err != nil {
		return "", err
	}

	return string(bs), nil
}
