package domain

import "encoding/json"

// MessageType defines the type of websocket message.
type MessageType string

const (
	MsgTypeInit   MessageType = "init"
	MsgTypeEcho   MessageType = "echo"
	MsgTypeError  MessageType = "error"
	MsgTypeCmd    MessageType = "cmd"
	MsgTypeUpdate MessageType = "update"
)

// CommandPayload represents a player action (Input).
type CommandPayload struct {
	Action string `json:"action"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Params string `json:"params,omitempty"`
}

// UpdatePayload represents a partial world update (Output).
type UpdatePayload struct {
	Tiles []Tile `json:"tiles"`
}

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
