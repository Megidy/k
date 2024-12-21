package types

import "github.com/gin-gonic/gin"

type User struct {
	ID       string
	UserName string
	Email    string
	Password string
}
type Topic struct {
	TopicID string
	UserID  string
	Name    string
}
type Question struct {
	Id            string
	Topic         *Topic
	Type          string
	ImageLink     string
	Question      string
	Answers       []string
	CorrectAnswer string
}
type Message struct {
	Headers map[string]any
	Answer  map[string]string
}

type ClientSideHandler interface {
	GetDataFromForm(context *gin.Context, key string) string
}

type GameStore interface {
	CreateTopic(questions []Question) error
	GetUsersTopics(userID string) ([]Topic, error)
	GetQuestionsByTopicName(TopicName string) ([]Question, error)
}
type UserStore interface {
	GetUserById(id string) (*User, error)
	CreateUser(user *User) error
	UserExists(user *User) (bool, error)
	GetUserByEmail(email string) (*User, error)
}
