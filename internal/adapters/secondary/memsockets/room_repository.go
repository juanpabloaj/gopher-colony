package memsockets

import (
	"sync"

	"github.com/juanpabloaj/gophercolony/internal/core/domain"
	"github.com/juanpabloaj/gophercolony/internal/core/ports"
	"github.com/juanpabloaj/gophercolony/pkg/utils"
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
	// Use room ID hash as seed for deterministic generation
	seed := utils.HashString(string(id))
	world := rm.generator.Generate(32, 32, seed)

	room := &domain.Room{
		ID:        id,
		World:     world,
		Gophers:   make(map[string]*domain.Gopher),
		Resources: make(map[string]int),
	}
	rm.rooms[id] = room
	return room, true
}

var _ ports.RoomRepository = &RoomManager{}
