package websockets

import (
	"context"
	"errors"
	"fmt"
	"io"

	"sync"

	"github.com/coder/websocket"

	"github.com/juanpabloaj/gophercolony/internal/core/ports"
)

// Adapter implements ports.Socket using github.com/coder/websocket.
type Adapter struct {
	conn      *websocket.Conn
	sendChan  chan []byte
	closeOnce sync.Once
}

// NewAdapter creates a new WebSocket adapter.
func NewAdapter(conn *websocket.Conn) *Adapter {
	return &Adapter{
		conn:     conn,
		sendChan: make(chan []byte, 256), // Buffered channel for backpressure
	}
}

// Send queues a message to be sent. Non-blocking.
func (a *Adapter) Send(ctx context.Context, msg []byte) error {
	select {
	case a.sendChan <- msg:
		return nil
	default:
		return fmt.Errorf("client buffer full, dropping message")
	}
}

// Close closes the connection.
func (a *Adapter) Close(code int) error {
	var err error
	a.closeOnce.Do(func() {
		close(a.sendChan)
		err = a.conn.Close(websocket.StatusCode(code), "closing")
	})
	return err
}

func (a *Adapter) writePump() {
	ctx := context.Background()
	for msg := range a.sendChan {
		// We use a separate context for writes to ensure they complete
		// independent of the request context (which might be cancelled)
		// but we respect the connection state.
		if err := a.conn.Write(ctx, websocket.MessageText, msg); err != nil {
			break // Connection closed or error
		}
	}
}

// Listen starts a loop to read messages. It blocks until error or close.
func (a *Adapter) Listen(ctx context.Context, onMessage func(msg []byte)) error {
	// Start Write Pump: linked to the lifecycle of the Read Loop
	go a.writePump()

	defer a.Close(int(websocket.StatusNormalClosure)) // Ensure cleanup on read error

	for {
		typ, reader, err := a.conn.Reader(ctx)
		if err != nil {
			var closeError websocket.CloseError
			if errors.As(err, &closeError) {
				return nil // Clean close
			}
			return fmt.Errorf("read error: %w", err)
		}

		if typ != websocket.MessageText {
			continue // We only support text for now
		}

		payload, err := io.ReadAll(reader)
		if err != nil {
			return fmt.Errorf("read payload error: %w", err)
		}

		onMessage(payload)
	}
}

// Ensure Adapter implements ports.Socket
var _ ports.Socket = &Adapter{}
