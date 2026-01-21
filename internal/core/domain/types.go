package domain

import "sync"

type PlayerID string
type RoomID string

// Player represents a connected user in the game.
type Player struct {
	ID   PlayerID
	Name string
}

// Room represents a game instance/world.
type Room struct {
	ID    RoomID
	World *World
	// mu is unexported to force usage of thread-safe methods
	mu sync.RWMutex
}

// Snapshot returns a thread-safe copy of the current world state.
func (r *Room) Snapshot() GameStatePayload {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tiles []Tile
	if r.World != nil {
		for y := 0; y < r.World.Height; y++ {
			for x := 0; x < r.World.Width; x++ {
				tiles = append(tiles, *r.World.Grid[y][x])
			}
		}

		return GameStatePayload{
			RoomID: r.ID,
			Width:  r.World.Width,
			Height: r.World.Height,
			Tiles:  tiles,
		}
	}

	// Return empty if no world
	return GameStatePayload{RoomID: r.ID}
}
