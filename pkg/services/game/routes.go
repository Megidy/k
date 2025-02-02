package game

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Megidy/k/pkg/auth"
	"github.com/Megidy/k/static/templates/game"
	"github.com/Megidy/k/static/templates/room"
	"github.com/Megidy/k/static/templates/topic"
	"github.com/Megidy/k/types"
	"github.com/Megidy/k/worker"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//TO DO: Create oportunity to check question of existing topics

type gameHandler struct {
	clienSideHandler types.ClientSideHandler
	gameStore        types.GameStore
	WorkerPool       worker.WorkerManager
}

func NewGameHandler(clienSideHandler types.ClientSideHandler, gameStore types.GameStore, WorkerPool worker.WorkerManager) *gameHandler {
	return &gameHandler{
		clienSideHandler: clienSideHandler,
		gameStore:        gameStore,
		WorkerPool:       WorkerPool,
	}
}

func (h *gameHandler) RegisterRoutes(router gin.IRouter, authHandler *auth.Handler) {
	//game Handlers:
	//1 : handle template for game
	router.GET("/game/:roomID", authHandler.WithJWT, h.HandleGame)
	//handler information about room , will add questions
	router.GET("/game/:roomID/info/:players/:questions/:playstyle", authHandler.WithJWT, h.LoadInfoTempl)
	//confirm information about room ,redirect to the game handler : /game/:roomID
	router.POST("/game/:roomID/info/:players/:questions/:playstyle/confirm", authHandler.WithJWT, h.ConfirmInfo)
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
	playstyle := c.Param("playstyle")
	log.Println("topic avaible : ", topics)
	comp := room.LoadInfoPage(topics, []types.Topic{
		{
			TopicID: "dfd62fd8-672c-4737-80fe-4dfbedabedda",
			UserID:  "public",
			Name:    "Geography",
		},
	}, roomID, p, q, playstyle)
	comp.Render(context.Background(), c.Writer)
}
func (h *gameHandler) ConfirmInfo(c *gin.Context) {

	u, _ := c.Get("user")

	topic := h.clienSideHandler.GetDataFromForm(c, "topic")
	roomID := c.Param("roomID")
	_, ok := globalRoomManager.GetManager(roomID)
	if ok {
		b := make([]byte, 6)
		rand.Read(b)
		roomID = base64.StdEncoding.EncodeToString(b)
		roomID = strings.ReplaceAll(roomID, "/", "d")
	}
	p := c.Param("players")
	q := c.Param("questions")
	play := c.Param("playstyle")
	playstyle, err := strconv.Atoi(play)
	if err != nil {
		log.Println("error when getting playstyle param :", err)
		return
	}
	players, err := strconv.Atoi(p)
	if err != nil {
		log.Println("error when getting players param :", err)
		return
	}
	numberOfQuestions, err := strconv.Atoi(q)
	if err != nil {
		log.Println("error when getting questions param :", err)
		return
	}
	//find question with this topic
	log.Println("topic : ", topic)
	var questions []types.Question
	if topic == "Geography" {
		questions, err = h.gameStore.GetQuestionsByTopicName("Geography", "public")
		if err != nil {
			log.Println("error when getting question from db by topic : ", err)
			return
		}

	} else {
		questions, err = h.gameStore.GetQuestionsByTopicName(topic, u.(*types.User).ID)
		if err != nil {
			log.Println("error when getting question from db by topic : ", err)
			return
		}
	}
	log.Println("questions : ", questions)
	globalRoomManager.CreateRoom(h.WorkerPool, u.(*types.User).UserName, roomID, players, playstyle, numberOfQuestions, questions)
	url := fmt.Sprintf("/game/%s", roomID)
	c.Writer.Header().Add("HX-Redirect", url)

}

func (h *gameHandler) HandleGame(c *gin.Context) {
	u, _ := c.Get("user")
	user := u.(*types.User)
	roomID := c.Param("roomID")
	m, ok := globalRoomManager.GetManager(roomID)
	if !ok {
		game.Game(roomID, true, false, false, false).Render(c.Request.Context(), c.Writer)
		return
	}
	isDuplicate := globalRoomManager.CheckDuplicate(m, user.UserName)
	if isDuplicate {
		game.Game(roomID, true, true, false, false).Render(c.Request.Context(), c.Writer)
		return
	} else {
		if m.owner.username == user.UserName {
			if m.owner.playStyle == 2 {
				log.Println("spectating")
				game.Game(roomID, false, true, true, true).Render(c.Request.Context(), c.Writer)
			} else {
				log.Println("playing")
				game.Game(roomID, false, true, true, false).Render(c.Request.Context(), c.Writer)
			}

		} else {
			game.Game(roomID, false, true, false, false).Render(c.Request.Context(), c.Writer)
		}

	}
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
	play := h.clienSideHandler.GetDataFromForm(c, "type")
	var playstyle int
	if play == "I want to play" {
		playstyle = 1
	} else {
		playstyle = 2
	}
	log.Println("number of players :", numberOfPlayers)
	log.Println("ramountOfQuestions : ", amountOfQuestions)
	b := make([]byte, 6)
	rand.Read(b)
	roomID := base64.StdEncoding.EncodeToString(b)
	roomID = strings.ReplaceAll(roomID, "/", "d")
	//create cookie for connection secure

	url := fmt.Sprintf("/game/%s/info/%s/%s/%d", roomID, numberOfPlayers, amountOfQuestions, playstyle)
	c.Writer.Header().Add("HX-Redirect", url)
}

func (h *gameHandler) LoadTopicCreation(c *gin.Context) {
	comp := topic.LoadCreateTopic("")
	comp.Render(context.Background(), c.Writer)
}
func (h *gameHandler) ConfirmTopicCreation(c *gin.Context) {
	u, _ := c.Get("user")

	id := uuid.NewString()
	name := h.clienSideHandler.GetDataFromForm(c, "name")
	number := h.clienSideHandler.GetDataFromForm(c, "number")
	exists, err := h.gameStore.TopicNameAlreadyExists(u.(*types.User).ID, name)
	if err != nil {
		log.Println("error when checking existance of topic :", err)
		comp := topic.LoadCreateTopic(err.Error())
		comp.Render(context.Background(), c.Writer)
		return
	}
	if exists {
		comp := topic.LoadCreateTopic("topic with this name already exists!")
		comp.Render(context.Background(), c.Writer)
		return
	}
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
		return
	}
	u, _ := c.Get("user")
	exists, err := h.gameStore.TopicNameAlreadyExists(u.(*types.User).ID, name)
	if err != nil {
		c.Redirect(http.StatusMovedPermanently, "/topic/create")
		return
	}
	if exists {
		c.Redirect(http.StatusMovedPermanently, "/topic/create")
		return
	}
	comp := topic.LoadCreateQuestions(name, topicID, nums)
	comp.Render(context.Background(), c.Writer)
}

func (h *gameHandler) ConfirmCreationOfQuestion(c *gin.Context) {
	var questions []types.Question

	u, _ := c.Get("user")

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
	err = h.gameStore.CreateTopic(questions, u.(*types.User).ID)
	if err != nil {
		log.Println("error when creating topic :", err)
		return
	}

	c.Writer.Header().Add("HX-Redirect", "/room/create")
}
