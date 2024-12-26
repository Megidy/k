package game

import (
	"context"
	"log"
	"sort"
	"sync"

	"github.com/Megidy/k/types"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type Manager struct {
	//statement of game
	//0 - not started yet
	//1 - in progress
	//-1 - ended
	statementOfGame int
	//unique ID of room
	roomID string
	//number of members
	maxPlayers int
	//number of questions
	numberOfQuestions int
	//mutex for concurenct safe reading and writing
	mu sync.Mutex
	//context
	ctx context.Context
	//cancel func of context
	cancel context.CancelFunc
	//number of current question
	numberOfCurrentQuestion int
	//questions
	questions []types.Question
	//curr question
	currentQuestion types.Question
	//clientsMap
	clientsMap map[string]*Client
	//stockMap for not repition of clients
	stockMap map[string]*Client
	//change question chan
	changeQuestionCh chan bool
	//change list for all clients
	changeListCh chan bool
	//overwrite question to new client
	overwriteQuestionCh chan string
	//overwrite list to new client if he answered question before
	overwriteListCh chan string
	//list of players who havent answered question yet
	waitList []string
	//channel for getting clients request about readiness
	readyCh chan *Client
	//channel to handle th connection
	clientsConnectionCh chan *Client
	//channel to start the game
	startGameCh chan bool
	//channel to add client to waitList after he connected to the game
	addClientToWaitList chan *Client
	//channel to remove client from wait list after he disconnected from the game
	removeClientFromWaitList chan *Client
	//channel to update notify when shoul question has to be updated
	updateQuestionCh chan bool
	//channel to load leaderboard
	loadLeaderBoardCh chan []types.Player
}

func NewManager(roomID string, numberOfPlayers, amountOfQuestions int, questions []types.Question) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	manager := &Manager{

		roomID:                   roomID,
		maxPlayers:               numberOfPlayers,
		numberOfQuestions:        amountOfQuestions,
		questions:                questions,
		mu:                       sync.Mutex{},
		ctx:                      ctx,
		cancel:                   cancel,
		currentQuestion:          questions[0],
		clientsMap:               make(map[string]*Client),
		stockMap:                 make(map[string]*Client),
		changeQuestionCh:         make(chan bool),
		changeListCh:             make(chan bool),
		overwriteQuestionCh:      make(chan string),
		overwriteListCh:          make(chan string),
		readyCh:                  make(chan *Client),
		startGameCh:              make(chan bool),
		clientsConnectionCh:      make(chan *Client),
		addClientToWaitList:      make(chan *Client),
		removeClientFromWaitList: make(chan *Client),
		updateQuestionCh:         make(chan bool),
		loadLeaderBoardCh:        make(chan []types.Player),
	}
	go manager.MessageQueue()
	go manager.ClientsStatusHandler()
	go manager.QuestionHandler()
	go manager.WaitListHandler()
	go manager.StartGame()

	return manager
}

func (m *Manager) ScoreHandler(client *Client, requestData *types.RequestData) {
	question := m.questions[client.currQuestion]

	if requestData.Answer == question.CorrectAnswer {
		log.Println("added point for : ", client)
		client.score++
	}
}

func (m *Manager) AddClientToConnectionPool(client *Client) bool {
	var wasInGameBefore bool
	m.mu.Lock()
	//checking if client was connected before
	c, ok1 := m.stockMap[client.userName]
	if ok1 {
		if c.currQuestion == m.numberOfCurrentQuestion {
			if c.isReady {
				client.isReady = true
			} else {
				client.isReady = false
			}
		} else {
			client.isReady = false
		}
		client.currQuestion = c.currQuestion
		client.score = c.score
		delete(m.stockMap, c.userName)
		wasInGameBefore = true
	} else {
		wasInGameBefore = false
	}
	// c2, ok2 := m.clientsMap[client.userName]
	// if !ok1 && ok2 {

	// }
	m.clientsMap[client.userName] = client
	log.Println("Added new client to connection pool : ", client.userName)
	m.mu.Unlock()

	return wasInGameBefore
}

