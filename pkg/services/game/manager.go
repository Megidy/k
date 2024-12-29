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

// TO DO
// 1 create timer which will later be dependent on the time of question
// 2 create force start of the game
// 3 make oportunity for owner of the session to spectate and collect data or play whi players
type Manager struct {
	//statement of game
	//0 - not started yet
	//1 - game just started , question =1
	//2 - game in progress
	//-1 - ended
	gameState int
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
	//channels to update curr players before the game
	beforeGameConnection chan bool
	beforeGameLeave      chan bool
	//channel to start gmae before all connnections are established
	forcedStartOfGame chan bool
}

//constructor

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
		beforeGameConnection:     make(chan bool),
		beforeGameLeave:          make(chan bool),
		forcedStartOfGame:        make(chan bool),
	}
	go manager.Writer()
	go manager.ClientsStatusHandler()
	go manager.QuestionHandler()
	go manager.WaitListHandler()
	go manager.StartGame()

	return manager
}

//function to handle scoreness of client

func (m *Manager) ScoreHandler(client *Client, requestData *types.RequestData) {

	//getting current question
	question := m.questions[client.currQuestion]

	//comparing data
	if requestData.Answer == question.CorrectAnswer {
		log.Println("added point for : ", client)
		client.score++
	}
}

//function which handles connection of clients

func (m *Manager) AddClientToConnectionPool(client *Client) bool {
	//varriable to check if client was in game before(was in stock map)
	var wasInGameBefore bool

	m.mu.Lock()

	//checking if client was connected before
	c, ok1 := m.stockMap[client.userName]
	if ok1 {
		//checking if he is joining to the game on the same question on which he leaved ,
		//if yes than handle isReady field,otherwise just set false , because he connected to new question
		if c.currQuestion == m.numberOfCurrentQuestion {
			//if he was ready than set isReady=true, otherwise false
			if c.isReady {
				client.isReady = true

			} else {
				client.isReady = false
			}
		} else {
			client.isReady = false
		}
		//setting currQuestion of client who was in stash to client who is connected , purpose of this is to handle correctness of question handling
		//problem was like : if he leaved on first question once , than he connects, there will be default  client.currQuestion = 0 , which could occur some problems
		client.currQuestion = c.currQuestion
		//same thing with score ,purpose of this is just ot handle correctess of results
		client.score = c.score
		//than deleting client from stockMap , because he connected
		delete(m.stockMap, c.userName)
		//and setting wasInGameBefore = true becuase he appeared in stockMap before
		wasInGameBefore = true
	} else {
		wasInGameBefore = false
	}
	//than adding this user to map
	m.clientsMap[client.userName] = client
	log.Println("Added new client to connection pool : ", client.userName)
	m.mu.Unlock()
	// this checking of game state was made for writing to beforeGameCh for each client, so if gameState=0 ,than game didnt started yet
	// so this updates list of players who is connected now
	if m.gameState == 0 {
		m.beforeGameConnection <- true
	}
	return wasInGameBefore
}

//function to delete client from connection pool and from globalRoomManager

func (m *Manager) DeleteClientFromConnectionPool(client *Client) {
	m.mu.Lock()
	//removing client from waitList to update him for other players
	m.removeClientFromWaitList <- client
	//checking if game is started , if yes than setting him to stockMap
	//purpose of this check , if players leaves before game started it will make errors
	if m.gameState == 1 || m.gameState == 2 {
		m.stockMap[client.userName] = client
	}

	delete(m.clientsMap, client.userName)
	log.Println("deleted from clientsMap username : ", client.userName)
	m.mu.Unlock()
	//deleteting from globalRoomManager
	globalRoomManager.DeleteConnectionFromList(m, client.userName)
	//updating beforeGame list of players
	if m.gameState == 0 {
		m.beforeGameLeave <- true
	}
}

