package services

import (
	"log/slog"
	"math/rand"

	"github.com/juanpabloaj/gophercolony/internal/core/domain"
)

// SimulationService handles the core game loop logic
type SimulationService struct {
	logger *slog.Logger
}

func NewSimulationService(logger *slog.Logger) *SimulationService {
	return &SimulationService{
		logger: logger,
	}
}

// Tick advances the state of the room by one step.
// It returns a list of tiles that changed.
func (s *SimulationService) Tick(room *domain.Room) []domain.Tile {
	// Retrieve a snapshot of the world (Wait for Read Lock internally if needed,
	// but Room.Snapshot() does that).
	// Actually, to modify the world, we need to iterate and SetTile.
	// Ideally, we shouldn't hold the lock for the whole iteration if the world is huge,
	// but for now, we will inspect the snapshot and apply changes.
	// BETTER: Room should provide an iterator or we use Snapshot.
	// Using Snapshot is safe but might be slightly stale if clicks happen, which is fine.

	snapshot := room.Snapshot()
	var changes []domain.Tile

	for _, tile := range snapshot.Tiles {
		// Rule: Grow Sapling -> Tree
		if tile.Terrain == domain.TerrainSapling {
			// 10% chance to grow per tick (1Hz means ~10s expected value)
			if rand.Float64() < 0.1 {
				newTile := domain.Tile{
					X:       tile.X,
					Y:       tile.Y,
					Terrain: domain.TerrainTree,
				}

				// Apply change to Room (Hardware lock)
				// Note: If room changed in between, we overwrite.
				// For "Grow", this is acceptable.
				if room.SetTile(newTile.X, newTile.Y, newTile.Terrain) {
					changes = append(changes, newTile)
				}
			}
		}
	}

	return changes
}
