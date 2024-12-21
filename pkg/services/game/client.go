package game

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Megidy/k/static/components"
	"github.com/Megidy/k/types"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 5 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 1 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type Client struct {
	ginCtx   *gin.Context
	userName string
	conn     *websocket.Conn
	manager  *Manager
	question types.Question
	exitCh   chan struct{}
}

func NewClient(userName string, conn *websocket.Conn, manager *Manager, ctx *gin.Context) *Client {

	return &Client{
		ginCtx:   ctx,
		userName: userName,
		conn:     conn,
		manager:  manager,
		exitCh:   make(chan struct{}),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		log.Println("READ PUMP : exited readpump goroutine of client: ", c.userName)
		close(c.exitCh)
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
		log.Println("data : ", data)
		//score of points
		c.manager.mu.Lock()
		log.Println("state of client ", c.userName, " before submit :", c.manager.done)
		log.Println("client ", c.userName, ", is ready")
		c.manager.done[c] = true
		log.Println("state of client ", c.userName, " after submit :", c.manager.done)
		c.manager.mu.Unlock()
	}

}

func (c *Client) WritePump() {
	defer func() {
		log.Println("WRITE PUMP : exited writepump goroutine of client: ", c.userName)

	}()
	for {
		select {
		case <-c.manager.ctx.Done():
			c.conn.Close()
			return
		case q, ok := <-c.manager.broadcast:
			if !ok {
				log.Println("channel closed while writing to user")
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
					log.Println("error when writing message: ", err)

				}
				continue
			}

			c.question = q
			log.Println("currenct question : ", q)
			comp := components.Question(q)
			buffer := &bytes.Buffer{}
			comp.Render(context.Background(), buffer)

			err := c.conn.WriteMessage(websocket.TextMessage, buffer.Bytes())
			if err != nil {
				log.Println("error when writing message: ", err)

			}

		}

	}

}
