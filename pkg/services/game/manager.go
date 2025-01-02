package game

import (
	"context"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/Megidy/k/types"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type ownerStruct struct {
	username string
	client   *Client
	//playsyle of the owner:
	// 1 - player
	// 2 - spectator
	playStyle int
}

// !!
// TEST WITH MUTEXES , IF IT WILL OCCUR ERRORS THAN CHECK PREVIOUS COMMITS TO REVERT CHANGES
// !!

// TO DO
// 1 create timer which will later be dependent on the time of question || HALF DONE
// 2 if 0 connection and game is not started and noone connects for 120 seconds , than delete room
// 3 make notifications for connection and disconnection , but this will have maybe some frontend issues ?
type Manager struct {
	//owner
	owner ownerStruct
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
	mu sync.RWMutex
	//context
	ctx context.Context
	//cancel func of context
	cancel context.CancelFunc
	//number of current question
	numberOfCurrentQuestion int
	//current time of question
	currTime int
	//questions
	questions []types.Question
	//curr question
	currentQuestion types.Question
	//clientsMap
	clientsMap map[string]*Client
	//stockMap for not repition of clients
	stockMap map[string]*Client
	//leaderboard
	leaderBoard map[string]int
	//change question chan
	writeQuestionCh chan bool
	//change list for all clients
	writeListCh chan bool
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
	writeLeaderBoardCh chan []types.Player
	//channels to update curr players before the game
	beforeGameConnection chan bool
	beforeGameLeave      chan bool
	//channel to start gmae before all connnections are established
	forcedStartOfGame chan bool
	//channel to update time
	writeTimeCh chan bool
	//channel to update time for clients who just connected
	overwriteTimeCh chan *Client
	//channel to restart timer in case if all clients is ready
	restartTimeCh chan bool
	//channel to update real-time leaderboard for spectator
	updateInnerLeaderboardCh chan bool
	//channel to write real-time leaderboard for specator
	writeInnerLeaderboardCh chan []types.Player
}

// constructor
func NewManager(owner, roomID string, playstyle, numberOfPlayers, amountOfQuestions int, questions []types.Question) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	manager := &Manager{
		owner:                    ownerStruct{username: owner, playStyle: playstyle},
		roomID:                   roomID,
		maxPlayers:               numberOfPlayers,
		numberOfQuestions:        amountOfQuestions,
		currTime:                 10,
		questions:                questions,
		mu:                       sync.RWMutex{},
		ctx:                      ctx,
		cancel:                   cancel,
		currentQuestion:          questions[0],
		clientsMap:               make(map[string]*Client),
		stockMap:                 make(map[string]*Client),
		leaderBoard:              make(map[string]int),
		writeQuestionCh:          make(chan bool),
		writeListCh:              make(chan bool),
		overwriteQuestionCh:      make(chan string),
		overwriteListCh:          make(chan string),
		readyCh:                  make(chan *Client),
		startGameCh:              make(chan bool),
		clientsConnectionCh:      make(chan *Client),
		addClientToWaitList:      make(chan *Client),
		removeClientFromWaitList: make(chan *Client),
		updateQuestionCh:         make(chan bool),
		writeLeaderBoardCh:       make(chan []types.Player),
		beforeGameConnection:     make(chan bool),
		beforeGameLeave:          make(chan bool),
		forcedStartOfGame:        make(chan bool),
		writeTimeCh:              make(chan bool),
		overwriteTimeCh:          make(chan *Client),
		restartTimeCh:            make(chan bool),
		updateInnerLeaderboardCh: make(chan bool),
		writeInnerLeaderboardCh:  make(chan []types.Player),
	}
	go manager.Writer()
	go manager.ClientsStatusHandler()
	go manager.QuestionHandler()
	go manager.WaitListHandler()
	go manager.StartGame()
	log.Println("manager : ", manager)
	return manager
}

// function to handle scoreness of client
func (m *Manager) ScoreHandler(client *Client, requestData *types.RequestData) {

	//getting current question
	question := m.questions[client.currQuestion]

	//comparing data
	if requestData.Answer == question.CorrectAnswer {
		log.Println("added point for : ", client.userName)
		//updating score
		client.mu.Lock()
		client.score++
		client.mu.Unlock()
		//updating leaderboard
		m.mu.Lock()
		m.leaderBoard[client.userName]++
		m.mu.Unlock()

	}
	//updating inner leaderboard only if playstyle is spectator
	if m.owner.playStyle == 2 {
		m.updateInnerLeaderboardCh <- true
	}

}

