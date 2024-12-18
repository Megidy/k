package game

import (
	"github.com/Megidy/k/static/game"
	"github.com/gin-gonic/gin"
)

type GameHandler struct {
	manager *Manager
}

func NewGameHandler(manager *Manager) *GameHandler {
	h := &GameHandler{manager: manager}
	go h.manager.MessageQueue()
	go h.manager.CheckReadiness()
	go h.manager.SetClientsInReadiness()
	return h
}

func (h *GameHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/", func(ctx *gin.Context) {
		game.Game().Render(ctx.Request.Context(), ctx.Writer)
	})
	router.GET("/ws/game", func(ctx *gin.Context) {
		h.manager.NewConnection(ctx)
	})
}
