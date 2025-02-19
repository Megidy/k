package manager

import (
	"context"
	"sync"

	"github.com/Megidy/k/config"
	"github.com/Megidy/k/types"
	"github.com/Megidy/k/worker"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type ownerStruct struct {
	Username string
	Client   *Client
	//playsyle of the Owner:
	// 1 - player
	// 2 - spectator
	PlayStyle int
}

// !!
// TEST WITH MUTEXES , IF IT WILL OCCUR ERRORS THAN CHECK PREVIOUS COMMITS TO REVERT CHANGES
// !!

// TO DO
// 1 create timer which will later be dependent on the time of question || HALF DONE
// 2 if 0 connection and game is not started and noone connects for 120 seconds , than delete room ||DONE

type Manager struct {
	//config
	config *config.Config
	//worker pool to handle score caching
	workerPool worker.WorkerManager
	//Owner
	Owner ownerStruct
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
	//channel to update before game reset of ticker to prevent self-liquidation of game
	updateBeforeGameTickerCh chan bool
	//chennel to stop ticker
	stopBeforeGameTickerCh chan bool
}

// constructor
func NewManager(cfg *config.Config, workerPool worker.WorkerManager, Owner, roomID string, playstyle, numberOfPlayers, amountOfQuestions int, questions []types.Question) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	if amountOfQuestions > len(questions) {
		amountOfQuestions = len(questions)
	}
	manager := &Manager{
		config:                   cfg,
		workerPool:               workerPool,
		Owner:                    ownerStruct{Username: Owner, PlayStyle: playstyle},
		roomID:                   roomID,
		maxPlayers:               numberOfPlayers,
		numberOfQuestions:        amountOfQuestions,
		currTime:                 cfg.TimeForAnswer,
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
		updateBeforeGameTickerCh: make(chan bool),
		stopBeforeGameTickerCh:   make(chan bool),
	}

	return manager
}
func (m *Manager) Run() {
	go m.Writer()
	go m.ClientsStatusHandler()
	go m.QuestionHandler()
	go m.WaitListHandler()
	go m.StartGame()
	go m.BeforeGameLiquidationHandler()
}
