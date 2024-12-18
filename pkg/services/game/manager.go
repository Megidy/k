package game

import (
	"log"
	"sync"

	"github.com/Megidy/k/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// ReadBufferSize:  1024,
	// WriteBufferSize: 1024,
	// CheckOrigin: func(r *http.Request) bool {
	// 	return false
	// },
}

// implement not with the sleep but with the members passsed the current question
type Manager struct {
	mu sync.Mutex

	points       map[*Client]int
	clients      map[*Client]bool
	currQuestion int
	questions    []types.Question
	broadcast    chan types.Question
	doneCh       chan bool
	done         map[*Client]bool
}

func NewManager() *Manager {
	return &Manager{
		mu:        sync.Mutex{},
		points:    make(map[*Client]int),
		clients:   make(map[*Client]bool),
		done:      make(map[*Client]bool),
		broadcast: make(chan types.Question),
		doneCh:    make(chan bool, 1),
		questions: []types.Question{
			{
				Id:            uuid.NewString(),
				Question:      "What is the capital of France?",
				Answers:       []string{"A. Berlin", "B. Madrid", "C. Paris", "D. Rome"},
				CorrectAnswer: "C",
			},
			{
				Id:            uuid.NewString(),
				Question:      "Which programming language is known as the backbone of the web?",
				Answers:       []string{"A. Python", "B. JavaScript", "C. Go", "D. Ruby"},
				CorrectAnswer: "B",
			},
			{
				Id:            uuid.NewString(),
				Question:      "What is the smallest prime number?",
				Answers:       []string{"A. 0", "B. 1", "C. 2", "D. 3"},
				CorrectAnswer: "C",
			},
			{
				Id:            uuid.NewString(),
				Question:      "Which planet is known as the Red Planet?",
				Answers:       []string{"A. Earth", "B. Mars", "C. Jupiter", "D. Venus"},
				CorrectAnswer: "B",
			},
		},
		currQuestion: 0,
	}
}
func (m *Manager) AddClientToConnectionPool(client *Client) error {

	m.mu.Lock()
	m.clients[client] = true
	m.points[client] = 0
	m.mu.Unlock()
	log.Println("added new client with: ", client)
	log.Println("currenct pool of connection: ", m.clients)

	return nil
}
func (m *Manager) DeleteClientFromConnectionPool(client *Client) error {
	m.mu.Lock()
	delete(m.clients, client)
	log.Println("deleted connection from pool ", client)
	log.Println("current connections : ", m.clients)
	m.mu.Unlock()
	return nil
}

// implement function start the game
func (m *Manager) NewConnection(c *gin.Context) {

	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("error while creating websocket connection : ", err)
		return
	}
	client := NewClient(uuid.NewString(), wsConn, m)
	m.AddClientToConnectionPool(client)

	go client.WritePump()
	go client.ReadPump()
}

func (m *Manager) MessageQueue() {
	for {
		select {
		case done, ok := <-m.doneCh:
			if !ok {
				log.Println("failed to read from doneCh : message queue")
			}
			if m.currQuestion < 4 {

				if done {
					m.currQuestion++
				}

			}
			if m.currQuestion == 4 {
				m.currQuestion = 0
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

// func (m *Manager) WriteQuestionToClients(question types.Question) {
// }

// func (m *Manager) WriteQuestionToAllUsers() {
// select {
// case v, ok := <-m.doneCh:
// 	if !ok {
// 		log.Println("done ch is not ok")
// 		return
// 	}
// 	if v {
// 		if m.currQuestion != 4 {

// 			m.currQuestion++
// 			m.doneCh <- false
// 		} else if m.currQuestion == 4 {
// 			m.currQuestion = 0
// 		}
// 	} else {
// 		m.broadcast <- m.questions[m.currQuestion]
// 	}
// }

// m.doneCh <- false

// }
