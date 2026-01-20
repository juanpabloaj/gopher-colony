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
	"github.com/juanpabloaj/gophercolony/internal/core/services"
)

func TestWebSocketConnection(t *testing.T) {
	// 1. Setup Server
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	connManager := services.NewConnectionManager(logger)

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

	// 3. Send Message
	msg := []byte("hello")
	if err := c.Write(ctx, websocket.MessageText, msg); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	// 4. Verify Echo (The current implementation echoes with [lobby])
	_, data, err := c.Read(ctx)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	expected := "[lobby] echo: hello"
	if string(data) != expected {
		t.Errorf("Expected %q, got %q", expected, string(data))
	}

	c.Close(websocket.StatusNormalClosure, "bye")
}

func TestRoomConnectivity(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	connManager := services.NewConnectionManager(logger)
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

	if err := c.Write(ctx, websocket.MessageText, []byte("test")); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	_, data, err := c.Read(ctx)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	expected := "[alpha] echo: test"
	if string(data) != expected {
		t.Errorf("Room verification failed. Expected %q, got %q", expected, string(data))
	}

	c.Close(websocket.StatusNormalClosure, "bye")
}

func TestConcurrentConnections(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	connManager := services.NewConnectionManager(logger)
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

			if err := c.Write(ctx, websocket.MessageText, []byte("ping")); err != nil {
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
			expected := fmt.Sprintf("[%s] echo: ping", roomID)
			if string(data) != expected {
				t.Errorf("Client %d: expected %q, got %q", id, expected, string(data))
				done <- false
				return
			}
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

func TestHTTPServerStartup(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	connManager := services.NewConnectionManager(logger)

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
