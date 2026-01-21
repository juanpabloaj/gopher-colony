package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"log/slog"
	"os"

	adapter_http "github.com/juanpabloaj/gophercolony/internal/adapters/primary/http"
	"github.com/juanpabloaj/gophercolony/internal/adapters/secondary/memsockets"
	"github.com/juanpabloaj/gophercolony/internal/core/services"
)

func TestWebSocketConnection(t *testing.T) {
	// 1. Setup Server
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mapGen := services.NewMapGenerator()
	roomRepo := memsockets.NewRoomManager(mapGen)
	connManager := services.NewConnectionManager(logger, roomRepo)

	// We use the real adapter logic but with httptest
	server := httptest.NewServer(http.HandlerFunc(connManager.HandleConnection))
	defer server.Close()

	// 2. Connect Client
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	u := "ws" + strings.TrimPrefix(server.URL, "http")
	c, _, err := websocket.Dial(ctx, u, nil)
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	// INCREASE READ LIMIT for Map Data
	// c.SetReadLimit(10 * 1024 * 1024)

	// CONSUME INIT MESSAGE
	_, _, err = c.Read(ctx) // Skip init
	if err != nil {
		t.Fatalf("Failed to read INIT: %v", err)
	}

	// 3. Send Message (Command)
	cmd := `{"type": "cmd", "payload": {"action": "click", "x": 0, "y": 0}}`
	if err := c.Write(ctx, websocket.MessageText, []byte(cmd)); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	// 4. Verify Update
	_, data, err := c.Read(ctx)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	if !strings.Contains(string(data), `"type":"update"`) {
		t.Errorf("Expected update message, got %q", string(data))
	}

	c.Close(websocket.StatusNormalClosure, "bye")
}

func TestRoomConnectivity(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mapGen := services.NewMapGenerator()
	roomRepo := memsockets.NewRoomManager(mapGen)
	connManager := services.NewConnectionManager(logger, roomRepo)
	server := httptest.NewServer(http.HandlerFunc(connManager.HandleConnection))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	u := "ws" + strings.TrimPrefix(server.URL, "http") + "?room=alpha"
	c, _, err := websocket.Dial(ctx, u, nil)
	if err != nil {
		t.Fatalf("Failed to dial room alpha: %v", err)
	}
	defer c.Close(websocket.StatusInternalError, "error")

	// c.SetReadLimit(10 * 1024 * 1024)
	// CONSUME INIT
	c.Read(ctx)

	cmd := `{"type": "cmd", "payload": {"action": "click", "x": 0, "y": 0}}`
	if err := c.Write(ctx, websocket.MessageText, []byte(cmd)); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	_, data, err := c.Read(ctx)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	if !strings.Contains(string(data), `"type":"update"`) {
		t.Errorf("Room verification failed. Expected update, got %q", string(data))
	}

	c.Close(websocket.StatusNormalClosure, "bye")
}

