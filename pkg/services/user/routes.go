package user

import (
	"github.com/Megidy/k/types"
	"github.com/gin-gonic/gin"
)

type userHandler struct {
	userStore types.UserStore
}

func NewUserHandler(userStore types.UserStore) *userHandler {
	return &userHandler{
		userStore: userStore,
	}
}
func (h *userHandler) RegisterRoutes(router gin.IRouter) {

}
