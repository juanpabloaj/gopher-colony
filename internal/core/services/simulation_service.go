package services

import (
	"log/slog"
	"math/rand"
	"time"

	"github.com/juanpabloaj/gophercolony/internal/core/domain"
	"github.com/juanpabloaj/gophercolony/pkg/utils"
)

// SimulationService handles the core game loop logic
type SimulationService struct {
	logger *slog.Logger
	rng    *rand.Rand
}

type SimulationOption func(*SimulationService)

func WithRNG(rng *rand.Rand) SimulationOption {
	return func(s *SimulationService) {
		s.rng = rng
	}
}

func NewSimulationService(logger *slog.Logger, opts ...SimulationOption) *SimulationService {
	s := &SimulationService{
		logger: logger,
		// Default to time-seeded random if not provided (non-deterministic by default for prod)
		// Or a fixed seed if we wanted reproducibility by default.
		// For games, usually time-seeded is standard unless testing.
		rng: rand.New(rand.NewSource(time.Now().UnixNano())), // Simple default, can replace with time.Now().UnixNano()
	}

	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Tick advances the state of the room by one step.
// It returns a list of tiles/gophers that changed.
func (s *SimulationService) Tick(room *domain.Room) domain.UpdatePayload {
	var changes domain.UpdatePayload

	// 1. Snapshot world (simplified for now to avoid long locking)
	// We operate directly on room with locking for each small operation or use a Safe iterator

	// Gopher Logic
	// ----------------

	// Spawn Gopher (Simple rule: if < 5 gophers, 5% chance to spawn one)
	// We need a safe way to count and add.

	// Better: simulation logic should be centralized or delegates.
	// Let's implement basics here for now.

	// We need to lock the room to read gophers count
	// This is getting complex for a single service method.
	// Ideally Gopher behavior is separate.

	s.simulateGophers(room, &changes)
	s.simulatePlants(room, &changes)

	return changes
}

func (s *SimulationService) simulatePlants(room *domain.Room, changes *domain.UpdatePayload) {
	// Need to iterate tiles.
	// To avoid locking the whole room for long, we might need random sampling or snapshot.
	// For now, snapshot Tiles is ok (Costly copying but safe).
	snapshot := room.Snapshot()

	for _, tile := range snapshot.Tiles {
		if tile.Terrain == domain.TerrainSapling {
			// 10% chance to grow per tick (1Hz means ~10s expected value)
			if s.rng.Float64() < 0.1 {
				newTile := domain.Tile{X: tile.X, Y: tile.Y, Terrain: domain.TerrainTree}
				if room.SetTile(newTile.X, newTile.Y, newTile.Terrain) {
					changes.Tiles = append(changes.Tiles, newTile)
				}
			}
		}
	}
}

func (s *SimulationService) simulateGophers(room *domain.Room, changes *domain.UpdatePayload) {
	gophers := room.GetGophers()

	// Spawn Logic
	if len(gophers) < 5 {
		if s.rng.Float64() < 0.05 { // 5% chance to spawn
			// Find random spawn point
			x := s.rng.Intn(32) // Hardcoded size for now, ideally room.World.Width
			y := s.rng.Intn(32)

			// Simple spawn
			newGopher := &domain.Gopher{
				ID:    utils.GenerateID(),
				X:     x,
				Y:     y,
				State: domain.GopherStateIdle,
			}
			room.AddGopher(newGopher)
			changes.Gophers = append(changes.Gophers, *newGopher)
			return // Spawn only one per tick max
		}
	}

	// Move Logic
	for _, g := range gophers {
		// Random Walk (20% chance)
		if s.rng.Float64() < 0.2 {
			// Random direction
			dx := s.rng.Intn(3) - 1 // -1, 0, 1
			dy := s.rng.Intn(3) - 1

			if dx == 0 && dy == 0 {
				continue
			}

			newX, newY := g.X+dx, g.Y+dy
			if room.MoveGopher(g.ID, newX, newY) {
				// Get updated state
				updatedGopher := g
				updatedGopher.X = newX
				updatedGopher.Y = newY
				updatedGopher.State = domain.GopherStateMoving
				changes.Gophers = append(changes.Gophers, updatedGopher)
			}
		}
	}
}
