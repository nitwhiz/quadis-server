package main

import "bloccs-server/internal/server"

func main() {
	srv := server.NewBloccsServer()

	err := srv.Start()

	if err != nil {
		panic(err)
	}
}
