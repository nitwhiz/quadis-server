package server

import (
	"errors"
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

type BloccsServer struct {
	rooms             map[string]*Room
	roomsMutex        *sync.Mutex
	systemWaitGroup   *sync.WaitGroup
	systemStopChannel chan bool
}

func NewBloccsServer() *BloccsServer {
	return &BloccsServer{
		rooms:             map[string]*Room{},
		roomsMutex:        &sync.Mutex{},
		systemWaitGroup:   &sync.WaitGroup{},
		systemStopChannel: make(chan bool),
	}
}

func (s *BloccsServer) CreateRoom() *Room {
	room := NewRoom()

	s.roomsMutex.Lock()
	defer s.roomsMutex.Unlock()

	s.rooms[room.ID] = room

	return room
}

func (s *BloccsServer) GetRoom(id string) *Room {
	s.roomsMutex.Lock()
	defer s.roomsMutex.Unlock()

	r, ok := s.rooms[id]

	if ok {
		return r
	}

	return nil
}

func (s *BloccsServer) connect(roomId string, w http.ResponseWriter, r *http.Request) error {
	room := s.GetRoom(roomId)

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

	return room.Join(conn)
}

func (s *BloccsServer) startHTTPServer() error {
	r := gin.Default()

	r.Use(cors.Default())

	r.POST("/rooms", func(c *gin.Context) {
		room := s.CreateRoom()

		c.JSON(http.StatusOK, gin.H{
			"roomId": room.ID,
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

func (s *BloccsServer) Stop() {
	close(s.systemStopChannel)

	s.systemWaitGroup.Wait()
}

func (s *BloccsServer) Start() error {
	go func() {
		s.systemWaitGroup.Add(1)
		defer s.systemWaitGroup.Done()

		for {
			select {
			case <-s.systemStopChannel:
				return
			case <-time.After(time.Second):
				s.roomsMutex.Lock()

				for _, r := range s.rooms {
					if r.ShouldClose() {
						r.Stop()
					}
				}

				s.roomsMutex.Unlock()
			}

		}
	}()

	if err := s.startHTTPServer(); err != nil {
		return err
	}

	return nil
}
