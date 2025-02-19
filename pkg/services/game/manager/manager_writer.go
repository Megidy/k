package manager

import (
	"log"

	"github.com/Megidy/k/worker"
)

// function to handle writing to players
func (m *Manager) Writer() {
	defer func() {
		log.Println("MESSAGE QUEUE | exited goroutine")
		close(m.overwriteListCh)
		close(m.overwriteQuestionCh)
		close(m.overwriteTimeCh)
		close(m.readyCh)
		close(m.clientsConnectionCh)
		close(m.addClientToWaitList)
		close(m.updateInnerLeaderboardCh)
		GlobalRoomManager.EndRoomSession(m.roomID)
	}()
	for {
		select {
		//change question for all connected clients
		case _, ok := <-m.writeQuestionCh:
			if !ok {
				log.Println("tried to read from closed changeQuestionCh in MessageQueue")
				return
			}
			//checking if Owner is spectator and is online to render for him
			if m.Owner.PlayStyle == 2 && m.Owner.Client.isOnline {
				m.Owner.Client.questionCh <- m.currentQuestion
			}

			//!!
			//TEST WITH MUTEXES , IF IT WILL OCCUR ERRORS THAN CHECK PREVIOUS COMMITS TO REVERT CHANGES
			//!!
			//writing the question for all players
			m.mu.RLock()
			for _, client := range m.clientsMap {
				client.mu.Lock()
				client.currQuestion = m.numberOfCurrentQuestion
				client.isReady = false
				client.mu.Unlock()
				client.questionCh <- m.currentQuestion
			}
			m.mu.RUnlock()
		//write curr question for new client
		case _, ok := <-m.writeListCh:
			if !ok {
				log.Println("tried to read from closed changeListCh in MessageQueue")
				return
			}

			//checking if Owner is spectator and is online to render for him
			if m.Owner.PlayStyle == 2 && m.Owner.Client.isOnline {
				m.Owner.Client.writeWaitCh <- m.waitList
			}
			//writing wait list for all players who is not ready
			m.mu.RLock()
			for _, client := range m.clientsMap {
				client.mu.Lock()
				if client.isReady {
					client.writeWaitCh <- m.waitList
				}
				client.mu.Unlock()
			}
			m.mu.RUnlock()
		//write curr questine for new client
		case username, ok := <-m.overwriteQuestionCh:
			if !ok {
				log.Println("tried to read from closed overwriteQuestionCh in MessageQueue")
				return
			}
			//checking if username is owners username nad he is online to overwrite question
			if username == m.Owner.Username && m.Owner.PlayStyle == 2 {
				m.Owner.Client.mu.Lock()
				if m.Owner.Client.isOnline {
					m.Owner.Client.questionCh <- m.currentQuestion
				}
				m.Owner.Client.mu.Unlock()
			} else {
				//getting user
				m.mu.RLock()
				client := m.clientsMap[username]
				m.mu.RUnlock()
				//writing question for player who connected and is ready
				client.mu.Lock()
				client.isReady = false
				client.currQuestion = m.numberOfCurrentQuestion
				client.mu.Unlock()
				client.questionCh <- m.currentQuestion
			}

		//write list of not done clients if client is answered question
		case username, ok := <-m.overwriteListCh:
			if !ok {
				log.Println("tried to read from closed overwriteListCh in MessageQueue")
				return
			}
			//checking if username is owners username nad he is online to overwrite question
			if username == m.Owner.Username && m.Owner.PlayStyle == 2 {
				m.Owner.Client.mu.Lock()
				if m.Owner.Client.isOnline {
					m.Owner.Client.writeWaitCh <- m.waitList
				}
				m.Owner.Client.mu.Unlock()
			} else {
				//writin waitList to new connected player
				m.mu.RLock()
				client := m.clientsMap[username]
				m.mu.RUnlock()
				client.writeWaitCh <- m.waitList
			}

		//write leaderboard to all connected players
		case players, ok := <-m.writeLeaderBoardCh:
			if !ok {
				log.Println("tried to read from closed loadLeaderBoardCh in MessageQueue")
				return
			}

			//checking if Owner is spectator and is online to render for him
			if m.Owner.PlayStyle == 2 && m.Owner.Client.isOnline {
				m.Owner.Client.leaderBoardCh <- players
			}
			//creating task to start tracking leaderboard for adding to the account
			task := worker.NewWorkerTask(m.currentQuestion.Topic.Name, players)
			m.workerPool.StartTrackingGame(task)

			//writing leaderboard to all connected players
			m.mu.RLock()
			for _, client := range m.clientsMap {
				client.leaderBoardCh <- players
			}
			m.mu.RUnlock()
			//calling context cancle func which will end end all goroutines and which will close all channels
			m.cancel()
			return

		//case to handle update for spectator
		case players := <-m.writeInnerLeaderboardCh:
			//updating inner leaderboard for render
			//checking if Owner is spectator and is online to render for him
			if m.Owner.PlayStyle == 2 && m.Owner.Client.isOnline {
				m.Owner.Client.innerLeaderBoardCh <- players
			}
		//case to handle update for time
		case <-m.writeTimeCh:
			//checking if Owner is spectator and is online to render for him
			if m.Owner.PlayStyle == 2 && m.Owner.Client.isOnline {
				m.Owner.Client.timeWriterCh <- m.currTime
			}
			//updating for all clients
			m.mu.RLock()
			for _, client := range m.clientsMap {
				client.timeWriterCh <- m.currTime
			}
			m.mu.RUnlock()
		//case to overwrite time for connected client
		case client := <-m.overwriteTimeCh:
			//checking if username is owners username nad he is online to overwrite question
			if client.userName == m.Owner.Username && m.Owner.PlayStyle == 2 {
				m.Owner.Client.mu.Lock()
				if client.isOnline {
					m.Owner.Client.timeWriterCh <- m.currTime
				}
				m.Owner.Client.mu.Unlock()
			} else {
				client.timeWriterCh <- m.currTime

			}
		}
	}
}
