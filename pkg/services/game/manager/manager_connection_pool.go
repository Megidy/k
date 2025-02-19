package manager

import "log"

// function which handles connection of clients
func (m *Manager) AddClientToConnectionPool(client *Client) bool {
	//varriable to check if client was in game before(was in stock map)
	var wasInGameBefore bool
	//checking if client is Owner

	if client.userName == m.Owner.Username && m.Owner.PlayStyle == 2 {
		client.mu.Lock()
		//updating connection
		m.Owner.Client = client
		//setting online as true to prevent issues with writing
		client.isOnline = true
		client.currQuestion = m.numberOfCurrentQuestion
		client.mu.Unlock()
		log.Println("Added Owner to connection pool : ", client.userName)
		log.Println("owners online status : ", m.Owner.Client.isOnline)
		//updating gameConnection for Owner , !not writing spectator to them!
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
	//checking if deleting Owner
	if client.userName == m.Owner.Username && m.Owner.PlayStyle == 2 {
		log.Println("deleted Owner : ", client.userName)
		m.Owner.Client.mu.Lock()
		m.Owner.Client.isOnline = false
		m.Owner.Client.mu.Unlock()
		GlobalRoomManager.DeleteConnectionFromList(m, client.userName)
		log.Println("owners online status : ", m.Owner.Client.isOnline)

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
		GlobalRoomManager.DeleteConnectionFromList(m, client.userName)
		//updating beforeGame list of players
		if m.gameState == 0 {
			m.beforeGameLeave <- true
		}
	}

}
