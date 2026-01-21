package domain

import "encoding/json"

// MessageType defines the type of websocket message.
type MessageType string

const (
	MsgTypeInit  MessageType = "init"
	MsgTypeEcho  MessageType = "echo"
	MsgTypeError MessageType = "error"
)

// Message is the standard container for all WS communication.
type Message struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// GameStatePayload represents the initial state sent to a client.
type GameStatePayload struct {
	RoomID RoomID `json:"room_id"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Tiles  []Tile `json:"tiles"` // Flattened for simple JSON serialization
}
