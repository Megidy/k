package api

import (
	"database/sql"

	"github.com/Megidy/k/pkg/services/game"
	"github.com/Megidy/k/pkg/services/user"
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

	//initializing of stores
	gameStore := game.NewGameStore(s.db)
	userStore := user.NewStore(s.db)

	//initializing of handlers

	//clientSideHandler
	clientSideHandler := client.NewClientSideHandler()

	//userHandler
	userHandler := user.NewUserHandler(userStore)
	userHandler.RegisterRoutes(router)

	//gameHandler
	gameHandler := game.NewGameHandler(clientSideHandler, gameStore)
	gameHandler.RegisterRoutes(router)

	return router.Run(s.addr)

}
