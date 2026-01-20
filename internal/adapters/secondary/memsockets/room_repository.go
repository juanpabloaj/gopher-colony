package services

import (
	"sync"

	"github.com/juanpabloaj/gophercolony/internal/core/domain"
	"github.com/juanpabloaj/gophercolony/internal/core/ports"
)

type RoomManager struct {
	rooms map[domain.RoomID]*domain.Room
	mu    sync.RWMutex
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[domain.RoomID]*domain.Room),
	}
}

func (rm *RoomManager) GetRoom(id domain.RoomID) (*domain.Room, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[id]
	return room, exists
}

func (rm *RoomManager) CreateRoom(id domain.RoomID) *domain.Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Double check
	if room, exists := rm.rooms[id]; exists {
		return room
	}

	room := &domain.Room{
		ID: id,
	}
	rm.rooms[id] = room
	return room
}

var _ ports.RoomRepository = &RoomManager{}
