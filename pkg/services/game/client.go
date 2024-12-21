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
	pongWait = 30 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (30 * time.Second * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type Client struct {
	ctx      *gin.Context
	userName string
	conn     *websocket.Conn
	manager  *Manager
	question types.Question
	exitCh   chan bool
}

func NewClient(userName string, conn *websocket.Conn, manager *Manager, c *gin.Context) *Client {

	return &Client{
		ctx:      c,
		userName: userName,
		conn:     conn,
		manager:  manager,
		exitCh:   make(chan bool),
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.exitCh <- true
		log.Println("exited readpump goroutine of client: ", c.userName)
		c.conn.Close()
		close(c.exitCh)
		c.manager.DeleteClientFromConnectionPool(c)
	}()
	// c.conn.SetReadLimit(maxMessageSize)
	// c.conn.SetReadDeadline(time.Now().Add(pongWait))
	// c.conn.SetPongHandler(func(string) error {
	// 	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	// 	log.Println("error when setting pong handler with client : ", c.ID)
	// 	return nil
	// })

	for {
		_, txt, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("error while reading data : ", txt)
			break
		}
		var data interface{}
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
		log.Println("exited writepump goroutine of client: ", c.userName)

	}()
	for {
		select {
		case q, ok := <-c.manager.broadcast:
			if q.Id == c.question.Id {
				continue
			}
			if !ok {
				log.Println("channel closed while writing to user")
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
		case <-c.exitCh:
			return
		}

	}

}

// for {
// 	q := types.Question{
// 		Question:      "how old am I?",
// 		Answers:       []string{"16", "17", "18", "19"},
// 		CorrectAnswer: "18",
// 	}

//		log.Println("readed message : ", q.Question)
//		time.Sleep(time.Second * 2)
//	}
// func (c *Client) GetDataFromForm(ctx *gin.Context, key string) string {
// 	return c.Request.PostFormValue(key)
// }
