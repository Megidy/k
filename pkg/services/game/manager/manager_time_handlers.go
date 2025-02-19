package manager

import (
	"log"
	"time"
)

// function to hande start of the game and list of players who is connected
func (m *Manager) StartGame() {
	defer func() {
		log.Println("STARTGAME | GAME STARTED!")
		log.Println("STARTGAME | exited goroutine")
		close(m.startGameCh)
		close(m.beforeGameConnection)
		close(m.beforeGameLeave)
		close(m.forcedStartOfGame)
	}()
	for {
		select {
		//case to handle connection
		case <-m.beforeGameConnection:
			m.mu.RLock()
			//updating list of players
			var listOfPlayers []string
			for username := range m.clientsMap {
				listOfPlayers = append(listOfPlayers, username)
			}
			if m.Owner.PlayStyle == 2 {
				if m.Owner.Client.isOnline {
					m.Owner.Client.beforeGameWriterCh <- listOfPlayers
				}
			}
			for _, client := range m.clientsMap {
				client.beforeGameWriterCh <- listOfPlayers
			}
			length := len(m.clientsMap)
			m.mu.RUnlock()
			//checking if lobby is full
			if length == m.maxPlayers {
				m.startGameCh <- true
				m.writeQuestionCh <- true
				m.stopBeforeGameTickerCh <- true
				go m.GameTimeHandler()
				return
			}
			m.updateBeforeGameTickerCh <- true
			log.Println("addded connection to before game state ,current list : ", listOfPlayers)
		//case to handle client leave
		case <-m.beforeGameLeave:
			//overwrite list
			m.mu.RLock()
			var listOfPlayers []string
			for username := range m.clientsMap {
				listOfPlayers = append(listOfPlayers, username)
			}
			if m.Owner.PlayStyle == 2 {
				if m.Owner.Client.isOnline {
					m.Owner.Client.beforeGameWriterCh <- listOfPlayers
				}
			}
			for _, client := range m.clientsMap {
				client.beforeGameWriterCh <- listOfPlayers
			}
			m.mu.RUnlock()
			log.Println("deleted connection from before game state ,current list : ", listOfPlayers)
		case <-m.forcedStartOfGame:
			log.Println("force start of game ")
			m.startGameCh <- true
			m.writeQuestionCh <- true
			m.stopBeforeGameTickerCh <- true
			go m.GameTimeHandler()
			return
		}
	}
}

// function to handle time updates
func (m *Manager) GameTimeHandler() {
	ticker := time.NewTicker(time.Second * 1)
	defer func() {
		ticker.Stop()
		close(m.writeTimeCh)
	}()
	var count int = m.config.TimeForAnswer
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			count--
			//checking if time is up
			if count == 0 {
				//updating counter
				count = m.config.TimeForAnswer
				m.currTime = count

				var tempWaitList []string
				m.mu.RLock()
				for username := range m.clientsMap {
					tempWaitList = append(tempWaitList, username)
				}
				m.mu.RUnlock()
				m.waitList = tempWaitList
				m.updateQuestionCh <- true
				m.writeTimeCh <- true
			} else {
				m.currTime = count
				m.writeTimeCh <- true
			}
		case <-m.restartTimeCh:
			ticker.Reset(time.Second * 1)
			count = m.config.TimeForAnswer
			m.currTime = count
			m.writeTimeCh <- true
		}

	}
}

func (m *Manager) BeforeGameLiquidationHandler() {
	ticker := time.NewTicker(time.Second * 1)
	counter := m.config.TimeForRoomLiquidation
	defer func() {
		ticker.Stop()
		close(m.updateBeforeGameTickerCh)
		close(m.stopBeforeGameTickerCh)
	}()
	for {
		select {
		case <-m.updateBeforeGameTickerCh:
			ticker.Reset(time.Second)
			counter = m.config.TimeForRoomLiquidation
		case <-ticker.C:
			counter--
			if counter == 0 {
				m.cancel()
				return
			}
		case <-m.stopBeforeGameTickerCh:
			m.gameState = 1
			log.Println("room will not liquidate itself :)")
			return
		}

	}
}
