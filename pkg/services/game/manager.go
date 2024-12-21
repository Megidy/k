package game

import (
	"context"
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
	//mutex for concurenct safe reading and writing
	mu sync.Mutex
	//context
	ctx context.Context
	//cancel func
	cancel context.CancelFunc
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
	ctx, cancel := context.WithCancel(context.Background())

	manager := &Manager{
		roomID:            id,
		numberOfPlayers:   numberOfPlayers,
		numberOfQuestions: amountOfQuestions,
		mu:                sync.Mutex{},
		ctx:               ctx,
		cancel:            cancel,
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
	go manager.StartGame()
	return manager
}

func (m *Manager) CheckDuplicates(client *Client) bool {
	_, ok := m.usernames[client.userName]
	return ok
}

func (m *Manager) AddClientToConnectionPool(client *Client) {

	m.mu.Lock()
	_, ok := m.points[client]
	if !ok {
		m.points[client] = 0
	}
	m.usernames[client.userName] = true
	m.clients[client] = true
	m.done[client] = false
	m.mu.Unlock()
	log.Println("added new client with: ", client)
	log.Println("currenct pool of connection: ", m.clients)

}
func (m *Manager) DeleteClientFromConnectionPool(client *Client) {
	m.mu.Lock()
	delete(m.usernames, client.userName)
	delete(m.clients, client)
	delete(m.done, client)
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

	// checking if user with this
	// if m.CheckDuplicates(&Client{userName: user.UserName}) {
	// 	return
	// }
	client := NewClient(user.UserName, wsConn, m, c)
	m.AddClientToConnectionPool(client)
	<-m.accessCh
	go client.WritePump()
	go client.ReadPump()
}

func (m *Manager) MessageQueue() {
	defer func() {
		log.Println("MESSAGE QUEUE: exited from goroutine")
	}()
	for {
		select {
		case done, ok := <-m.doneCh:
			if !ok {
				log.Println("MESSAGE QUEUE: failed to read from doneCh : message queue")
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
				m.mu.Lock()
				for i := 0; i < len(m.clients)*10; i++ {
					m.broadcast <- doneQuestion
				}

				m.mu.Unlock()
				m.cancel()

				close(m.broadcast)
				return

			}
		default:
			m.broadcast <- m.questions[m.currQuestion]
		}

	}

}

func (m *Manager) CheckReadiness() {

	for {
		select {
		case <-m.ctx.Done():
			log.Println("CHECK READINESS : exited goroutine via ctx")
			close(m.doneCh)
			return
		default:
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
