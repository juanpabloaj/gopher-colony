package domain

type PlayerID string
type RoomID string

// Player represents a connected user in the game.
type Player struct {
	ID   PlayerID
	Name string
}

// Room represents a game instance/world.
type Room struct {
	ID RoomID
}