//Funtion to handle Websocket connection and handle creating of client

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

	// channel for waiting start of the game
	<-m.startGameCh

	//purpose of this check : handle updates in render ONLY if game is started, because it could occur error
	if m.gameState == 1 || m.gameState == 2 {
		//checking if player was in game before
		//if he was in game before than checking if he is already answered question and render for him waitList, if not than render question
		//if he wasnt in game before than writing him questions and also adding to waitList to update for all plyers
		if wasInGameBefore {
			if client.isReady {
				m.overwriteListCh <- client.userName
			} else {
				if m.gameState == 2 {
					m.mu.Lock()
					_, ok := m.clientsMap[client.userName]
					m.mu.Unlock()
					if !ok {
						return
					}
				}
				m.addClientToWaitList <- client
				m.overwriteQuestionCh <- client.userName
			}
			log.Println("added new client who was in game before")

		} else {
			if m.gameState == 2 {
				m.mu.Lock()
				_, ok := m.clientsMap[client.userName]
				m.mu.Unlock()
				if !ok {
					return
				}
				m.addClientToWaitList <- client
				m.overwriteQuestionCh <- client.userName
				log.Println("added new client")
				log.Println("started goroutine for new client : ", client.userName)
			}
		}

	}
}

//function to handle writing to players

func (m *Manager) Writer() {
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

			//writing the question for all players
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
			//writing wait list for all players who is not ready
			for _, client := range m.clientsMap {
				if client.isReady {
					client.writeWaitCh <- m.waitList
				}
			}
			m.mu.Unlock()
		//write curr questine for new client
		case username, ok := <-m.overwriteQuestionCh:
			if !ok {
				log.Println("tried to read from closed overwriteQuestionCh in MessageQueue")
				return
			}
			m.mu.Lock()
			//writing question for player who connected and is ready
			client, ok := m.clientsMap[username]
			if !ok {
				log.Println("error when getting client : ", username)
				break
			}
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
			//writin waitList to new connected player
			client := m.clientsMap[username]
			client.writeWaitCh <- m.waitList
			m.mu.Unlock()
		//write leaderboard to all connected players
		case players, ok := <-m.loadLeaderBoardCh:
			if !ok {
				log.Println("tried to read from closed loadLeaderBoardCh in MessageQueue")
				return
			}
			m.mu.Lock()
			//writing leaderboadr to all connected players
			for _, client := range m.clientsMap {
				client.leaderBoardCh <- players
			}
			m.mu.Unlock()

			//calling context cancle func which will end end all goroutines and which will close all channels
			m.cancel()
			return
		}
	}
}

//function to handle status of clients such as :
//1 readiness
//2 connection

func (m *Manager) ClientsStatusHandler() {
	defer func() {
		log.Println("CLIENTSSTATUSHANDLER | exited goroutine")
		close(m.removeClientFromWaitList)
	}()
	for {
		select {
		case <-m.ctx.Done():
			return
		//case to handle readiness of client
		case client, ok := <-m.readyCh:
			if !ok {
				log.Println("tried to read from closed readyCh in ClientsStatusHandler")
				return
			}
			//setting client ready status to true
			client.isReady = true
			//removing him from waitList
			m.removeClientFromWaitList <- client
		//case to handle connection
		case client, ok := <-m.clientsConnectionCh:
			if !ok {
				log.Println("tried to read from closed clientsConnectionCh in ClientsStatusHandler")
				return
			}
			log.Println("started goroutines for : ", client.userName)
			//starting r/w goroutines
			go client.ReadPump()
			go client.WritePump()
		}

	}
}

//function to handle list update

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
		//case to add client to wait list after connection
		case client, ok := <-m.addClientToWaitList:
			if !ok {
				log.Println("tried to read from closed addClientToWaitList in WaitListHandler")
				return
			}
			m.waitList = append(m.waitList, client.userName)
			m.changeListCh <- true
		//case to remove client from wait list
		case client, ok := <-m.removeClientFromWaitList:
			if !ok {
				log.Println("tried to read from closed removeClientFromWaitList in WaitListHandler")
				return
			}

			//creating temporary waitList
			newWaitList := []string{}

			//looping through the waitList to overwrite him
			for _, value := range m.waitList {
				if value != client.userName {
					newWaitList = append(newWaitList, value)
				}
			}
			//set the waitList as a temp one
			m.waitList = newWaitList
			m.mu.Lock()
			length := len(m.clientsMap)
			m.mu.Unlock()
			// puprose of this check
			// len(m.waitList) == 0 is to check if waitList is empty -> everyone is ready than update question
			// length != 0 is to check if there are connected clients left, this check is important one , because for example:
			// there are 1 player left and he leaves, len(m.waitList) == 0 so this will go the next question
			if len(m.waitList) == 0 && length != 0 {
				//checking if game is started
				if m.gameState == 1 || m.gameState == 2 {
					m.mu.Lock()
					//looping througn the client map to fill the waitList
					for username := range m.clientsMap {
						m.waitList = append(m.waitList, username)
					}
					m.mu.Unlock()
					//updating question
					m.updateQuestionCh <- true
				}

			} else {
				//checking if game is started
				if m.gameState == 1 || m.gameState == 2 {
					//updating list
					m.changeListCh <- true
				}

			}
		}
	}
}

