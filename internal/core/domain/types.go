package domain

import (
	"sync"
)

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

// SetTile updates a tile safely and returns true if changed.
func (r *Room) SetTile(x, y int, terrain TerrainType) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.World == nil || x < 0 || y < 0 || x >= r.World.Width || y >= r.World.Height {
		return false
	}

	tile := r.World.Grid[y][x]
	if tile.Terrain == terrain {
		return false
	}

	tile.Terrain = terrain
	return true
}

// ToggleTile cycles the terrain type at x,y and returns the new type and true if changed.
func (r *Room) ToggleTile(x, y int) (TerrainType, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.World == nil || x < 0 || y < 0 || x >= r.World.Width || y >= r.World.Height {
		return TerrainGrass, false
	}

	tile := r.World.Grid[y][x]
	switch tile.Terrain {
	case TerrainGrass:
		tile.Terrain = TerrainStone
	case TerrainStone:
		tile.Terrain = TerrainWater
	case TerrainWater:
		tile.Terrain = TerrainGrass
	default:
		tile.Terrain = TerrainGrass
	}

	return tile.Terrain, true
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
