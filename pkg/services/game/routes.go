package game

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/Megidy/k/pkg/auth"
	"github.com/Megidy/k/static/game"
	"github.com/Megidy/k/static/room"
	"github.com/Megidy/k/static/topic"
	"github.com/Megidy/k/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type gameHandler struct {
	clienSideHandler types.ClientSideHandler
	gameStore        types.GameStore
}

func NewGameHandler(clienSideHandler types.ClientSideHandler, gameStore types.GameStore) *gameHandler {
	return &gameHandler{
		clienSideHandler: clienSideHandler,
		gameStore:        gameStore,
	}
}

func (h *gameHandler) RegisterRoutes(router gin.IRouter, authHandler *auth.Handler) {
	//game Handlers:
	//1 : handle template for game
	router.GET("/game/:roomID", authHandler.WithJWT, h.HandleGame)
	//handler information about room , will add questions
	router.GET("/game/:roomID/info/:players/:questions", authHandler.WithJWT, h.LoadInfoTempl)
	//confirm information about room ,redirect to the game handler : /game/:roomID
	router.POST("/game/:roomID/info/:players/:questions/confirm", authHandler.WithJWT, h.ConfirmInfo)
	//2 : handle ws connection for game
	router.GET("/ws/game/handler/:roomID", authHandler.WithJWT, h.HandleWSConnectionForGame)

	//connection to room
	router.GET("/room/connect", authHandler.WithJWT, h.LoadTemplConnectToRoom)
	router.POST("/room/connect/confirm", authHandler.WithJWT, h.ConfirmConnectionToRoom)

	//room creation
	router.GET("/room/create", authHandler.WithJWT, h.LoadTemplCreateRoom)
	router.POST("/room/create/confirm", authHandler.WithJWT, h.ConfirmCreationOfRoom)
	//creating question
	router.GET("/topic/create", authHandler.WithJWT, h.LoadTopicCreation)
	router.POST("/topic/create/confirm", authHandler.WithJWT, h.ConfirmTopicCreation)
	router.GET("/topic/:topicID/:name/:amountOfQuestions/questions", authHandler.WithJWT, h.LoadCreationOfQuestions)
	router.POST("/topic/:topicID/:name/:amountOfQuestions/questions/confirm", authHandler.WithJWT, h.ConfirmCreationOfQuestion)
}

func (h *gameHandler) LoadInfoTempl(c *gin.Context) {
	roomID := c.Param("roomID")
	u, _ := c.Get("user")
	user := u.(*types.User)
	log.Println("user :", user)
	topics, err := h.gameStore.GetUsersTopics(user.ID)
	if err != nil {
		log.Println("error : ", err)
		return
	}
	p := c.Param("players")

	q := c.Param("questions")

	log.Println("topic avaible : ", topics)
	comp := room.LoadInfoPage(topics, topics, roomID, p, q)
	comp.Render(context.Background(), c.Writer)

}
func (h *gameHandler) ConfirmInfo(c *gin.Context) {
	topic := h.clienSideHandler.GetDataFromForm(c, "topic")
	roomID := c.Param("roomID")
	p := c.Param("players")

	q := c.Param("questions")
	players, err := strconv.Atoi(p)
	if err != nil {
		log.Println("error when getting players query :", err)
		return
	}
	numberOfQuestions, err := strconv.Atoi(q)
	if err != nil {
		log.Println("error when getting questions query :", err)
		return
	}
	//find question with this topic
	log.Println("topic : ", topic)
	questions, err := h.gameStore.GetQuestionsByTopicName(topic)
	if err != nil {
		log.Println("error when getting question from db by topic : ", err)
		return
	}
	log.Println("questions : ", questions)
	globalRoomManager.CreateRoom(roomID, players, numberOfQuestions, questions)
	url := fmt.Sprintf("/game/%s", roomID)
	c.Writer.Header().Add("HX-Redirect", url)
	//redirect to game handler : /game/:roomID
}

func (h *gameHandler) HandleGame(c *gin.Context) {
	roomID := c.Param("roomID")
	game.Game(roomID).Render(c.Request.Context(), c.Writer)
}

func (h *gameHandler) HandleWSConnectionForGame(c *gin.Context) {
	roomID := c.Param("roomID")
	log.Println("room id : ", roomID)
	manager, exists := globalRoomManager.GetManager(roomID)
	if !exists {

		return
	}
	manager.NewConnection(c)
}

