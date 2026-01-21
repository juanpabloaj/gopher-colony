package ticker

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/juanpabloaj/gophercolony/internal/core/domain"
	"github.com/juanpabloaj/gophercolony/internal/core/ports"
	"github.com/juanpabloaj/gophercolony/internal/core/services"
)

type RoomTicker struct {
	logger      *slog.Logger
	simService  *services.SimulationService
	connService ports.ConnectionService
	room        *domain.Room
	stopChan    chan struct{}
}

func NewRoomTicker(
	logger *slog.Logger,
	sim *services.SimulationService,
	conn ports.ConnectionService,
	room *domain.Room,
) *RoomTicker {
	return &RoomTicker{
		logger:      logger,
		simService:  sim,
		connService: conn,
		room:        room,
		stopChan:    make(chan struct{}),
	}
}

func (t *RoomTicker) Start() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	t.logger.Info("Starting ticker for room", "room_id", t.room.ID)

	for {
		select {
		case <-t.stopChan:
			t.logger.Info("Stopping ticker for room", "room_id", t.room.ID)
			return
		case <-ticker.C:
			// Execute Simulation Tick
			changes := t.simService.Tick(t.room)
			if len(changes) > 0 {
				t.broadcastChanges(changes)
			}
		}
	}
}

func (t *RoomTicker) Stop() {
	close(t.stopChan)
}

func (t *RoomTicker) broadcastChanges(tiles []domain.Tile) {
	update := domain.UpdatePayload{
		Tiles: tiles,
	}

	bytes, err := json.Marshal(update)
	if err != nil {
		t.logger.Error("failed to marshal ticker update", "error", err)
		return
	}

	msg := domain.Message{
		Type:    domain.MsgTypeUpdate,
		Payload: bytes,
	}

	finalBytes, _ := json.Marshal(msg)

	// Broadcast using ConnectionService
	// Note: We are using Broadcast(roomID, msg)
	t.connService.Broadcast(t.room.ID, finalBytes)
}