// function which handles connection of clients
func (m *Manager) AddClientToConnectionPool(client *Client) bool {
	//varriable to check if client was in game before(was in stock map)
	var wasInGameBefore bool
	//checking if client is owner

	if client.userName == m.owner.username && m.owner.playStyle == 2 {
		client.mu.Lock()
		//updating connection
		m.owner.client = client
		//setting online as true to prevent issues with writing
		client.isOnline = true
		client.currQuestion = m.numberOfCurrentQuestion
		client.mu.Unlock()
		log.Println("Added owner to connection pool : ", client.userName)
		log.Println("owners online status : ", m.owner.client.isOnline)
		//updating gameConnection for owner , !not writing spectator to them!
		if m.gameState == 0 {
			m.beforeGameConnection <- true
		}
	} else {

		//checking if client was connected before
		m.mu.RLock()
		c, ok1 := m.stockMap[client.userName]
		m.mu.RUnlock()
		if ok1 {
			//checking if he is joining to the game on the same question on which he leaved ,
			//if yes than handle isReady field,otherwise just set false , because he connected to new question
			c.mu.Lock()
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
			c.mu.Unlock()

			//than deleting client from stockMap , because he connected
			m.mu.Lock()
			delete(m.stockMap, c.userName)
			m.mu.Unlock()
			//and setting wasInGameBefore = true becuase he appeared in stockMap before
			wasInGameBefore = true
		} else {
			wasInGameBefore = false
		}
		//setting client as online ,because he connected
		client.mu.Lock()
		client.isOnline = true
		client.mu.Unlock()
		//than adding this user to map
		m.mu.Lock()
		m.clientsMap[client.userName] = client
		m.mu.Unlock()
		log.Println("Added new client to connection pool : ", client.userName)
		// this checking of game state was made for writing to beforeGameCh for each client, so if gameState=0 ,than game didnt started yet
		// so this updates list of players who is connected now
		if m.gameState == 0 {
			m.beforeGameConnection <- true
		}
	}

	return wasInGameBefore
}

// function to delete client from connection pool and from globalRoomManager
func (m *Manager) DeleteClientFromConnectionPool(client *Client) {
	//checking if deleting owner
	if client.userName == m.owner.username && m.owner.playStyle == 2 {
		log.Println("deleted owner : ", client.userName)
		m.owner.client.mu.Lock()
		m.owner.client.isOnline = false
		m.owner.client.mu.Unlock()
		globalRoomManager.DeleteConnectionFromList(m, client.userName)
		log.Println("owners online status : ", m.owner.client.isOnline)

	} else {

		//removing client from waitList to update him for other players
		m.removeClientFromWaitList <- client
		//setting online status as false ,because client leaved
		client.mu.Lock()
		client.isOnline = false
		client.mu.Unlock()
		//deleting client from main map and adding to stash
		m.mu.Lock()
		m.stockMap[client.userName] = client
		delete(m.clientsMap, client.userName)
		m.mu.Unlock()
		log.Println("deleted from clientsMap username : ", client.userName)
		//deleteting from globalRoomManager
		globalRoomManager.DeleteConnectionFromList(m, client.userName)
		//updating beforeGame list of players
		if m.gameState == 0 {
			m.beforeGameLeave <- true
		}
	}

}

// function to correctly write data to different client
func (m *Manager) WriteDataWithConnection(client *Client, wasInGameBefore bool) {
	//purpose of this check : handle updates in render ONLY if game is started, because it could occur error
	if m.gameState != 0 {
		//checking if player was in game before
		//if he was in game before than checking if he is already answered question and render for him waitList, if not than render question
		//if he wasnt in game before than writing him questions and also adding to waitList to update for all plyers
		if client.userName == m.owner.username && m.owner.playStyle == 2 {
			if m.owner.client.isOnline {
				m.overwriteQuestionCh <- m.owner.username
				m.overwriteListCh <- m.owner.username
				if m.gameState != 0 {
					m.overwriteTimeCh <- m.owner.client
				}
				m.updateInnerLeaderboardCh <- true
			}
		} else {
			if wasInGameBefore {
				//checking if client is online to prevent writing to not established connections
				if client.isOnline {
					if client.isReady {
						m.overwriteListCh <- client.userName
					} else {
						m.addClientToWaitList <- client
						m.overwriteQuestionCh <- client.userName
						if m.gameState != 0 {
							m.overwriteTimeCh <- client
						}
					}
					log.Println("added new client who was in game before")
				}

			} else {
				//checking if client is online to prevent writing to not established connections
				if client.isOnline {
					m.addClientToWaitList <- client
					m.overwriteQuestionCh <- client.userName
					if m.gameState != 0 {
						m.overwriteTimeCh <- client
					}
					log.Println("added new client")
					log.Println("started goroutine for new client : ", client.userName)
				}

			}
		}

	}
}

