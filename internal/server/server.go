package server

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type BloccsServer struct {
	rooms map[string]*Room
}

func NewBloccsServer() *BloccsServer {
	return &BloccsServer{
		rooms: map[string]*Room{},
	}
}

func (s *BloccsServer) CreateRoom() *Room {
	room := NewRoom()

	s.rooms[room.ID] = room

	return room
}

func (s *BloccsServer) GetRoom(id string) *Room {
	r, ok := s.rooms[id]

	if ok {
		return r
	}

	return nil
}

func (s *BloccsServer) connect(playerName string, room *Room, w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return err
	}

	player := NewPlayer(playerName, conn)

	room.AddPlayer(player)

	return nil
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

	r.GET("/rooms/:roomId/socket", func(c *gin.Context) {
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

		playerName := "test"

		if err := s.connect(playerName, room, c.Writer, c.Request); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "unable to connect: " + err.Error(),
			})
		}
	})

	return r.Run("0.0.0.0:7000")
}

func (s *BloccsServer) Start() error {
	// todo: start room cleaner

	if err := s.startHTTPServer(); err != nil {
		return err
	}

	return nil
}
