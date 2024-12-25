package game

import (
	"context"
	"log"
	"sync"

	"github.com/Megidy/k/types"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

//TO DO :

//1 : make password

//2 : when client connects twice from one account make second request unavaible ,maybe 401 or smth like that

//3 : make normal leave and connection while game for players who was in stock -> write normal isReady statement.
//
//	there are 1 solution probably made , but implement this right and add timer.

// 4 : when new client connects , those who answered qeustion dont see him in waitList
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
	//channel to handle safe clients leave
	clientsLeaveCh chan *Client
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
		clientsLeaveCh:           make(chan *Client),
		addClientToWaitList:      make(chan *Client),
		removeClientFromWaitList: make(chan *Client),
		updateQuestionCh:         make(chan bool),
	}
	go manager.MessageQueue()
	go manager.ClientsStatusHandler()
	go manager.QuestionHandler()
	go manager.WaitListHandler()
	go manager.StartGame()

	return manager
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

		delete(m.stockMap, c.userName)
		wasInGameBefore = true
	} else {
		wasInGameBefore = false
	}
	// c2, ok2 := m.clientsMap[client.userName]
	// if !ok1 && ok2 {

	// }
	m.clientsMap[client.userName] = client
	log.Println("Added new client : ", client.userName)
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
	//start of the game
	<-m.startGameCh
	//starting w/r pumps
	m.clientsConnectionCh <- client

	if wasInGameBefore {
		if client.isReady {
			m.overwriteListCh <- client.userName
		} else {
			m.addClientToWaitList <- client
			m.overwriteQuestionCh <- client.userName
		}

	} else {
		if m.numberOfCurrentQuestion > 0 {
			log.Println("d")
			m.addClientToWaitList <- client
			m.overwriteQuestionCh <- client.userName
		}
		log.Println("started goroutine for new client : ", client.userName)
	}

}

func (m *Manager) MessageQueue() {
	for {
		select {
		//change question for all connected clients
		case <-m.changeQuestionCh:
			m.mu.Lock()
			for _, client := range m.clientsMap {
				m.waitList = append(m.waitList, client.userName)
				client.currQuestion = m.numberOfCurrentQuestion
				client.isReady = false
				client.questionCh <- m.currentQuestion
			}
			m.mu.Unlock()
		//write curr question for new client
		case <-m.changeListCh:
			m.mu.Lock()
			for _, client := range m.clientsMap {
				if client.isReady {
					client.writeWaitCh <- m.waitList
				}
			}
			m.mu.Unlock()
		//write curr waitList for new client
		case username := <-m.overwriteQuestionCh:
			m.mu.Lock()

			client := m.clientsMap[username]
			client.isReady = false
			client.currQuestion = m.numberOfCurrentQuestion
			client.questionCh <- m.currentQuestion
			m.mu.Unlock()
		//write list of not done clients if client is answered question
		case username := <-m.overwriteListCh:
			m.mu.Lock()
			client := m.clientsMap[username]
			client.writeWaitCh <- m.waitList
			m.mu.Unlock()
		}
	}
}

func (m *Manager) ClientsStatusHandler() {

	for {
		select {
		case client := <-m.readyCh:
			client.isReady = true
			m.removeClientFromWaitList <- client

		case <-m.clientsLeaveCh:
		case client := <-m.clientsConnectionCh:
			go client.ReadPump()
			go client.WritePump()
		}

	}
}

func (m *Manager) WaitListHandler() {
	for {
		select {
		case client := <-m.addClientToWaitList:
			m.waitList = append(m.waitList, client.userName)
			m.changeListCh <- true
		case client := <-m.removeClientFromWaitList:
			for index, value := range m.waitList {
				if value == client.userName {
					m.waitList = append(m.waitList[:index], m.waitList[index+1:]...)
				}
			}
			m.mu.Lock()
			length := len(m.clientsMap)
			m.mu.Unlock()
			if len(m.waitList) == 0 && length != 0 {
				m.updateQuestionCh <- true
			} else {
				m.changeListCh <- true
			}
		}
	}
}

func (m *Manager) QuestionHandler() {
	for {
		select {
		case <-m.updateQuestionCh:
			if m.numberOfCurrentQuestion == len(m.questions)-1 {

				//later : handler end of the game
				m.numberOfCurrentQuestion = 0
			} else {
				m.numberOfCurrentQuestion++

			}
			//changing question and sending message to channel that everyone is ready and new question can be delivered
			m.currentQuestion = m.questions[m.numberOfCurrentQuestion]
			m.changeQuestionCh <- true
		}
	}
}

func (m *Manager) StartGame() {
	for {
		// log.Println("clients pool : ", len(m.clientsMap))
		if len(m.clientsMap) == m.maxPlayers {
			log.Println("GAME STARTED !!")
			m.startGameCh <- true
			m.changeQuestionCh <- true
			close(m.startGameCh)
			m.statementOfGame = 1
			return
		}
	}
}
