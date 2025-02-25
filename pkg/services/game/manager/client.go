package manager

import (
	"sync"

	"github.com/Megidy/k/types"
	"github.com/gorilla/websocket"
)

type Client struct {
	//field to prevent error with double connection before game
	isOnline bool
	//field to check if client has answered question
	isReady bool
	//username - used as and indentifier 'primary key'
	userName string
	//score of player
	score int
	//mutex for concurrent safe
	mu sync.Mutex
	//websocket connection
	conn *websocket.Conn
	//pointer to the manager to contact with him
	manager *Manager
	//channel to render questions
	questionCh chan types.Question
	//channel to render players who is not ready 'waitList'
	writeWaitCh chan []string
	//channel to close WritePump in case if client leaved
	endWriteCh chan bool
	//channel to render leaderboard
	leaderBoardCh chan []types.Player
	//channel to render before game connetions
	beforeGameWriterCh chan []string
	//channel to render time
	timeWriterCh chan int
	//channel to render inner leaderboard for Owner of ther game , only used if Owner is spectator
	innerLeaderBoardCh chan []types.Player
	//number of currQuestion to maintain correct reconnection
	currQuestion int
}

func NewClient(userName string, conn *websocket.Conn, manager *Manager) *Client {

	return &Client{
		userName:           userName,
		conn:               conn,
		manager:            manager,
		mu:                 sync.Mutex{},
		questionCh:         make(chan types.Question),
		writeWaitCh:        make(chan []string),
		endWriteCh:         make(chan bool),
		leaderBoardCh:      make(chan []types.Player),
		beforeGameWriterCh: make(chan []string),
		timeWriterCh:       make(chan int),
		innerLeaderBoardCh: make(chan []types.Player),
	}
}