func (m *Manager) DeleteClientFromConnectionPool(client *Client) {
	m.mu.Lock()
	m.removeClientFromWaitList <- client
	m.stockMap[client.userName] = client
	delete(m.clientsMap, client.userName)
	m.mu.Unlock()
	globalRoomManager.DeleteConnectionFromList(m, client.userName)
}

// implement function start the game
func (m *Manager) NewConnection(c *gin.Context) {
	//getting user from data
	u, _ := c.Get("user")
	user := u.(*types.User)
	//upgrading connection aka 'handshake'
	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("error while creating websocket connection : ", err)
		return
	}
	//creating new client
	client := NewClient(user.UserName, wsConn, m, c)

	//adding him to connection pool, meanwhile cheking if he was connected before
	wasInGameBefore := m.AddClientToConnectionPool(client)
	globalRoomManager.AddConnectionToList(m, client.userName)
	//starting w/r pumps
	m.clientsConnectionCh <- client
	// <-updateQueueCh
	//start of the game
	<-m.startGameCh
	if wasInGameBefore {
		if client.isReady {
			m.overwriteListCh <- client.userName
		} else {
			m.addClientToWaitList <- client
			m.overwriteQuestionCh <- client.userName
		}
		log.Println("added new client who was in game before")

	} else {
		if m.statementOfGame == 1 {
			log.Println("d")
			m.addClientToWaitList <- client
			m.overwriteQuestionCh <- client.userName
			log.Println("added new client")

		}
		log.Println("started goroutine for new client : ", client.userName)
	}

}

func (m *Manager) MessageQueue() {
	defer func() {
		log.Println("MESSAGE QUEUE | exited goroutine")
		close(m.overwriteListCh)
		close(m.overwriteQuestionCh)
		close(m.readyCh)
		close(m.clientsConnectionCh)
		close(m.addClientToWaitList)
		globalRoomManager.EndRoomSession(m.roomID)
	}()
	for {
		select {
		//change question for all connected clients
		case _, ok := <-m.changeQuestionCh:
			if !ok {
				log.Println("tried to read from closed changeQuestionCh in MessageQueue")
				return
			}
			m.mu.Lock()
			for _, client := range m.clientsMap {
				// m.waitList = append(m.waitList, client.userName)
				client.currQuestion = m.numberOfCurrentQuestion
				client.isReady = false
				client.questionCh <- m.currentQuestion
			}
			m.mu.Unlock()
		//write curr question for new client
		case _, ok := <-m.changeListCh:
			if !ok {
				log.Println("tried to read from closed changeListCh in MessageQueue")
				return
			}
			m.mu.Lock()
			for _, client := range m.clientsMap {

				if client.isReady {
					client.writeWaitCh <- m.waitList
				}
			}
			m.mu.Unlock()
		//write curr waitList for new client
		case username, ok := <-m.overwriteQuestionCh:
			if !ok {
				log.Println("tried to read from closed overwriteQuestionCh in MessageQueue")
				return
			}
			m.mu.Lock()

			client := m.clientsMap[username]
			client.isReady = false
			client.currQuestion = m.numberOfCurrentQuestion
			client.questionCh <- m.currentQuestion
			m.mu.Unlock()
		//write list of not done clients if client is answered question
		case username, ok := <-m.overwriteListCh:
			if !ok {
				log.Println("tried to read from closed overwriteListCh in MessageQueue")
				return
			}
			m.mu.Lock()
			client := m.clientsMap[username]
			client.writeWaitCh <- m.waitList
			m.mu.Unlock()
		case players, ok := <-m.loadLeaderBoardCh:
			if !ok {
				log.Println("tried to read from closed loadLeaderBoardCh in MessageQueue")
				return
			}
			m.mu.Lock()
			for _, client := range m.clientsMap {
				client.leaderBoardCh <- players
			}
			m.mu.Unlock()
			m.cancel()
			return
		}
	}
}

