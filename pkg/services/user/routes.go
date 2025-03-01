package user

import (
	"context"
	"log"
	"net/http"

	"github.com/Megidy/k/pkg/auth"
	"github.com/Megidy/k/static/templates/overall"
	"github.com/Megidy/k/static/templates/user"
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
func (h *userHandler) RegisterRoutes(router gin.IRouter, authHandler *auth.Handler) {
	//create account
	router.GET("/account/create", h.LoadCreateAccountTempl)
	router.POST("/account/create/confirm", h.ConfirmCreateAccount)

	//login
	router.GET("/account/login", h.LoadLoginAccountTempl)
	router.POST("/account/login/confirm", h.ConfirmLoginAccount)

	router.GET("/account/info", authHandler.WithJWT, h.LoadUserAccount)
	router.GET("/account/info/:userID", authHandler.WithJWT, h.LoadUserAccount)

	router.GET("/account/info/:userID/leaderboard-history", authHandler.WithJWT, h.LoadUserLeaderBoardHistory)
	router.GET("/account/info/leaderboard-history", authHandler.WithJWT, h.LoadUserLeaderBoardHistory)

	router.POST("/redirect-to-leaderboard-history", authHandler.WithJWT, h.RedirectToLeaderBoardHistory)

	router.POST("/account/info/description/confirm", authHandler.WithJWT, h.ConfirmDescriptionCreation)

	router.GET("/home", h.LoadHome)
	router.GET("/", h.LoadHome)
}

func (h *userHandler) LoadHome(c *gin.Context) {
	overall.Home().Render(c.Request.Context(), c.Writer)
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
		user.Signup("Такий користувач вже інсує").Render(c.Request.Context(), c.Writer)
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
		user.Login("Дані не є валідними, будь ласка попробуйте ще раз").Render(c.Request.Context(), c.Writer)
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
		user.Login("Дані не є валідними, будь ласка попробуйте ще раз").Render(c.Request.Context(), c.Writer)
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

func (h *userHandler) LoadUserAccount(c *gin.Context) {
	userID := c.Param("userID")
	if userID != "" {
		usr, err := h.userStore.GetUserById(userID)
		if err != nil {
			log.Println("error : ", err)
			return
		}
		if usr.UserName == "" {
			u, _ := c.Get("user")
			user.LoadUserAccount(u.(*types.User), true).Render(context.Background(), c.Writer)
			return
		}
		log.Println("user : ", usr)
		user.LoadUserAccount(usr, false).Render(context.Background(), c.Writer)
		return
	} else {
		u, _ := c.Get("user")
		user.LoadUserAccount(u.(*types.User), true).Render(context.Background(), c.Writer)
		return
	}

}

func (h *userHandler) ConfirmDescriptionCreation(c *gin.Context) {
	u, _ := c.Get("user")
	description := h.clientSideHandler.GetDataFromForm(c, "description")
	log.Println("description : ", description)
	err := h.userStore.UpdateDescription(u.(*types.User).ID, description)
	if err != nil {
		log.Println("error : ", err)
		return
	}
	c.Writer.Header().Add("HX-Redirect", "/account/info")

}

func (h *userHandler) RedirectToLeaderBoardHistory(c *gin.Context) {
	c.Writer.Header().Add("HX-Redirect", "/account/info/leaderboard-history")
}

func (h *userHandler) LoadUserLeaderBoardHistory(c *gin.Context) {
	userID := c.Param("userID")

	if userID != "" {
		usr, err := h.userStore.GetUserById(userID)
		if err != nil {
			log.Println("error : ", err)
			return
		}

		if usr.UserName == "" {
			u, _ := c.Get("user")
			leaderboard, hasGamesPlayed, err := h.userStore.GetUserGameScore(u.(*types.User).UserName)
			if err != nil {
				log.Println("error : ", err)
				return
			}
			user.LoadUserLeaderBoardHistory(hasGamesPlayed, leaderboard).Render(context.Background(), c.Writer)

			return
		}
		log.Println("user : ", usr)
		user.LoadUserAccount(usr, false).Render(context.Background(), c.Writer)
		return
	} else {
		u, _ := c.Get("user")
		usr := u.(*types.User)
		leaderboard, hasGamesPlayed, err := h.userStore.GetUserGameScore(usr.UserName)
		if err != nil {
			log.Println("error : ", err)
			return
		}
		log.Println("hasGamesPlayed: ", hasGamesPlayed)
		log.Println("leaderboard : ", leaderboard)
		user.LoadUserLeaderBoardHistory(hasGamesPlayed, leaderboard).Render(context.Background(), c.Writer)

	}
}
