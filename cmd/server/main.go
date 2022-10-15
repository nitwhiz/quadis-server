package main

import (
	"github.com/nitwhiz/quadis-server/internal/server"
	"log"
)

func main() {
	log.Println("Quadis Server")

	s := server.New()

	err := s.Start()

	if err != nil {
		panic(err)
	}
}