// Function to handle Websocket connection and handle creating of client
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
	client := NewClient(user.UserName, wsConn, m)

	//adding him to connection pool, meanwhile cheking if he was connected before
	wasInGameBefore := m.AddClientToConnectionPool(client)
	globalRoomManager.AddConnectionToList(m, client.userName)

	//starting w/r pumps
	m.clientsConnectionCh <- client

	// channel for waiting start of the game
	<-m.startGameCh

	m.WriteDataWithConnection(client, wasInGameBefore)
}

// function to handle writing to players
func (m *Manager) Writer() {
	defer func() {
		log.Println("MESSAGE QUEUE | exited goroutine")
		close(m.overwriteListCh)
		close(m.overwriteQuestionCh)
		close(m.overwriteTimeCh)
		close(m.readyCh)
		close(m.clientsConnectionCh)
		close(m.addClientToWaitList)
		close(m.updateInnerLeaderboardCh)
		globalRoomManager.EndRoomSession(m.roomID)
	}()
	for {
		select {
		//change question for all connected clients
		case _, ok := <-m.writeQuestionCh:
			if !ok {
				log.Println("tried to read from closed changeQuestionCh in MessageQueue")
				return
			}
			//checking if owner is spectator and is online to render for him
			if m.owner.playStyle == 2 && m.owner.client.isOnline {
				m.owner.client.questionCh <- m.currentQuestion
			}

			//!!
			//TEST WITH MUTEXES , IF IT WILL OCCUR ERRORS THAN CHECK PREVIOUS COMMITS TO REVERT CHANGES
			//!!
			//writing the question for all players
			m.mu.RLock()
			for _, client := range m.clientsMap {
				client.mu.Lock()
				client.currQuestion = m.numberOfCurrentQuestion
				client.isReady = false
				client.mu.Unlock()
				client.questionCh <- m.currentQuestion
			}
			m.mu.RUnlock()
		//write curr question for new client
		case _, ok := <-m.writeListCh:
			if !ok {
				log.Println("tried to read from closed changeListCh in MessageQueue")
				return
			}

			//checking if owner is spectator and is online to render for him
			if m.owner.playStyle == 2 && m.owner.client.isOnline {
				m.owner.client.writeWaitCh <- m.waitList
			}
			//writing wait list for all players who is not ready
			m.mu.RLock()
			for _, client := range m.clientsMap {
				client.mu.Lock()
				if client.isReady {
					client.writeWaitCh <- m.waitList
				}
				client.mu.Unlock()
			}
			m.mu.RUnlock()
		//write curr questine for new client
		case username, ok := <-m.overwriteQuestionCh:
			if !ok {
				log.Println("tried to read from closed overwriteQuestionCh in MessageQueue")
				return
			}
			//checking if username is owners username nad he is online to overwrite question
			if username == m.owner.username && m.owner.playStyle == 2 {
				m.owner.client.mu.Lock()
				if m.owner.client.isOnline {
					m.owner.client.questionCh <- m.currentQuestion
				}
				m.owner.client.mu.Unlock()
			} else {
				//getting user
				m.mu.RLock()
				client := m.clientsMap[username]
				m.mu.RUnlock()
				//writing question for player who connected and is ready
				client.mu.Lock()
				client.isReady = false
				client.currQuestion = m.numberOfCurrentQuestion
				client.mu.Unlock()
				client.questionCh <- m.currentQuestion
			}

		//write list of not done clients if client is answered question
		case username, ok := <-m.overwriteListCh:
			if !ok {
				log.Println("tried to read from closed overwriteListCh in MessageQueue")
				return
			}
			//checking if username is owners username nad he is online to overwrite question
			if username == m.owner.username && m.owner.playStyle == 2 {
				m.owner.client.mu.Lock()
				if m.owner.client.isOnline {
					m.owner.client.writeWaitCh <- m.waitList
				}
				m.owner.client.mu.Unlock()
			} else {
				//writin waitList to new connected player
				m.mu.RLock()
				client := m.clientsMap[username]
				m.mu.RUnlock()
				client.writeWaitCh <- m.waitList
			}

		//write leaderboard to all connected players
		case players, ok := <-m.writeLeaderBoardCh:
			if !ok {
				log.Println("tried to read from closed loadLeaderBoardCh in MessageQueue")
				return
			}

			//checking if owner is spectator and is online to render for him
			if m.owner.playStyle == 2 && m.owner.client.isOnline {
				m.owner.client.leaderBoardCh <- players
			}
			//writing leaderboard to all connected players
			m.mu.RLock()
			for _, client := range m.clientsMap {
				client.leaderBoardCh <- players
			}
			m.mu.RUnlock()
			//calling context cancle func which will end end all goroutines and which will close all channels
			m.cancel()
			return

		//case to handle update for spectator
		case players := <-m.writeInnerLeaderboardCh:
			//updating inner leaderboard for render
			//checking if owner is spectator and is online to render for him
			if m.owner.playStyle == 2 && m.owner.client.isOnline {
				m.owner.client.innerLeaderBoardCh <- players
			}
		//case to handle update for time
		case <-m.writeTimeCh:
			//checking if owner is spectator and is online to render for him
			if m.owner.playStyle == 2 && m.owner.client.isOnline {
				m.owner.client.timeWriterCh <- m.currTime
			}
			//updating for all clients
			m.mu.RLock()
			for _, client := range m.clientsMap {
				client.timeWriterCh <- m.currTime
			}
			m.mu.RUnlock()
		//case to overwrite time for connected client
		case client := <-m.overwriteTimeCh:
			//checking if username is owners username nad he is online to overwrite question
			if client.userName == m.owner.username && m.owner.playStyle == 2 {
				m.owner.client.mu.Lock()
				if client.isOnline {
					m.owner.client.timeWriterCh <- m.currTime
				}
				m.owner.client.mu.Unlock()
			} else {
				client.timeWriterCh <- m.currTime

			}
		}
	}
}

