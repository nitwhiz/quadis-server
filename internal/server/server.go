package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Server struct {
	rooms            map[string]*Room
	roomsMutex       *sync.Mutex
	systemWaitGroup  *sync.WaitGroup
	systemCancelFunc context.CancelFunc
}

func New() *Server {
	return &Server{
		rooms:            map[string]*Room{},
		roomsMutex:       &sync.Mutex{},
		systemWaitGroup:  &sync.WaitGroup{},
		systemCancelFunc: nil,
	}
}

func (s *Server) CreateRoom() *Room {
	room := NewRoom()

	s.roomsMutex.Lock()
	defer s.roomsMutex.Unlock()

	s.rooms[room.GetId()] = room

	return room
}

func (s *Server) GetRoom(id string) *Room {
	s.roomsMutex.Lock()
	defer s.roomsMutex.Unlock()

	r, ok := s.rooms[id]

	if ok {
		return r
	}

	return nil
}

func (s *Server) connect(roomId string, w http.ResponseWriter, r *http.Request) error {
	room := s.GetRoom(roomId)

	fmt.Println("upgrading connection")

	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return err
	}

	if room == nil {
		if err = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "room_has_games_running"), time.Now().Add(time.Second)); err != nil {
			return err
		}

		if err = conn.Close(); err != nil {
			return err
		}

		return errors.New("room_not_found")
	}

	if room.AreGamesRunning() {
		log.Println("games running")

		if err = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "room_has_games_running"), time.Now().Add(time.Second)); err != nil {
			return err
		}

		if err = conn.Close(); err != nil {
			return err
		}

		return errors.New("room_has_games_running")
	}

	fmt.Println("connection established")

	return room.Join(conn)
}

func (s *Server) startHTTPServer() error {
	r := gin.Default()

	r.Use(cors.Default())

	r.POST("/rooms", func(c *gin.Context) {
		room := s.CreateRoom()

		c.JSON(http.StatusOK, gin.H{
			"roomId": room.Id,
		})
	})

	r.POST("/rooms/:roomId/start", func(c *gin.Context) {
		roomId := c.Param("roomId")

		if roomId == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "room id is empty",
			})

			return
		}

		room := s.GetRoom(roomId)

		if room == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "room not found",
			})

			return
		}

		room.Start()

		c.Status(http.StatusNoContent)
	})

	r.GET("/rooms/:roomId", func(c *gin.Context) {
		roomId := c.Param("roomId")

		if roomId == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "room id is empty",
			})

			return
		}

		room := s.GetRoom(roomId)

		if room == nil {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusNoContent)
		}
	})

	r.GET("/rooms/:roomId/socket", func(c *gin.Context) {
		roomId := c.Param("roomId")

		if err := s.connect(roomId, c.Writer, c.Request); err != nil {
			c.Abort()
		}
	})

	return r.Run("0.0.0.0:7000")
}

func (s *Server) Stop() {
	if s.systemCancelFunc != nil {
		s.systemCancelFunc()
	}

	s.systemWaitGroup.Wait()
}

func (s *Server) Start() error {
	ctx, cancel := context.WithCancel(context.Background())

	s.systemCancelFunc = cancel

	// room watching
	go func() {
		s.systemWaitGroup.Add(1)
		defer s.systemWaitGroup.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Minute):
				s.roomsMutex.Lock()

				runningRooms := map[string]*Room{}
				playerCount := 0

				// todo: rework this

				for rid, r := range s.rooms {
					if r.ShouldClose() {
						r.Stop()
					} else {
						runningRooms[rid] = r
						playerCount += r.GetPlayerCount()
					}
				}

				log.Println(fmt.Sprintf("ROOMS: %d/%d", len(runningRooms), len(s.rooms)))
				log.Println(fmt.Sprintf("PLAYERS: %d", playerCount))

				s.rooms = runningRooms

				s.roomsMutex.Unlock()
			}

		}
	}()

	if err := s.startHTTPServer(); err != nil {
		return err
	}

	return nil
}
