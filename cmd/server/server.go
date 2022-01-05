package main

import (
	"bloccs-server/pkg/bloccs"
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
)

var wsMutex = &sync.Mutex{}

type Message struct {
	Event *bloccs.Event `json:"event"`
}

func sendUpdate(field *bloccs.Field, conn *websocket.Conn) error {
	bs, err := json.Marshal(Message{
		Event: &bloccs.Event{
			Type: bloccs.EventUpdate,
			Data: map[string]interface{}{
				"field": field,
			},
		},
	})

	if err != nil {
		log.Println("json failed:", err)
	}

	wsMutex.Lock()
	err = conn.WriteMessage(websocket.TextMessage, bs)
	wsMutex.Unlock()

	return err
}

func main() {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// todo: game rooms

	game := bloccs.NewGame()

	game.Field.SetBedrock(3)

	http.HandleFunc("/game", func(w http.ResponseWriter, r *http.Request) {
		// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
		conn, err := upgrader.Upgrade(w, r, nil)

		if err != nil {
			log.Print("upgrade failed: ", err)
			return
		}

		defer func(conn *websocket.Conn) {
			_ = conn.Close()
		}(conn)

		_ = game.EventBus.AddHandler(func(event *bloccs.Event) {
			bs, err := json.Marshal(Message{
				Event: event,
			})

			if err != nil {
				log.Println("json failed:", err)
			}

			wsMutex.Lock()
			err = conn.WriteMessage(websocket.TextMessage, bs)
			wsMutex.Unlock()

			if err != nil {
				log.Println("write failed:", err)
			}
		})

		game.AddUpdateHandler(func() {
			err := sendUpdate(game.Field, conn)

			if err != nil {
				log.Println("write failed:", err)
				return
			}
		})

		game.Start()

		for {
			_, message, err := conn.ReadMessage()

			if err != nil {
				log.Println("read failed:", err)
				break
			}

			if string(message) == "L" {
				game.Field.MoveFallingPiece(-1, 0, 0)

				_ = sendUpdate(game.Field, conn)
			} else if string(message) == "R" {
				game.Field.MoveFallingPiece(1, 0, 0)

				_ = sendUpdate(game.Field, conn)
			} else if string(message) == "D" {
				game.Field.MoveFallingPiece(0, 1, 0)

				_ = sendUpdate(game.Field, conn)
			} else if string(message) == "X" {
				game.Field.MoveFallingPiece(0, 0, 1)

				_ = sendUpdate(game.Field, conn)
			}
		}
	})

	_ = http.ListenAndServe(":7000", nil)
}
