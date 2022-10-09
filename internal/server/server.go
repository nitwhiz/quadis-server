package server

import (
	"errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nitwhiz/quadis-server/pkg/room"
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

func (s *Server) getRoom(id string) *room.Room {
	defer s.roomsMutex.Unlock()
	s.roomsMutex.Lock()

	if r, ok := s.rooms[id]; ok {
		return r
	}

	return nil
}

func (s *Server) connect(roomId string, resp http.ResponseWriter, req *http.Request) error {
	r := s.getRoom(roomId)

	conn, err := upgrader.Upgrade(resp, req, nil)

	if r == nil {
		if err = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "room_not_found"), time.Now().Add(time.Second)); err != nil {
			return err
		}

		if err = conn.Close(); err != nil {
			return err
		}

		return errors.New("room_not_found")
	}

	// todo: if room is already active

	return r.CreatePlayer(conn)
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
			c.Abort()
		}
	})

	return r.Run("0.0.0.0:7000")
}
