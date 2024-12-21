package user

import (
	"log"
	"net/http"

	"github.com/Megidy/k/pkg/auth"
	"github.com/Megidy/k/static/user"
	"github.com/Megidy/k/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type userHandler struct {
	clientSideHandler types.ClientSideHandler
	userStore         types.UserStore
}

func NewUserHandler(userStore types.UserStore, clientSideHandler types.ClientSideHandler) *userHandler {
	return &userHandler{
		clientSideHandler: clientSideHandler,
		userStore:         userStore,
	}
}
func (h *userHandler) RegisterRoutes(router gin.IRouter) {
	//create account
	router.GET("/account/create", h.LoadCreateAccountTempl)
	router.POST("/account/create/confirm", h.ConfirmCreateAccount)

	//login
	router.GET("/account/login", h.LoadLoginAccountTempl)
	router.POST("/account/login/confirm", h.ConfirmLoginAccount)
}
func (h *userHandler) LoadCreateAccountTempl(c *gin.Context) {
	user.Signup("").Render(c.Request.Context(), c.Writer)

}
func (h *userHandler) ConfirmCreateAccount(c *gin.Context) {
	var u types.User
	u.UserName = h.clientSideHandler.GetDataFromForm(c, "username")
	u.Email = h.clientSideHandler.GetDataFromForm(c, "email")

	exists, err := h.userStore.UserExists(&u)
	if err != nil {
		log.Println("error when cheking user")
		user.Signup(err.Error()).Render(c.Request.Context(), c.Writer)
		return
	}
	if exists {
		log.Println("user already exists")
		user.Signup("user exists , sorry ;(((").Render(c.Request.Context(), c.Writer)
		return
	}
	u.ID = uuid.NewString()
	password := h.clientSideHandler.GetDataFromForm(c, "password")
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		log.Println("error when hashing password")
		user.Signup(err.Error()).Render(c.Request.Context(), c.Writer)
		return
	}
	u.Password = hashedPassword
	err = h.userStore.CreateUser(&u)
	if err != nil {
		log.Println("error when creating user")
		user.Signup(err.Error()).Render(c.Request.Context(), c.Writer)
		return
	}
	token, err := auth.CreateJWT([]byte("blablabla-sosecretthaticantevenhideitpleasehelpme"), u.ID)
	if err != nil {
		log.Println("error when creating jwt token")
		user.Signup(err.Error()).Render(c.Request.Context(), c.Writer)
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", token, 3600*24*10, "/", "", false, true)
	c.Writer.Header().Add("HX-Redirect", "/room/connect")

}
func (h *userHandler) LoadLoginAccountTempl(c *gin.Context) {
	user.Login("").Render(c.Request.Context(), c.Writer)
}
func (h *userHandler) ConfirmLoginAccount(c *gin.Context) {
	var u types.User
	u.Email = h.clientSideHandler.GetDataFromForm(c, "email")
	nativePassword := h.clientSideHandler.GetDataFromForm(c, "password")
	exists, err := h.userStore.UserExists(&u)
	if err != nil {
		user.Login(err.Error()).Render(c.Request.Context(), c.Writer)
		log.Println("error when checking existion of the user : ", err)
		return
	}
	if !exists {
		user.Login("data is invalid").Render(c.Request.Context(), c.Writer)
		log.Println("user with this written data doesn't exsit")
		return
	}
	usr, err := h.userStore.GetUserByEmail(u.Email)
	if err != nil {
		user.Login(err.Error()).Render(c.Request.Context(), c.Writer)
		log.Println("error when getting user by email : ", err)
		return
	}
	err = auth.CheckPasswordCorrectness(usr.Password, nativePassword)
	if err != nil {
		user.Login("data is invalid").Render(c.Request.Context(), c.Writer)
		log.Println("error when comparing native and hashed passwords : ", err)
		return
	}
	token, err := auth.CreateJWT([]byte("blablabla-sosecretthaticantevenhideitpleasehelpme"), usr.ID)
	if err != nil {
		user.Login(err.Error()).Render(c.Request.Context(), c.Writer)
		log.Println("error when creating jwt token : ", err)
		return
	}
	log.Println("token : ", token)
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", token, 3600*24*10, "/", "", false, true)
	c.Writer.Header().Add("HX-Redirect", "/room/connect")

}
