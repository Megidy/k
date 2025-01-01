package game

import (
	"bytes"
	"context"
	"encoding/json"
	"log"

	"github.com/Megidy/k/static/components"
	"github.com/Megidy/k/types"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// const (
// // Time allowed to write a message to the peer.
//  writeWait = 10 * time.Second

// // Time allowed to read the next pong message from the peer.
//  pongWait = 5 * time.Second

// //Send pings to peer with this period. Must be less than pongWait.
// pingPeriod = 1 * time.Second

// // Maximum message size allowed from peer.
//  maxMessageSize = 512
// )

type Client struct {
	//field to prevent error with double connection before game
	isOnline           bool
	isReady            bool
	ginCtx             *gin.Context
	userName           string
	score              int
	conn               *websocket.Conn
	manager            *Manager
	questionCh         chan types.Question
	writeWaitCh        chan []string
	endWriteCh         chan bool
	leaderBoardCh      chan []types.Player
	beforeGameWriterCh chan []string
	timeWriterCh       chan int
	innerLeaderBoardCh chan []types.Player
	currQuestion       int
}

func NewClient(userName string, conn *websocket.Conn, manager *Manager, ctx *gin.Context) *Client {

	return &Client{
		ginCtx:             ctx,
		userName:           userName,
		conn:               conn,
		manager:            manager,
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
		// c.manager.clientsLeaveCh <- c

		// select {
		// case c.endWriteCh <- true:
		// default:
		// 	log.Println("endWriteCh is already closed or not accessible")
		// }
		c.endWriteCh <- true
		close(c.endWriteCh)
		c.manager.DeleteClientFromConnectionPool(c)
	}()
	for {
		_, txt, err := c.conn.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.manager.readyCh <- c
		var data types.RequestData
		json.Unmarshal(txt, &data)
		if data.Answer == "" && c.manager.gameState == 0 {
			c.manager.forcedStartOfGame <- true
		} else {
			c.manager.ScoreHandler(c, &data)
			// c.manager.HandlePointScoreness(c, data)
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
		case list := <-c.writeWaitCh:
			comp := components.Waiting(list)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)

			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
			continue
		case players := <-c.leaderBoardCh:
			comp := components.LeaderBoard(players)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)
			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		case list := <-c.beforeGameWriterCh:
			comp := components.BeforeGameWaitList(list)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)
			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		case time := <-c.timeWriterCh:
			comp := components.TimeLoader(time)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)
			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
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
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		case list := <-c.beforeGameWriterCh:
			comp := components.BeforeGameWaitList(list)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)
			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		case time := <-c.timeWriterCh:
			comp := components.TimeLoader(time)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)
			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		case players := <-c.innerLeaderBoardCh:
			comp := components.SpectatorsLeaderBoard(players)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)
			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		case players := <-c.leaderBoardCh:
			log.Println("client received")
			comp := components.LeaderBoard(players)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)
			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		}
	}
}
