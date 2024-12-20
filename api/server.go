package api

import (
	"database/sql"

	"github.com/Megidy/k/pkg/services/game"
	"github.com/Megidy/k/static/client"
	"github.com/gin-gonic/gin"
)

type Server struct {
	addr string
	db   *sql.DB
}

func NewServer(addr string, db *sql.DB) *Server {
	return &Server{
		addr: addr,
		db:   db,
	}
}

func (s *Server) Run() error {
	router := gin.Default()
	clientSideHandler := client.NewClientSideHandler()
	gameStore := game.NewGameStore(s.db)
	handler := game.NewGameHandler(clientSideHandler, gameStore)

	handler.RegisterRoutes(router)

	return router.Run(s.addr)

}
