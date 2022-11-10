package item

import (
	"log"
)

const TypeTornado = "item_tornado"

// todo: the necessity of these interfaces must mean that this architecture could be better (?)

type Game interface {
}

type Room interface {
}

type ActivateFunc func(sourceGame Game, room Room)

type Item struct {
	Name     string
	Activate ActivateFunc
}

func NewTornado() *Item {
	return &Item{
		Name: TypeTornado,
		Activate: func(sourceGame Game, room Room) {
			if room == nil {
				return
			}

			log.Println("hello from tornado item!")
		},
	}
}
