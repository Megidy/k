package game

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/Megidy/k/static/components"
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
	//channel to render inner leaderboard for owner of ther game , only used if owner is spectator
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

func (c *Client) ReadPump() {
	defer func() {
		log.Println("READ PUMP : exited readpump goroutine of client: ", c.userName)
		c.endWriteCh <- true
		close(c.endWriteCh)
		c.manager.DeleteClientFromConnectionPool(c)
	}()
	for {
		//reading message
		_, txt, err := c.conn.ReadMessage()
		//checking for errors
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		//pinging to manager that client is ready
		c.manager.readyCh <- c

		var data types.RequestData

		//unmarshaling data
		json.Unmarshal(txt, &data)

		//checking if answer is empty to force start the game
		if data.Answer == "" && c.manager.gameState == 0 {
			c.manager.forcedStartOfGame <- true
		} else {
			//handlign score
			c.manager.ScoreHandler(c, &data)
			log.Println("READ PUMP : client : ", c.userName, " answered question")
		}
	}

}

func (c *Client) WritePump() {
	defer func() {
		log.Println("WRITE PUMP : exited writepump goroutine of client: ", c.userName)
		c.conn.Close()
		close(c.questionCh)
		close(c.writeWaitCh)
		close(c.leaderBoardCh)
		close(c.beforeGameWriterCh)
		close(c.timeWriterCh)
	}()
	for {
		select {
		case <-c.endWriteCh:
			return
		case <-c.manager.ctx.Done():
			return
		//case to render questions
		case q, ok := <-c.questionCh:
			if !ok {
				log.Println("WRITE PUMP : channel closed while writing to user")
				return
			}
			comp := components.Question(q)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)

			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				if err == websocket.ErrCloseSent {
					log.Println("WRITE PUMP : no connection was established ")
					return
				}
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		//case to render waitList
		case list := <-c.writeWaitCh:
			comp := components.Waiting(list)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)

			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				if err == websocket.ErrCloseSent {
					log.Println("WRITE PUMP : no connection was established ")
					return
				}
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		//case to render leaderboard
		case players := <-c.leaderBoardCh:
			comp := components.LeaderBoard(players)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)
			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				if err == websocket.ErrCloseSent {
					log.Println("WRITE PUMP : no connection was established ")
					return
				}
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		//case to render before game connections
		case list := <-c.beforeGameWriterCh:
			comp := components.BeforeGameWaitList(list)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)
			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				if err == websocket.ErrCloseSent {
					log.Println("WRITE PUMP : no connection was established ")
					return
				}
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		//case to render time
		case time := <-c.timeWriterCh:
			comp := components.TimeLoader(time)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)
			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				if err == websocket.ErrCloseSent {
					log.Println("WRITE PUMP : no connection was established ")
					return
				}
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		}
	}
}

func (c *Client) SpectatorsWritePump() {
	defer func() {
		log.Println("WRITE PUMP : exited writepump goroutine for owner : ", c.userName)
		c.conn.Close()
		close(c.questionCh)
		close(c.writeWaitCh)
		close(c.leaderBoardCh)
		close(c.beforeGameWriterCh)
		close(c.timeWriterCh)
		close(c.innerLeaderBoardCh)
	}()
	for {
		select {
		case <-c.endWriteCh:
			return
		case <-c.manager.ctx.Done():
			return
		case question := <-c.questionCh:
			comp := components.SpectatorQuestion(question)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)

			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				if err == websocket.ErrCloseSent {
					log.Println("WRITE PUMP : no connection was established ")
					return
				}
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		case list := <-c.writeWaitCh:
			comp := components.SpectatorWaitList(list)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)

			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				if err == websocket.ErrCloseSent {
					log.Println("WRITE PUMP : no connection was established ")
					return
				}
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		case list := <-c.beforeGameWriterCh:
			comp := components.BeforeGameWaitList(list)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)
			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				if err == websocket.ErrCloseSent {
					log.Println("WRITE PUMP : no connection was established ")
					return
				}
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		case time := <-c.timeWriterCh:
			comp := components.TimeLoader(time)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)
			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				if err == websocket.ErrCloseSent {
					log.Println("WRITE PUMP : no connection was established ")
					return
				}
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		case players := <-c.innerLeaderBoardCh:
			comp := components.SpectatorsLeaderBoard(players)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)
			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				if err == websocket.ErrCloseSent {
					log.Println("WRITE PUMP : no connection was established ")
					return
				}
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		case players := <-c.leaderBoardCh:
			log.Println("client received")
			comp := components.LeaderBoard(players)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)
			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				if err == websocket.ErrCloseSent {
					log.Println("WRITE PUMP : no connection was established ")
					return
				}
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		}
	}
}
