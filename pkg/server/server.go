package server

import (
	"errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nitwhiz/quadis-server/pkg/metrics"
	"github.com/nitwhiz/quadis-server/pkg/room"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
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
	rooms      map[string]*room.Room
	roomsMutex *sync.Mutex
}

func New() *Server {
	return &Server{
		rooms:      map[string]*room.Room{},
		roomsMutex: &sync.Mutex{},
	}
}

func (s *Server) createRoom() *room.Room {
	r := room.New()

	s.roomsMutex.Lock()
	defer s.roomsMutex.Unlock()

	s.rooms[r.GetId()] = r

	return r
}

func (s *Server) removeRoom(r *room.Room) {
	rId := r.GetId()

	s.roomsMutex.Lock()
	defer s.roomsMutex.Unlock()

	if _, ok := s.rooms[rId]; ok {
		delete(s.rooms, rId)
	}
}

func (s *Server) getRoom(id string) *room.Room {
	s.roomsMutex.Lock()
	defer s.roomsMutex.Unlock()

	if r, ok := s.rooms[id]; ok {
		return r
	}

	return nil
}

func (s *Server) connect(roomId string, resp http.ResponseWriter, req *http.Request) error {
	r := s.getRoom(roomId)

	if r != nil {
		conn, err := upgrader.Upgrade(resp, req, nil)

		if err != nil {
			return err
		}

		return r.CreateGame(conn)
	}

	return errors.New("room_not_found")

}

func (s *Server) WaitForRoomShutdown(r *room.Room) {
	<-r.ShutdownDone()
	s.removeRoom(r)
}

func (s *Server) StartMetricsCollector() {
	go func() {
		for {
			s.roomsMutex.Lock()

			metrics.RoomsTotal.Set(float64(len(s.rooms)))

			gamesTotal := 0
			runningGamesTotal := 0

			for _, r := range s.rooms {
				gamesTotal += r.GetGamesCount()
				runningGamesTotal += r.GetRunningGamesCount()
			}

			s.roomsMutex.Unlock()

			metrics.GamesTotal.Set(float64(gamesTotal))
			metrics.GamesRunningTotal.Set(float64(runningGamesTotal))

			time.Sleep(time.Second * 5)
		}
	}()
}

func (s *Server) Start() error {
	r := gin.Default()

	r.Use(cors.Default())

	r.POST("/rooms", func(c *gin.Context) {
		r := s.createRoom()

		c.JSON(http.StatusOK, gin.H{
			"roomId": r.GetId(),
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

		r := s.getRoom(roomId)

		if r == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "room not found",
			})

			return
		}

		r.Start()

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

		r := s.getRoom(roomId)

		if r == nil {
			c.Status(http.StatusNotFound)
		} else {
			c.Status(http.StatusNoContent)
		}
	})

	r.GET("/rooms/:roomId/socket", func(c *gin.Context) {
		roomId := c.Param("roomId")

		if err := s.connect(roomId, c.Writer, c.Request); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if r := s.getRoom(roomId); r != nil {
			go s.WaitForRoomShutdown(r)
		}
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	if gin.IsDebugging() {
		r.POST("/rooms/:roomId/console", func(c *gin.Context) {
			roomId := c.Param("roomId")

			if r := s.getRoom(roomId); r != nil {
				requestBody, err := io.ReadAll(c.Request.Body)

				if err != nil {
					c.AbortWithStatus(http.StatusBadRequest)
				} else {
					if err := s.handleRoomConsoleCommand(r, requestBody); err != nil {
						c.AbortWithStatus(http.StatusBadRequest)
					}
				}
			}
		})
	}

	s.StartMetricsCollector()

	return r.Run("0.0.0.0:7000")
}
