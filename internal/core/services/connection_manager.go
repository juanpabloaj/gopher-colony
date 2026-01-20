package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/coder/websocket"

	"github.com/juanpabloaj/gophercolony/internal/adapters/secondary/websockets"
	"github.com/juanpabloaj/gophercolony/internal/core/domain"
	"github.com/juanpabloaj/gophercolony/internal/core/ports"
)

type ConnectionManager struct {
	logger *slog.Logger
}

func NewConnectionManager(logger *slog.Logger) *ConnectionManager {
	return &ConnectionManager{
		logger: logger,
	}
}

func (s *ConnectionManager) HandleConnection(w http.ResponseWriter, r *http.Request) {
	// 1. Parse Room ID from Query Params
	roomIDStr := r.URL.Query().Get("room")
	if roomIDStr == "" {
		roomIDStr = "lobby" // Default
	}
	roomID := domain.RoomID(roomIDStr)

	// Accept the WebSocket connection
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"}, // Allow all origins for dev
	})
	if err != nil {
		s.logger.Error("failed to accept websocket", "error", err)
		return
	}

	// In Phase 1, we just wrap it and log.
	socketAdapter := websockets.NewAdapter(c)

	s.logger.Info("New client connected", "remote_addr", r.RemoteAddr, "room_id", roomID)

	ctx := context.Background()

	// Echo loop for Phase 1 verification
	err = socketAdapter.Listen(ctx, func(msg []byte) {
		s.logger.Info("Received message", "msg", string(msg), "room_id", roomID)
		response := fmt.Sprintf("[%s] echo: %s", roomID, string(msg))
		if err := socketAdapter.Send(ctx, []byte(response)); err != nil {
			s.logger.Error("failed to echo", "error", err)
		}
	})

	if err != nil {
		s.logger.Info("Client disconnected", "reason", err)
	}
}

var _ ports.ConnectionService = &ConnectionManager{}
