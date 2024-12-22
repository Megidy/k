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
	ginCtx      *gin.Context
	userName    string
	conn        *websocket.Conn
	manager     *Manager
	question    types.Question
	writeWaitCh chan []string
	lenOfWait   int
}

func NewClient(userName string, conn *websocket.Conn, manager *Manager, ctx *gin.Context) *Client {

	return &Client{
		ginCtx:      ctx,
		userName:    userName,
		conn:        conn,
		manager:     manager,
		writeWaitCh: make(chan []string),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		log.Println("READ PUMP : exited readpump goroutine of client: ", c.userName)
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
		var data types.RequestData

		json.Unmarshal(txt, &data)
		c.manager.HandlePointScoreness(c, data)
		c.manager.mu.Lock()
		c.manager.doneMap[c] = true
		c.manager.mu.Unlock()
		log.Println("READ PUMP : client : ", c.userName, " answered question")
	}

}

func (c *Client) WritePump() {
	defer func() {
		log.Println("WRITE PUMP : exited writepump goroutine of client: ", c.userName)
		close(c.writeWaitCh)
	}()
	for {
		select {
		case <-c.manager.ctx.Done():
			c.conn.Close()
			return
		case q, ok := <-c.manager.broadcast:
			if !ok {
				log.Println("WRITE PUMP : channel closed while writing to user")
				return
			}
			if q.Id == c.question.Id {
				continue
			}
			if q.Id == "ID-leaderBoard" {
				comp := components.LeaderBoard()
				buffer := &bytes.Buffer{}
				comp.Render(context.Background(), buffer)

				err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
				if err != nil {
					log.Println("WRITE PUMP : error when writing message: ", err)
				}
				continue
			}

			c.question = q
			log.Println("current question : ", q)
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
		case list, ok := <-c.writeWaitCh:
			if !ok {
				log.Println("WRITE PUMP : error when reading from writeWaitCh of user , : ", c.userName)
				return
			}
			if len(list) == c.lenOfWait {
				continue
			}
			c.lenOfWait = len(list)

			comp := components.Waiting(list)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)

			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				log.Println("WRITE PUMP : error when writing message: ", err)
			}
		}

	}

}
