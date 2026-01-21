package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	"github.com/coder/websocket"

	"github.com/juanpabloaj/gophercolony/internal/adapters/secondary/websockets"
	"github.com/juanpabloaj/gophercolony/internal/core/domain"
	"github.com/juanpabloaj/gophercolony/internal/core/ports"
)

type ConnectionManager struct {
	logger   *slog.Logger
	roomRepo ports.RoomRepository
	delegate ports.RoomLifecycleDelegate

	// Registry per room
	clients map[domain.RoomID]map[ports.Socket]struct{}
	mu      sync.RWMutex
}

func NewConnectionManager(logger *slog.Logger, roomRepo ports.RoomRepository, delegate ports.RoomLifecycleDelegate) *ConnectionManager {
	return &ConnectionManager{
		logger:   logger,
		roomRepo: roomRepo,
		delegate: delegate,
		clients:  make(map[domain.RoomID]map[ports.Socket]struct{}),
	}
}

func (s *ConnectionManager) addClient(roomID domain.RoomID, socket ports.Socket) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.clients[roomID]; !ok {
		s.clients[roomID] = make(map[ports.Socket]struct{})
	}
	s.clients[roomID][socket] = struct{}{}
}

func (s *ConnectionManager) removeClient(roomID domain.RoomID, socket ports.Socket) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if clients, ok := s.clients[roomID]; ok {
		delete(clients, socket)
		if len(clients) == 0 {
			delete(s.clients, roomID)
		}
	}
}

func (s *ConnectionManager) Broadcast(roomID domain.RoomID, msg []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients, ok := s.clients[roomID]
	if !ok {
		return
	}

	ctx := context.Background() // Context for send
	for socket := range clients {
		// Send is now non-blocking (buffered channel implementation in Adapter)
		// We handle slow clients by dropping messages (Send returns error)
		if err := socket.Send(ctx, msg); err != nil {
			s.logger.Warn("dropping message to slow client", "error", err)
			// Optional: Disconnect client if buffer full?
			// s.removeClient(roomID, socket)
			// socket.Close(...)
		}
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
	room, created := s.roomRepo.CreateRoom(roomID)

	if created && s.delegate != nil {
		s.delegate.OnRoomCreated(room)
	}

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

	clientID := generateClientID()

	s.addClient(roomID, socketAdapter)
	defer s.removeClient(roomID, socketAdapter)

	s.logger.Debug("New client connected", "client_id", clientID, "remote_addr", r.RemoteAddr, "room_id", roomID)
	defer s.logger.Debug("Client disconnected", "client_id", clientID, "room_id", roomID)

	// 3. Send Initial Game State
	if err := s.sendGameState(ctx, socketAdapter, room); err != nil {
		s.logger.Error("failed to send game state", "error", err)
		return
	}

	// Command loop
	err = socketAdapter.Listen(ctx, func(msgBytes []byte) {
		var msg domain.Message
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			s.logger.Error("failed to unmarshal message", "error", err)
			return
		}

		if msg.Type == domain.MsgTypeCmd {
			s.handleCommand(ctx, room, msg.Payload)
		} else {
			// Phase 1 Echo Fallback (Optional, maybe remove later)
			s.logger.Info("Ignored non-cmd message", "type", msg.Type)
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

func (s *ConnectionManager) handleCommand(ctx context.Context, room *domain.Room, payloadBytes json.RawMessage) {
	var cmd domain.CommandPayload
	if err := json.Unmarshal(payloadBytes, &cmd); err != nil {
		s.logger.Error("failed to unmarshal command", "error", err)
		return
	}

	// Simple Logic: Toggle Terrain
	if cmd.Action == "click" {
		// Toggle logic: Grass -> Stone -> Water -> Grass
		// Need to get current state (safe read first, or blind write)
		// Simpler: Just cycle blindly or use a "Set" command.
		// For Phase 3 Goal, let's just make it Stone if Grass, Water if Stone, Grass if Water.
		// But Room.SetTile takes 'Terrain'. We need to know 'Current'.

		// Optimization: We could add a 'CycleTile' method to Room, but keeping logic here is fine for now if we accept a read-lock then write-lock race, OR we just trust client? No, never trust client.
		// Let's implement Cycle logic inside Room.SetTile? No, SetTile should be explicit.
		// Let's implement func (r *Room) ToggleTile(x, y) (domain.Tile, bool)

		// For now, let's just set to Stone to verify mutation works.
		// TODO: Implement real game logic.
		newTerrain := domain.TerrainStone

		changed := room.SetTile(cmd.X, cmd.Y, newTerrain)
		if changed {
			s.broadcastUpdate(room.ID, cmd.X, cmd.Y, newTerrain)
		}
	}
}

func (s *ConnectionManager) broadcastUpdate(roomID domain.RoomID, x, y int, terrain domain.TerrainType) {
	update := domain.UpdatePayload{
		Tiles: []domain.Tile{
			{X: x, Y: y, Terrain: terrain},
		},
	}

	bytes, _ := json.Marshal(update)
	msg := domain.Message{
		Type:    domain.MsgTypeUpdate,
		Payload: bytes,
	}

	finalBytes, _ := json.Marshal(msg)
	s.Broadcast(roomID, finalBytes)
}

var _ ports.ConnectionService = &ConnectionManager{}

func generateClientID() string {
	b := make([]byte, 4) // 4 bytes = 8 hex chars
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
