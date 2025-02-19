package manager

import (
	"log"

	"github.com/Megidy/k/types"
	"github.com/gin-gonic/gin"
)

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
	GlobalRoomManager.AddConnectionToList(m, client.userName)

	//starting w/r pumps
	m.clientsConnectionCh <- client

	// channel for waiting start of the game
	<-m.startGameCh

	m.WriteDataWithConnection(client, wasInGameBefore)
}

// function to correctly write data to different client
func (m *Manager) WriteDataWithConnection(client *Client, wasInGameBefore bool) {
	//purpose of this check : handle updates in render ONLY if game is started, because it could occur error
	if m.gameState != 0 {
		//checking if player was in game before
		//if he was in game before than checking if he is already answered question and render for him waitList, if not than render question
		//if he wasnt in game before than writing him questions and also adding to waitList to update for all plyers
		if client.userName == m.Owner.Username && m.Owner.PlayStyle == 2 {
			if m.Owner.Client.isOnline {
				m.overwriteQuestionCh <- m.Owner.Username
				m.overwriteListCh <- m.Owner.Username
				if m.gameState != 0 {
					m.overwriteTimeCh <- m.Owner.Client
				}
				m.updateInnerLeaderboardCh <- true
			}
		} else {
			if wasInGameBefore {
				//checking if client is online to prevent writing to not established connections
				if client.isOnline {
					if client.isReady {
						m.overwriteListCh <- client.userName
						if m.gameState != 0 {
							m.overwriteTimeCh <- client
						}
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
