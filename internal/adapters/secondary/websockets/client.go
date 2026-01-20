package websockets

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/coder/websocket"

	"github.com/juanpabloaj/gophercolony/internal/core/ports"
)

// Adapter implements ports.Socket using github.com/coder/websocket.
type Adapter struct {
	conn *websocket.Conn
}

// NewAdapter creates a new WebSocket adapter.
func NewAdapter(conn *websocket.Conn) *Adapter {
	return &Adapter{
		conn: conn,
	}
}

// Send writes a message to the websocket.
func (a *Adapter) Send(ctx context.Context, msg []byte) error {
	return a.conn.Write(ctx, websocket.MessageText, msg)
}

// Close closes the connection with a status code.
func (a *Adapter) Close(code int) error {
	return a.conn.Close(websocket.StatusCode(code), "closing")
}

// Listen starts a loop to read messages. It blocks until error or close.
func (a *Adapter) Listen(ctx context.Context, onMessage func(msg []byte)) error {
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