func (m *Manager) ClientsStatusHandler() {
	defer func() {
		log.Println("CLIENTSSTATUSHANDLER | exited goroutine")
		close(m.removeClientFromWaitList)
	}()
	for {
		select {
		case <-m.ctx.Done():
			return
		case client, ok := <-m.readyCh:
			if !ok {
				log.Println("tried to read from closed readyCh in ClientsStatusHandler")
				return
			}
			client.isReady = true
			m.removeClientFromWaitList <- client
		case client, ok := <-m.clientsConnectionCh:
			if !ok {
				log.Println("tried to read from closed clientsConnectionCh in ClientsStatusHandler")
				return
			}
			go client.ReadPump()
			go client.WritePump()
		}

	}
}

func (m *Manager) WaitListHandler() {
	defer func() {
		log.Println("WAITLISTHANDLER | exited goroutine")
		close(m.changeListCh)
		close(m.updateQuestionCh)
	}()
	for {
		select {
		case <-m.ctx.Done():
			return
		case client, ok := <-m.addClientToWaitList:
			if !ok {
				log.Println("tried to read from closed addClientToWaitList in WaitListHandler")
				return
			}
			m.waitList = append(m.waitList, client.userName)
			m.changeListCh <- true
		case client, ok := <-m.removeClientFromWaitList:
			if !ok {
				log.Println("tried to read from closed removeClientFromWaitList in WaitListHandler")
				return
			}
			newWaitList := []string{}
			for _, value := range m.waitList {
				if value != client.userName {
					newWaitList = append(newWaitList, value)
				}
			}
			m.waitList = newWaitList
			m.mu.Lock()
			length := len(m.clientsMap)
			m.mu.Unlock()
			if len(m.waitList) == 0 && length != 0 {
				m.mu.Lock()
				for username := range m.clientsMap {
					m.waitList = append(m.waitList, username)
				}
				m.mu.Unlock()
				m.updateQuestionCh <- true
			} else {
				m.changeListCh <- true
			}
		}
	}
}

func (m *Manager) QuestionHandler() {
	defer func() {
		log.Println("QUESTIONHANDLER | exited goroutine")
		close(m.changeQuestionCh)
		close(m.loadLeaderBoardCh)
	}()
	for {
		select {
		case <-m.ctx.Done():
			return
		case _, ok := <-m.updateQuestionCh:
			if !ok {
				log.Println("tried to read from closed updateQuestionCh in questionHandler")
				return
			}
			if m.numberOfCurrentQuestion == m.numberOfQuestions-1 {

				log.Println("game ended")

				var leaderBoard = make(map[string]int)
				m.mu.Lock()
				for name, client := range m.clientsMap {
					leaderBoard[name] = client.score
				}

				m.mu.Unlock()
				players := make([]types.Player, 0)
				for name, points := range leaderBoard {
					players = append(players, types.Player{Username: name, Score: points})
				}
				sort.Slice(players, func(i, j int) bool {
					return players[i].Score > players[j].Score

				})
				m.loadLeaderBoardCh <- players
			} else {
				m.numberOfCurrentQuestion++
				//changing question and sending message to channel that everyone is ready and new question can be delivered
				m.currentQuestion = m.questions[m.numberOfCurrentQuestion]
				log.Println("current question : ", m.currentQuestion)
				m.changeQuestionCh <- true
			}

		}
	}
}

func (m *Manager) StartGame() {
	defer func() {
		log.Println("STARTGAME | GAME STARTED!")
		log.Println("STARTGAME | exited goroutine")
		close(m.startGameCh)
	}()
	for {
		// log.Println("clients pool : ", len(m.clientsMap))
		if len(m.clientsMap) == m.maxPlayers {
			m.startGameCh <- true
			m.changeQuestionCh <- true
			m.statementOfGame = 1
			return
		}
	}
}
