package memsockets

import (
	"sync"

	"github.com/juanpabloaj/gophercolony/internal/core/domain"
	"github.com/juanpabloaj/gophercolony/internal/core/ports"
)

type RoomManager struct {
	rooms     map[domain.RoomID]*domain.Room
	generator ports.MapGenerator
	mu        sync.RWMutex
}

func NewRoomManager(gen ports.MapGenerator) *RoomManager {
	return &RoomManager{
		rooms:     make(map[domain.RoomID]*domain.Room),
		generator: gen,
	}
}

func (rm *RoomManager) GetRoom(id domain.RoomID) (*domain.Room, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	room, exists := rm.rooms[id]
	return room, exists
}

func (rm *RoomManager) CreateRoom(id domain.RoomID) (*domain.Room, bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Double check
	if room, exists := rm.rooms[id]; exists {
		return room, false
	}

	// Create new room with a generated world
	// 32x32 is the Phase 2 standard
	world := rm.generator.Generate(32, 32)

	room := &domain.Room{
		ID:    id,
		World: world,
	}
	rm.rooms[id] = room
	return room, true
}

var _ ports.RoomRepository = &RoomManager{}
