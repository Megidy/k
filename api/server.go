package api

import (
	"github.com/Megidy/k/pkg/services/game"
	"github.com/gin-gonic/gin"
)

type Server struct {
	addr string
}

func NewServer(addr string) *Server {
	return &Server{
		addr: addr,
	}
}

func (s *Server) Run() error {
	router := gin.Default()
	manager := game.NewManager()
	handler := game.NewGameHandler(manager)
	handler.RegisterRoutes(router)

	return router.Run(s.addr)

}
