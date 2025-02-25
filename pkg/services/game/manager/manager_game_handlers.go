package manager

import (
	"log"
	"sort"

	"github.com/Megidy/k/types"
)

// function to handle scoreness of client
func (m *Manager) ScoreHandler(client *Client, requestData *types.RequestData) {

	//getting current question
	question := m.questions[client.currQuestion]

	//comparing data
	if requestData.Answer == question.CorrectAnswer {
		log.Println("added point for : ", client.userName)
		//updating score
		client.mu.Lock()
		client.score++
		client.mu.Unlock()
		//updating leaderboard
		m.mu.Lock()
		m.leaderBoard[client.userName]++
		m.mu.Unlock()

	}
	//updating inner leaderboard only if playstyle is spectator
	if m.Owner.PlayStyle == 2 {
		m.updateInnerLeaderboardCh <- true
	}

}

// function to handle status of clients such as :
// 1 readiness
// 2 connection
func (m *Manager) ClientsStatusHandler() {
	defer func() {
		log.Println("CLIENTSSTATUSHANDLER | exited goroutine")
		close(m.removeClientFromWaitList)
	}()
	for {
		select {
		case <-m.ctx.Done():
			return
		//case to handle readiness of client
		case client, ok := <-m.readyCh:
			if !ok {
				log.Println("tried to read from closed readyCh in ClientsStatusHandler")
				return
			}
			//setting client ready status to true
			client.mu.Lock()
			client.isReady = true
			client.mu.Unlock()
			//removing him from waitList
			m.removeClientFromWaitList <- client
		//case to handle connection
		case client, ok := <-m.clientsConnectionCh:
			if !ok {
				log.Println("tried to read from closed clientsConnectionCh in ClientsStatusHandler")
				return
			}
			log.Println("started goroutines for : ", client.userName)

			//starting r/w goroutines
			if client.userName == m.Owner.Username && m.Owner.PlayStyle == 2 {
				go client.ReadPump()
				go client.SpectatorsWritePump()
			} else {
				go client.ReadPump()
				go client.WritePump()
			}

		}

	}
}

// function to handle list update
func (m *Manager) WaitListHandler() {
	defer func() {
		log.Println("WAITLISTHANDLER | exited goroutine")
		close(m.writeListCh)
		close(m.updateQuestionCh)
	}()
	for {
		select {
		case <-m.ctx.Done():
			return
		//case to add client to wait list after connection
		case client, ok := <-m.addClientToWaitList:
			if !ok {
				log.Println("tried to read from closed addClientToWaitList in WaitListHandler")
				return
			}
			//!!
			//should be tested , maybe this also needs with mutex
			m.waitList = append(m.waitList, client.userName)
			m.writeListCh <- true
		//case to remove client from wait list
		case client, ok := <-m.removeClientFromWaitList:
			if !ok {
				log.Println("tried to read from closed removeClientFromWaitList in WaitListHandler")
				return
			}

			//creating temporary waitList
			newWaitList := []string{}

			//looping through the waitList to overwrite him
			for _, value := range m.waitList {
				if value != client.userName {
					newWaitList = append(newWaitList, value)
				}
			}
			//set the waitList as a temp one
			m.waitList = newWaitList
			m.mu.RLock()
			length := len(m.clientsMap)
			m.mu.RUnlock()
			// puprose of this check
			// len(m.waitList) == 0 is to check if waitList is empty -> everyone is ready than update question
			// length != 0 is to check if there are connected clients left, this check is important one , because for example:
			// there are 1 player left and he leaves, len(m.waitList) == 0 so this will go the next question
			if len(m.waitList) == 0 && length != 0 {
				//checking if game is started
				if m.gameState != 0 {
					m.mu.Lock()
					//looping througn the client map to fill the waitList
					for username := range m.clientsMap {
						m.waitList = append(m.waitList, username)
					}
					m.mu.Unlock()
					//updating question
					m.updateQuestionCh <- true
				}

			} else {
				//checking if game is started
				if m.gameState != 0 {
					//updating list
					m.writeListCh <- true
				}

			}
		}
	}
}

// function to handle question changes and leaderboard
func (m *Manager) QuestionHandler() {
	defer func() {
		log.Println("QUESTIONHANDLER | exited goroutine")
		close(m.writeQuestionCh)
		close(m.writeLeaderBoardCh)
		close(m.writeInnerLeaderboardCh)
		close(m.restartTimeCh)
	}()
	for {
		select {
		case <-m.ctx.Done():
			return
		//case to update question
		case _, ok := <-m.updateQuestionCh:
			if !ok {
				log.Println("tried to read from closed updateQuestionCh in questionHandler")
				return
			}
			//checking if game was just started ,if yes set to in progress
			if m.numberOfCurrentQuestion == 1 {
				m.gameState = 2
			}
			//checking if curr number of question equal to questions of the topic
			// if yes than handle end of the game
			if m.numberOfCurrentQuestion == m.numberOfQuestions-1 {

				log.Println("game ended")
				//creating leaderboard map
				var leaderBoard = make(map[string]int)
				m.mu.RLock()
				//looping through the client map to fill leaderboard
				for name, client := range m.clientsMap {
					leaderBoard[name] = client.score
				}
				//also dont forget about players who leaved!
				for name, client := range m.stockMap {
					leaderBoard[name] = client.score
				}
				m.mu.RUnlock()
				m.gameState = -1
				//filling and sorting
				players := make([]types.Player, 0)
				for name, points := range leaderBoard {
					players = append(players, types.Player{Username: name, Score: points})
				}
				sort.Slice(players, func(i, j int) bool {
					return players[i].Score > players[j].Score

				})
				//writing to writer to end session
				log.Println("writing leaderboard ")
				m.writeLeaderBoardCh <- players
			} else {
				//changing question and sending message to channel that everyone is ready so new question can be delivered
				m.numberOfCurrentQuestion++
				m.currentQuestion = m.questions[m.numberOfCurrentQuestion]
				log.Println("current question : ", m.currentQuestion)
				if m.currTime > 0 {
					m.restartTimeCh <- true
				}
				m.writeQuestionCh <- true
				if m.Owner.PlayStyle == 2 && m.Owner.Client.isOnline {
					m.overwriteListCh <- m.Owner.Username
				}

			}
		case <-m.updateInnerLeaderboardCh:
			players := make([]types.Player, 0)
			m.mu.RLock()
			for name, points := range m.leaderBoard {
				players = append(players, types.Player{Username: name, Score: points})
			}
			sort.Slice(players, func(i, j int) bool {
				return players[i].Score > players[j].Score

			})
			m.mu.RUnlock()
			m.writeInnerLeaderboardCh <- players
		}
	}
}
