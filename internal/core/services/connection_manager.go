package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/coder/websocket"

	"github.com/juanpabloaj/gophercolony/internal/adapters/secondary/websockets"
	"github.com/juanpabloaj/gophercolony/internal/core/domain"
	"github.com/juanpabloaj/gophercolony/internal/core/ports"
)

type ConnectionManager struct {
	logger   *slog.Logger
	roomRepo ports.RoomRepository
}

func NewConnectionManager(logger *slog.Logger, roomRepo ports.RoomRepository) *ConnectionManager {
	return &ConnectionManager{
		logger:   logger,
		roomRepo: roomRepo,
	}
}

func (s *ConnectionManager) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// 1. Parse Room ID from Query Params
	roomIDStr := r.URL.Query().Get("room")
	if roomIDStr == "" {
		roomIDStr = "lobby" // Default
	}
	roomID := domain.RoomID(roomIDStr)

	// 2. Get or Create Room (and World)
	// We use CreateRoom which handles get-or-create logic in our simple repo
	room := s.roomRepo.CreateRoom(roomID)

	// Accept the WebSocket connection
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"}, // Allow all origins for dev
	})
	if err != nil {
		s.logger.Error("failed to accept websocket", "error", err)
		return
	}

	socketAdapter := websockets.NewAdapter(c)
	ctx := context.Background()

	s.logger.Info("New client connected", "remote_addr", r.RemoteAddr, "room_id", roomID)

	// 3. Send Initial Game State
	if err := s.sendGameState(ctx, socketAdapter, room); err != nil {
		s.logger.Error("failed to send game state", "error", err)
		return
	}

	// Echo loop for Phase 1 verification (and keep alive for Phase 2)
	err = socketAdapter.Listen(ctx, func(msg []byte) {
		s.logger.Info("Received message", "msg", string(msg), "room_id", roomID)
		// We still echo for debug, but validation will rely on the INIT message
		response := fmt.Sprintf("[%s] echo: %s", roomID, string(msg))
		if err := socketAdapter.Send(ctx, []byte(response)); err != nil {
			s.logger.Error("failed to echo", "error", err)
		}
	})

	if err != nil {
		s.logger.Info("Client disconnected", "reason", err)
	}
}

func (s *ConnectionManager) sendGameState(ctx context.Context, socket ports.Socket, room *domain.Room) error {
	// Use Snapshot() to safely retrieve state
	payload := room.Snapshot()

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	msg := domain.Message{
		Type:    domain.MsgTypeInit,
		Payload: payloadBytes,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	return socket.Send(ctx, msgBytes)
}

var _ ports.ConnectionService = &ConnectionManager{}
