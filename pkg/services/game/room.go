package game

import (
	"sync"

	"github.com/Megidy/k/types"
)

// temporary
type RoomManager struct {
	mu    sync.Mutex
	rooms map[string]*Manager
}

var globalRoomManager = &RoomManager{
	rooms: make(map[string]*Manager),
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
