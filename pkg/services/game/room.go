package game

import (
	"sync"

	"github.com/google/uuid"
)

// temporary
type RoomManager struct {
	mu    sync.Mutex
	rooms map[string]*Manager
}

var globalRoomManager = &RoomManager{
	rooms: make(map[string]*Manager),
}

func (rm *RoomManager) CreateRoom() string {

	roomID := uuid.NewString()
	manager := NewManager(roomID)
	rm.mu.Lock()
	rm.rooms[roomID] = manager
	rm.mu.Unlock()
	return roomID
}

func (rm *RoomManager) GetManager(roomID string) (*Manager, bool) {
	rm.mu.Lock()
	manager, exists := rm.rooms[roomID]
	rm.mu.Unlock()
	return manager, exists
}
