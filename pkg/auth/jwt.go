package auth

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Megidy/k/types"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type Handler struct {
	UserStore types.UserStore
}

func NewHandler(userStore types.UserStore) *Handler {
	return &Handler{
		UserStore: userStore,
	}
}

func (h *Handler) WithJWT(c *gin.Context) {

	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		log.Println("error : ", err)
		c.Writer.Header().Add("authorization", "false")
		RedirectToLogin(c)
		return
	}
	token, err := ValidateJWT(tokenString)
	if err != nil {
		log.Println("error : ", err)
		c.Writer.Header().Add("authorization", "false")
		RedirectToLogin(c)
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.Writer.Header().Add("authorization", "false")
			RedirectToLogin(c)
			return
		}
		id := claims["userID"].(string)
		user, err := h.UserStore.GetUserById(id)
		if err != nil {
			log.Println("error : ", err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("blablabla-sosecretthaticantevenhideitpleasehelpme"), nil
	})
}

func CreateJWT(secret []byte, userId string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": userId,
		"exp":    time.Now().Add(time.Second * 60 * 24 * 30).Unix(),
	})
	tokenString, err := token.SignedString(secret)
	if err != nil {
		log.Println("error : ", err)
		return "", err
	}
	return tokenString, nil
}

func RedirectToLogin(c *gin.Context) {
	if c.Request.Method == http.MethodPost {
		c.Writer.Header().Add("HX-Redirect", "/account/login")
	} else {
		c.Redirect(http.StatusFound, "/account/login")
		c.Abort()
	}

}