//function to handle question changes and leaderboard

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
		//case to update question
		case _, ok := <-m.updateQuestionCh:
			if !ok {
				log.Println("tried to read from closed updateQuestionCh in questionHandler")
				return
			}
			//checking if game was just started ,if yes set to in progress
			if m.numberOfCurrentQuestion == 1 {
				m.gameState = 2
			}
			//checking if curr number of question equal to questions of the topic
			// if yes than handle end of the game
			if m.numberOfCurrentQuestion == m.numberOfQuestions-1 {

				log.Println("game ended")
				//creating leaderboard map
				var leaderBoard = make(map[string]int)
				m.mu.Lock()
				//looping through the client map to fill leaderboard
				for name, client := range m.clientsMap {
					leaderBoard[name] = client.score
				}
				//also dont forget about players who leaved!
				for name, client := range m.stockMap {
					leaderBoard[name] = client.score
				}

				m.mu.Unlock()
				//filling and sorting
				players := make([]types.Player, 0)
				for name, points := range leaderBoard {
					players = append(players, types.Player{Username: name, Score: points})
				}
				sort.Slice(players, func(i, j int) bool {
					return players[i].Score > players[j].Score

				})
				//writing to writer to end session
				m.loadLeaderBoardCh <- players
			} else {
				//changing question and sending message to channel that everyone is ready so new question can be delivered
				m.numberOfCurrentQuestion++
				m.currentQuestion = m.questions[m.numberOfCurrentQuestion]
				log.Println("current question : ", m.currentQuestion)
				m.changeQuestionCh <- true
			}

		}
	}
}

//function to hande start of the game and list of players who is connected

func (m *Manager) StartGame() {
	defer func() {
		log.Println("STARTGAME | GAME STARTED!")
		log.Println("STARTGAME | exited goroutine")
		close(m.startGameCh)
		close(m.beforeGameConnection)
		close(m.beforeGameLeave)
	}()
	for {

		select {
		//case to handle connection
		case <-m.beforeGameConnection:
			m.mu.Lock()
			//updating list of players
			var listOfPlayers []string
			for username := range m.clientsMap {
				listOfPlayers = append(listOfPlayers, username)
			}
			for _, client := range m.clientsMap {
				client.beforeGameWriterCh <- listOfPlayers
			}
			length := len(m.clientsMap)
			m.mu.Unlock()
			//checking if lobby is full
			//if yes than start game and overwrite waitList
			if length == m.maxPlayers {
				m.mu.Lock()
				for _, client := range m.clientsMap {
					m.waitList = append(m.waitList, client.userName)
				}
				m.mu.Unlock()
				m.startGameCh <- true
				m.changeQuestionCh <- true
				m.gameState = 1
				return
			}
			log.Println("addded connection to before game state ,current list : ", listOfPlayers)
		//case to handle client leave
		case <-m.beforeGameLeave:
			//overwrite list
			m.mu.Lock()
			var listOfPlayers []string
			for username := range m.clientsMap {
				listOfPlayers = append(listOfPlayers, username)
			}
			for _, client := range m.clientsMap {
				client.beforeGameWriterCh <- listOfPlayers
			}
			m.mu.Unlock()
			log.Println("deleted connection from before game state ,current list : ", listOfPlayers)
		case <-m.forcedStartOfGame:

		}
	}
}