// function to handle status of clients such as :
// 1 readiness
// 2 connection
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
			client.mu.Lock()
			client.isReady = true
			client.mu.Unlock()
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
			if client.userName == m.owner.username && m.owner.playStyle == 2 {
				go client.ReadPump()
				go client.SpectatorsWritePump()
			} else {
				go client.ReadPump()
				go client.WritePump()
			}

		}

	}
}

// function to handle list update
func (m *Manager) WaitListHandler() {
	defer func() {
		log.Println("WAITLISTHANDLER | exited goroutine")
		close(m.writeListCh)
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
			//!!
			//should be tested , maybe this also needs with mutex
			m.waitList = append(m.waitList, client.userName)
			m.writeListCh <- true
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
			m.mu.RLock()
			length := len(m.clientsMap)
			m.mu.RUnlock()
			// puprose of this check
			// len(m.waitList) == 0 is to check if waitList is empty -> everyone is ready than update question
			// length != 0 is to check if there are connected clients left, this check is important one , because for example:
			// there are 1 player left and he leaves, len(m.waitList) == 0 so this will go the next question
			if len(m.waitList) == 0 && length != 0 {
				//checking if game is started
				if m.gameState != 0 {
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
				if m.gameState != 0 {
					//updating list
					m.writeListCh <- true
				}

			}
		}
	}
}

