package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"log"

	"github.com/Megidy/k/static/templates/components"
	"github.com/Megidy/k/types"
	"github.com/gorilla/websocket"
)

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
		close(c.innerLeaderBoardCh)
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
		log.Println("WRITE PUMP : exited writepump goroutine for Owner : ", c.userName)
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
