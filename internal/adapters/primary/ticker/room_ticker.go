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
			// Tick the simulation
			changes := t.simService.Tick(t.room)
			if len(changes.Tiles) > 0 || len(changes.Gophers) > 0 {
				t.broadcastChanges(changes)
			}
		}
	}
}

func (t *RoomTicker) Stop() {
	close(t.stopChan)
}

func (t *RoomTicker) broadcastChanges(updates domain.UpdatePayload) {
	// Marshal the specific payload first
	payloadBytes, err := json.Marshal(updates)
	if err != nil {
		t.logger.Error("Failed to marshal update payload", "error", err)
		return
	}

	msg := domain.Message{
		Type:    domain.MsgTypeUpdate,
		Payload: payloadBytes,
	}

	// Double marshal for the envelope (ConnectionService expects []byte for the whole message usually?
	// Or does Broadcast logic handle framing?
	// ConnectionService.Broadcast signature is (roomID, []byte).
	// Usually this byte slice is the FINAL WS Text Message.
	// So we need to marshal the outer Message too.

	finalBytes, err := json.Marshal(msg)
	if err != nil {
		t.logger.Error("Failed to marshal wrapper message", "error", err)
		return
	}

	t.connService.Broadcast(t.room.ID, finalBytes)
}