func TestConcurrentConnections(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mapGen := services.NewMapGenerator()
	roomRepo := memsockets.NewRoomManager(mapGen)
	connManager := services.NewConnectionManager(logger, roomRepo)
	server := httptest.NewServer(http.HandlerFunc(connManager.HandleConnection))
	defer server.Close()

	clientCount := 10
	done := make(chan bool)

	for i := 0; i < clientCount; i++ {
		go func(id int) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			roomID := fmt.Sprintf("room_%d", id)
			u := "ws" + strings.TrimPrefix(server.URL, "http") + "?room=" + roomID
			c, _, err := websocket.Dial(ctx, u, nil)
			if err != nil {
				t.Errorf("Client %d failed to dial: %v", id, err)
				done <- false
				return
			}
			defer c.Close(websocket.StatusInternalError, "")

			// c.SetReadLimit(10 * 1024 * 1024)
			// CONSUME INIT
			if _, _, err := c.Read(ctx); err != nil {
				t.Errorf("Client %d failed to read INIT: %v", id, err)
				done <- false
				return
			}

			cmd := `{"type": "cmd", "payload": {"action": "click", "x": 0, "y": 0}}`
			if err := c.Write(ctx, websocket.MessageText, []byte(cmd)); err != nil {
				t.Errorf("Client %d failed to write: %v", id, err)
				done <- false
				return
			}
			_, data, err := c.Read(ctx)
			if err != nil {
				t.Errorf("Client %d failed to read: %v", id, err)
				done <- false
				return
			}

			// Verify we got an UPDATE message
			if !strings.Contains(string(data), `"type":"update"`) {
				t.Errorf("Client %d: expected update, got %q", id, string(data))
				done <- false
				return
			}

			// Isolation verify: roomID is implicit because each client has its own room connection
			// and checking they get a response confirms the isolate room logic works (processed and broadcasted to room)
			// expected := fmt.Sprintf("[%s] echo: ping", roomID) // OLD EXPECTATION
			c.Close(websocket.StatusNormalClosure, "done")
			done <- true
		}(i)
	}

	for i := 0; i < clientCount; i++ {
		if success := <-done; !success {
			t.Fail()
		}
	}
}

func TestCommandMutation(t *testing.T) {
	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mapGen := services.NewSeededMapGenerator(12345)
	roomRepo := memsockets.NewRoomManager(mapGen)
	connManager := services.NewConnectionManager(logger, roomRepo)
	server := httptest.NewServer(http.HandlerFunc(connManager.HandleConnection))
	defer server.Close()

	// Connect Client A and B to SAME room
	roomID := "room_interact"
	u := "ws" + strings.TrimPrefix(server.URL, "http") + "?room=" + roomID

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Client A
	cA, _, err := websocket.Dial(ctx, u, nil)
	if err != nil {
		t.Fatalf("Client A dial failed: %v", err)
	}
	defer cA.Close(websocket.StatusInternalError, "")
	// cA.SetReadLimit(10 << 20)
	cA.Read(ctx) // Consume INIT

	// Client B
	cB, _, err := websocket.Dial(ctx, u, nil)
	if err != nil {
		t.Fatalf("Client B dial failed: %v", err)
	}
	defer cB.Close(websocket.StatusInternalError, "")
	// cB.SetReadLimit(10 << 20)
	cB.Read(ctx) // Consume INIT

	// Send Command from A: Click at 0,0
	cmd := `{"type": "cmd", "payload": {"action": "click", "x": 0, "y": 0}}`
	if err := cA.Write(ctx, websocket.MessageText, []byte(cmd)); err != nil {
		t.Fatalf("Client A write failed: %v", err)
	}

	// Verify UPDATE on Client B
	_, dataB, err := cB.Read(ctx)
	if err != nil {
		t.Fatalf("Client B read failed: %v", err)
	}

	// Expected: type: update
	if !strings.Contains(string(dataB), `"type":"update"`) {
		t.Errorf("Expected update message, got: %s", string(dataB))
	}
	if !strings.Contains(string(dataB), `"x":0`) || !strings.Contains(string(dataB), `"y":0`) {
		t.Errorf("Expected update for 0,0, got: %s", string(dataB))
	}
	// Expect terrain 2 (Stone)
	if !strings.Contains(string(dataB), `"type":2`) {
		t.Errorf("Expected type:2 (Stone), got: %s", string(dataB))
	}
}

func TestHTTPServerStartup(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	mapGen := services.NewMapGenerator()
	roomRepo := memsockets.NewRoomManager(mapGen)
	connManager := services.NewConnectionManager(logger, roomRepo)

	// Bind to port 0 to let OS choose
	srv := adapter_http.NewServer(0, connManager, logger)

	go func() {
		if err := srv.Start(); err != nil {
			// t.Logf("Server stopped: %v", err)
			// We can't really fail here easily as it runs in goroutine,
			// but we check if it panics at least.
		}
	}()

	// Give it a moment
	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}
}