func (h *gameHandler) LoadTemplConnectToRoom(c *gin.Context) {
	comp := room.LoadConnectionToRoom("")
	comp.Render(c.Request.Context(), c.Writer)

}
func (h *gameHandler) ConfirmConnectionToRoom(c *gin.Context) {
	//get data from form
	roomID := h.clienSideHandler.GetDataFromForm(c, "code")
	_, ok := globalRoomManager.GetManager(roomID)
	if !ok {
		log.Println("room not found")
		room.LoadConnectionToRoom("room not found :(").Render(c.Request.Context(), c.Writer)
		return
	}
	url := fmt.Sprintf("/game/%s", roomID)
	c.Writer.Header().Add("HX-Redirect", url)
}

func (h *gameHandler) LoadTemplCreateRoom(c *gin.Context) {
	comp := room.LoadCreationOfRoom()
	comp.Render(context.Background(), c.Writer)
}
func (h *gameHandler) ConfirmCreationOfRoom(c *gin.Context) {

	//read data about room from form

	// handle creation of room:
	//number of players which will play
	numberOfPlayers := h.clienSideHandler.GetDataFromForm(c, "players")
	//amount of question they will answer
	amountOfQuestions := h.clienSideHandler.GetDataFromForm(c, "questions")
	log.Println("number of players :", numberOfPlayers)
	log.Println("ramountOfQuestions : ", amountOfQuestions)
	b := make([]byte, 6)
	rand.Read(b)
	roomID := base64.StdEncoding.EncodeToString(b)
	roomID = strings.ReplaceAll(roomID, "/", "d")
	//create cookie for connection secure
	url := fmt.Sprintf("/game/%s/info/%s/%s", roomID, numberOfPlayers, amountOfQuestions)
	c.Writer.Header().Add("HX-Redirect", url)
}

func (h *gameHandler) LoadTopicCreation(c *gin.Context) {
	comp := topic.LoadCreateTopic()
	comp.Render(context.Background(), c.Writer)
}
func (h *gameHandler) ConfirmTopicCreation(c *gin.Context) {
	id := uuid.NewString()
	name := h.clienSideHandler.GetDataFromForm(c, "name")
	number := h.clienSideHandler.GetDataFromForm(c, "number")

	url := fmt.Sprintf("/topic/%s/%s/%s/questions", id, name, number)
	c.Writer.Header().Add("HX-Redirect", url)
}

func (h *gameHandler) LoadCreationOfQuestions(c *gin.Context) {

	topicID := c.Param("topicID")
	name := c.Param("name")
	amount := c.Param("amountOfQuestions")
	nums, err := strconv.Atoi(amount)
	if err != nil {
		log.Println("error when converting  : ", err)
	}
	comp := topic.LoadCreateQuestions(name, topicID, nums)
	comp.Render(context.Background(), c.Writer)
}

func (h *gameHandler) ConfirmCreationOfQuestion(c *gin.Context) {
	var questions []types.Question

	var topic types.Topic

	name := c.Param("name")
	topicID := c.Param("topicID")
	topic.Name = name
	topic.TopicID = topicID
	number := c.Param("amountOfQuestions")
	num, err := strconv.Atoi(number)
	if err != nil {
		log.Println("bad url parametr : amountOfQuestions")
		return
	}
	for i := 0; i < num; i++ {
		var q types.Question
		q.Answers = make([]string, 4)
		q.Topic = &topic
		q.Id = uuid.NewString()
		q.Question = h.clienSideHandler.GetDataFromForm(c, fmt.Sprintf("name-%d", i))
		q.Answers[0] = h.clienSideHandler.GetDataFromForm(c, fmt.Sprintf("a-1-%d", i))
		q.Answers[1] = h.clienSideHandler.GetDataFromForm(c, fmt.Sprintf("a-2-%d", i))
		q.Answers[2] = h.clienSideHandler.GetDataFromForm(c, fmt.Sprintf("a-3-%d", i))
		q.Answers[3] = h.clienSideHandler.GetDataFromForm(c, fmt.Sprintf("a-4-%d", i))
		ca := h.clienSideHandler.GetDataFromForm(c, fmt.Sprintf("correctA-%d", i))
		n, err := strconv.Atoi(ca)
		if err != nil {
			log.Println("error when converting correct answer :", err)
			return
		}
		n = n - 1
		q.CorrectAnswer = q.Answers[n]
		image := h.clienSideHandler.GetDataFromForm(c, fmt.Sprintf("image-%d", i))
		if image == "" {
			q.Type = "text"
			q.ImageLink = "NONE"
		} else {
			q.Type = "image"
			q.ImageLink = image
		}
		log.Println("Question №", i+1, " : ", q)
		questions = append(questions, q)
	}
	u, _ := c.Get("user")
	err = h.gameStore.CreateTopic(questions, u.(*types.User).ID)
	if err != nil {
		log.Println("error when creating topic :", err)
		return
	}

}
