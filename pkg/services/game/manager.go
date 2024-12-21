package game

import (
	"log"
	"sync"

	"github.com/Megidy/k/types"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// ReadBufferSize:  1024,
	// WriteBufferSize: 1024,
	// CheckOrigin: func(r *http.Request) bool {
	// 	return false
	// },
}

//TO DO :

//!!!! make password !!

// implement not with the sleep but with the members passsed the current question
type Manager struct {

	//unique of room
	roomID string

	//number of members
	numberOfPlayers int
	//number of questions
	numberOfQuestions int
	mu                sync.Mutex
	//points score
	usernames map[string]bool
	points    map[*Client]int
	//map of all clients
	clients map[*Client]bool
	//question
	currQuestion int
	questions    []types.Question
	broadcast    chan types.Question
	//doneCh to comunicate with message queue
	doneCh chan bool
	done   map[*Client]bool
	//accessCh to start game
	accessCh chan bool
}

func NewManager(id string, numberOfPlayers, amountOfQuestions int, questions []types.Question) *Manager {
	manager := &Manager{
		roomID:            id,
		numberOfPlayers:   numberOfPlayers,
		numberOfQuestions: amountOfQuestions,
		mu:                sync.Mutex{},
		points:            make(map[*Client]int),
		clients:           make(map[*Client]bool),
		done:              make(map[*Client]bool),
		usernames:         make(map[string]bool),
		broadcast:         make(chan types.Question),
		doneCh:            make(chan bool, 1),
		accessCh:          make(chan bool),
		questions:         questions,
		currQuestion:      0,
	}
	go manager.MessageQueue()
	go manager.CheckReadiness()
	go manager.SetClientsInReadiness()
	go manager.StartGame()
	return manager
}

func (m *Manager) CheckDuplicates(client *Client) bool {
	_, ok := m.usernames[client.userName]
	return ok
}

func (m *Manager) AddClientToConnectionPool(client *Client) {

	m.mu.Lock()
	m.usernames[client.userName] = true
	m.clients[client] = true
	m.points[client] = 0
	m.mu.Unlock()
	log.Println("added new client with: ", client)
	log.Println("currenct pool of connection: ", m.clients)

}
func (m *Manager) DeleteClientFromConnectionPool(client *Client) {
	m.mu.Lock()
	delete(m.usernames, client.userName)
	delete(m.clients, client)
	log.Println("deleted connection from pool ", client)
	log.Println("current connections : ", m.clients)
	m.mu.Unlock()
}

// implement function start the game
func (m *Manager) NewConnection(c *gin.Context) {
	u, _ := c.Get("user")
	user := u.(*types.User)

	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("error while creating websocket connection : ", err)
		return
	}
	client := NewClient(user.UserName, wsConn, m, c)
	//checking if user with this
	if m.CheckDuplicates(client) {
		return
	}
	m.AddClientToConnectionPool(client)
	<-m.accessCh
	go client.WritePump()
	go client.ReadPump()
}

func (m *Manager) MessageQueue() {
	for {
		select {
		case done, ok := <-m.doneCh:
			if !ok {
				log.Println("failed to read from doneCh : message queue")
				return
			}
			if m.currQuestion < m.numberOfQuestions {
				if done {
					m.currQuestion++
				}

			}
			if m.currQuestion == m.numberOfQuestions {
				doneQuestion := types.Question{
					Id:       "ID-leaderBoard",
					Question: "leaderboard",
				}
				for {

					for range m.clients {
						m.broadcast <- doneQuestion
					}
				}
			}
		default:
			m.broadcast <- m.questions[m.currQuestion]
		}

	}

}

func (m *Manager) CheckReadiness() {
	for {
		var count int
		var check int
		m.mu.Lock()
		count = len(m.clients)
		m.mu.Unlock()
		if count == 0 {
			continue
		}
		m.mu.Lock()
		for _, ready := range m.done {
			if ready {
				check++
			}
			// log.Println("count :", count)
			// log.Println("check :", check)

			if count == check {
				for k := range m.done {
					m.done[k] = false
				}
				m.doneCh <- true
			}
		}
		m.mu.Unlock()
		check = 0

	}
}

func (m *Manager) SetClientsInReadiness() {
	for {
		m.mu.Lock()
		for c := range m.clients {

			_, ok := m.done[c]
			if !ok {
				m.done[c] = false

			}

		}
		m.mu.Unlock()

	}

}

func (m *Manager) StartGame() {
	for {
		if len(m.clients) == m.numberOfPlayers {
			m.accessCh <- true
			close(m.accessCh)
			return
		}
	}
}