// function to handle question changes and leaderboard
func (m *Manager) QuestionHandler() {
	defer func() {
		log.Println("QUESTIONHANDLER | exited goroutine")
		close(m.writeQuestionCh)
		close(m.writeLeaderBoardCh)
		close(m.writeInnerLeaderboardCh)
		close(m.restartTimeCh)
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
				m.mu.RLock()
				//looping through the client map to fill leaderboard
				for name, client := range m.clientsMap {
					leaderBoard[name] = client.score
				}
				//also dont forget about players who leaved!
				for name, client := range m.stockMap {
					leaderBoard[name] = client.score
				}
				m.mu.RUnlock()
				m.gameState = -1
				//filling and sorting
				players := make([]types.Player, 0)
				for name, points := range leaderBoard {
					players = append(players, types.Player{Username: name, Score: points})
				}
				sort.Slice(players, func(i, j int) bool {
					return players[i].Score > players[j].Score

				})
				//writing to writer to end session
				log.Println("writing leaderboard ")
				m.writeLeaderBoardCh <- players
			} else {
				//changing question and sending message to channel that everyone is ready so new question can be delivered
				m.numberOfCurrentQuestion++
				m.currentQuestion = m.questions[m.numberOfCurrentQuestion]
				log.Println("current question : ", m.currentQuestion)
				if m.currTime > 0 {
					m.restartTimeCh <- true
				}
				m.writeQuestionCh <- true
				if m.owner.playStyle == 2 && m.owner.client.isOnline {
					m.overwriteListCh <- m.owner.username
				}

			}
		case <-m.updateInnerLeaderboardCh:
			players := make([]types.Player, 0)
			m.mu.RLock()
			for name, points := range m.leaderBoard {
				players = append(players, types.Player{Username: name, Score: points})
			}
			sort.Slice(players, func(i, j int) bool {
				return players[i].Score > players[j].Score

			})
			m.mu.RUnlock()
			m.writeInnerLeaderboardCh <- players
		}
	}
}

// function to hande start of the game and list of players who is connected
func (m *Manager) StartGame() {
	defer func() {
		log.Println("STARTGAME | GAME STARTED!")
		log.Println("STARTGAME | exited goroutine")
		close(m.startGameCh)
		close(m.beforeGameConnection)
		close(m.beforeGameLeave)
		close(m.forcedStartOfGame)
	}()
	for {
		select {
		//case to handle connection
		case <-m.beforeGameConnection:
			m.mu.RLock()
			//updating list of players
			var listOfPlayers []string
			for username := range m.clientsMap {
				listOfPlayers = append(listOfPlayers, username)
			}
			if m.owner.playStyle == 2 {
				if m.owner.client.isOnline {
					m.owner.client.beforeGameWriterCh <- listOfPlayers
				}
			}
			for _, client := range m.clientsMap {
				client.beforeGameWriterCh <- listOfPlayers
			}
			length := len(m.clientsMap)
			m.mu.RUnlock()
			//checking if lobby is full
			if length == m.maxPlayers {
				m.startGameCh <- true
				m.writeQuestionCh <- true
				m.gameState = 1
				go m.TimeHandler()
				return
			}
			log.Println("addded connection to before game state ,current list : ", listOfPlayers)
		//case to handle client leave
		case <-m.beforeGameLeave:
			//overwrite list
			m.mu.RLock()
			var listOfPlayers []string
			for username := range m.clientsMap {
				listOfPlayers = append(listOfPlayers, username)
			}
			if m.owner.playStyle == 2 {
				if m.owner.client.isOnline {
					m.owner.client.beforeGameWriterCh <- listOfPlayers
				}
			}
			for _, client := range m.clientsMap {
				client.beforeGameWriterCh <- listOfPlayers
			}
			m.mu.RUnlock()
			log.Println("deleted connection from before game state ,current list : ", listOfPlayers)
		case <-m.forcedStartOfGame:
			log.Println("force start of game ")
			m.startGameCh <- true
			m.writeQuestionCh <- true
			m.gameState = 1
			go m.TimeHandler()
			return
		}
	}
}

// function to handle time updates
func (m *Manager) TimeHandler() {
	ticker := time.NewTicker(time.Second * 1)
	defer func() {
		ticker.Stop()
		close(m.writeTimeCh)
	}()
	var count int = 10
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			count--
			//checking if time is up
			if count == 0 {
				//updating counter
				count = 10
				m.currTime = count

				var tempWaitList []string
				m.mu.RLock()
				for username := range m.clientsMap {
					tempWaitList = append(tempWaitList, username)
				}
				m.mu.RUnlock()
				m.waitList = tempWaitList
				m.updateQuestionCh <- true
				m.writeTimeCh <- true
			} else {
				m.currTime = count
				log.Println("tick : ", count)
				m.writeTimeCh <- true
			}
		case <-m.restartTimeCh:
			ticker.Reset(time.Second * 1)
			count = 10
			m.currTime = count
			m.writeTimeCh <- true
		}

	}
}
