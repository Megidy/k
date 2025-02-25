package api

import (
	"database/sql"

	"github.com/Megidy/k/config"
	"github.com/Megidy/k/pkg/auth"
	"github.com/Megidy/k/pkg/services/game"
	"github.com/Megidy/k/pkg/services/user"
	"github.com/Megidy/k/static/client"
	"github.com/Megidy/k/worker"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	config  *config.Config
	addr    string
	sqlDB   *sql.DB
	redisDB *redis.Client
}

func NewServer(cfg *config.Config, addr string, sqlDB *sql.DB, redisDB *redis.Client) *Server {
	return &Server{
		config:  cfg,
		addr:    addr,
		sqlDB:   sqlDB,
		redisDB: redisDB,
	}
}

func (s *Server) Run() error {

	router := gin.Default()
	router.Static("/static", "./static")

	//initializing of stores
	gameStore := game.NewGameStore(s.sqlDB, s.redisDB)
	userStore := user.NewStore(s.sqlDB, s.redisDB)

	workerPool := worker.NewWorkerPool(userStore, 100)
	workerPool.Run()
	//initializing of handlers

	//initialization of auth Service
	authHandler := auth.NewHandler(userStore)
	//clientSideHandler
	clientSideHandler := client.NewClientSideHandler()

	//userHandler
	userHandler := user.NewUserHandler(userStore, clientSideHandler)
	userHandler.RegisterRoutes(router, authHandler)

	//gameHandler
	gameHandler := game.NewGameHandler(s.config, clientSideHandler, gameStore, workerPool)
	gameHandler.RegisterRoutes(router, authHandler)

	return router.Run(s.addr)

}
