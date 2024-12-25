package game

import (
	"log"
	"sync"

	"github.com/Megidy/k/types"
)

// temporary
type RoomManager struct {
	mu            sync.Mutex
	rooms         map[string]*Manager
	listOfPlayers map[*Manager][]string
}

var globalRoomManager = &RoomManager{
	rooms:         make(map[string]*Manager),
	listOfPlayers: make(map[*Manager][]string),
}

func (rm *RoomManager) CreateRoom(roomID string, numberOfPlayers, amountOfQuestions int, questions []types.Question) {

	manager := NewManager(roomID, numberOfPlayers, amountOfQuestions, questions)
	rm.mu.Lock()
	rm.rooms[roomID] = manager
	rm.mu.Unlock()
}

func (rm *RoomManager) GetManager(roomID string) (*Manager, bool) {
	rm.mu.Lock()
	manager, exists := rm.rooms[roomID]
	rm.mu.Unlock()
	return manager, exists
}

func (rm *RoomManager) EndRoomSession(roomID string) {
	rm.mu.Lock()
	delete(rm.rooms, roomID)
	rm.mu.Unlock()
}

func (rm *RoomManager) AddConnectionToList(manager *Manager, username string) {
	rm.mu.Lock()

	list := rm.listOfPlayers[manager]
	rm.mu.Unlock()
	list = append(list, username)
	rm.mu.Lock()
	rm.listOfPlayers[manager] = list
	rm.mu.Unlock()
	log.Println("added connetion to global room manager")
}
func (rm *RoomManager) DeleteConnectionFromList(manager *Manager, username string) {
	rm.mu.Lock()
	list := rm.listOfPlayers[manager]
	rm.mu.Unlock()
	for index, value := range list {
		if value == username {
			list = append(list[:index], list[index+1:]...)
		}
	}
	rm.mu.Lock()
	rm.listOfPlayers[manager] = list
	rm.mu.Unlock()
	log.Println("deleted connetion from global room manager")
}
func (rm *RoomManager) CheckDuplicate(manager *Manager, username string) bool {
	rm.mu.Lock()
	list := rm.listOfPlayers[manager]
	rm.mu.Unlock()
	for _, value := range list {
		if value == username {
			return true
		}
	}
	return false

}
