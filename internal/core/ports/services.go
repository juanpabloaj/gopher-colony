package ports

import (
	"context"
	"net/http"

	"github.com/juanpabloaj/gophercolony/internal/core/domain"
)

// ConnectionService defines how the primary adapter (HTTP) hands off connections to the core.
type ConnectionService interface {
	HandleConnection(w http.ResponseWriter, r *http.Request)
	Broadcast(roomID domain.RoomID, msg []byte)
}

type RoomLifecycleDelegate interface {
	OnRoomCreated(room *domain.Room)
}

// Socket represents the abstraction of a WebSocket connection (Output Port).
// It decouples the core from the specific websocket library.
type Socket interface {
	Send(ctx context.Context, msg []byte) error
	Close(code int) error
	// Listen starts a loop to read messages. It should block until error or close.
	Listen(ctx context.Context, onMessage func(msg []byte)) error
}

// RoomRepository defines storage for active rooms.
type RoomRepository interface {
	GetRoom(id domain.RoomID) (*domain.Room, bool)
	CreateRoom(id domain.RoomID) (*domain.Room, bool)
}

// MapGenerator defines logic to create new worlds.
type MapGenerator interface {
	Generate(width, height int) *domain.World
}
