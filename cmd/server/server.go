package main

import "github.com/nitwhiz/bloccs-server/internal/server"

func main() {
	srv := server.NewBloccsServer()

	err := srv.Start()

	srv.Stop()

	if err != nil {
		panic(err)
	}
}
