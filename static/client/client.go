package client

import "github.com/gin-gonic/gin"

type ClientSideHandler struct {
}

func NewClientSideHandler() *ClientSideHandler {
	return &ClientSideHandler{}
}
func (h *ClientSideHandler) GetDataFromForm(c *gin.Context, key string) string {
	return c.Request.PostFormValue(key)
}
