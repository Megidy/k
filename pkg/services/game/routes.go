package game

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/Megidy/k/static/game"
	"github.com/Megidy/k/static/room"
	"github.com/Megidy/k/static/topic"
	"github.com/Megidy/k/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GameHandler struct {
	clienSideHandler types.ClientSideHandler
	gameStore        types.GameStore
}

func NewGameHandler(clienSideHandler types.ClientSideHandler, gameStore types.GameStore) *GameHandler {
	return &GameHandler{
		clienSideHandler: clienSideHandler,
		gameStore:        gameStore,
	}
}

func (h *GameHandler) RegisterRoutes(router gin.IRouter) {
	//game Handlers:
	//1 : handle template for game
	router.GET("/game/:roomID", h.HandleGame)
	//handler information about room , will add questions
	router.GET("/game/:roomID/info", h.LoadInfoTempl)
	//confirm information about room ,redirect to the game handler : /game/:roomID
	router.POST("/game/:roomID/info/confirm", h.ConfirmInfo)
	//2 : handle ws connection for game
	router.GET("/ws/game/handler/:roomID", h.HandleWSConnectionForGame)

	//connection to room
	router.GET("/room/connect", h.LoadTemplConnectToRoom)
	router.POST("/room/connect/confirm", h.ConfirmConnectionToRoom)

	//room creation
	router.GET("/room/create", h.LoadTemplCreateRoom)
	router.POST("/room/create/confirm", h.ConfirmCreationOfRoom)
	//creating question
	router.GET("/topic/create", h.LoadTopicCreation)
	router.POST("/topic/create/confirm", h.ConfirmTopicCreation)
	router.GET("/topic/:topicID/:name/:amountOfQuestions/questions", h.LoadCreationOfQuestions)
	router.POST("/topic/:topicID/:name/:amountOfQuestions/questions/confirm", h.ConfirmCreationOfQuestion)
}

func (h *GameHandler) LoadInfoTempl(c *gin.Context) {
	roomID := c.Param("roomID")
	topics, err := h.gameStore.GetUsersTopics("35fbe6ad-7fe8-4ee7-9732-bae69f3ee80a")
	if err != nil {
		log.Println("error : ", err)
		return
	}
	log.Println("topic avaible : ", topics)
	comp := room.LoadInfoPage(topics, topics, roomID)
	comp.Render(context.Background(), c.Writer)

}
func (h *GameHandler) ConfirmInfo(c *gin.Context) {
	topic := h.clienSideHandler.GetDataFromForm(c, "topic")
	roomID := c.Param("roomID")
	//find question with this topic
	log.Println("topic : ", topic)

	url := fmt.Sprintf("/game/%s", roomID)
	c.Writer.Header().Add("HX-Redirect", url)
	//redirect to game handler : /game/:roomID
}

func (h *GameHandler) HandleGame(c *gin.Context) {
	roomID := c.Param("roomID")
	game.Game(roomID).Render(c.Request.Context(), c.Writer)
}

func (h *GameHandler) HandleWSConnectionForGame(c *gin.Context) {
	roomID := c.Param("roomID")
	log.Println("room id : ", roomID)
	manager, exists := globalRoomManager.GetManager(roomID)
	if !exists {

		return
	}
	manager.NewConnection(c)
}

func (h *GameHandler) LoadTemplConnectToRoom(c *gin.Context) {
	comp := room.LoadConnectionToRoom("")
	comp.Render(c.Request.Context(), c.Writer)

}
func (h *GameHandler) ConfirmConnectionToRoom(c *gin.Context) {
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

func (h *GameHandler) LoadTemplCreateRoom(c *gin.Context) {
	comp := room.LoadCreationOfRoom()
	comp.Render(context.Background(), c.Writer)
}
func (h *GameHandler) ConfirmCreationOfRoom(c *gin.Context) {

	//read data about room from form

	// handle creation of room:
	//number of players which will play
	numberOfPlayers := h.clienSideHandler.GetDataFromForm(c, "players")
	//amount of question they will answer
	amountOfQuestions := h.clienSideHandler.GetDataFromForm(c, "questions")
	log.Println("number of players :", numberOfPlayers)
	log.Println("ramountOfQuestions : ", amountOfQuestions)

	roomID := globalRoomManager.CreateRoom()

	//create cookie for connection secure
	url := fmt.Sprintf("/game/%s/info", roomID)
	c.Writer.Header().Add("HX-Redirect", url)
}

func (h *GameHandler) LoadTopicCreation(c *gin.Context) {
	comp := topic.LoadCreateTopic()
	comp.Render(context.Background(), c.Writer)
}
func (h *GameHandler) ConfirmTopicCreation(c *gin.Context) {
	id := uuid.NewString()
	name := h.clienSideHandler.GetDataFromForm(c, "name")
	number := h.clienSideHandler.GetDataFromForm(c, "number")

	url := fmt.Sprintf("/topic/%s/%s/%s/questions", id, name, number)
	c.Writer.Header().Add("HX-Redirect", url)
}

func (h *GameHandler) LoadCreationOfQuestions(c *gin.Context) {

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

func (h *GameHandler) ConfirmCreationOfQuestion(c *gin.Context) {
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
		q.CorrectAnswer = h.clienSideHandler.GetDataFromForm(c, fmt.Sprintf("correctA-%d", i))
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
	err = h.gameStore.CreateTopic(questions)
	if err != nil {
		log.Println("error when creating topic :", err)
		return
	}

}
