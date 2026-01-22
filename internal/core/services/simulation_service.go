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
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Tick advances the state of the room by one step.
func (s *SimulationService) Tick(room *domain.Room) domain.UpdatePayload {
	var changes domain.UpdatePayload

	s.simulateGophers(room, &changes)
	s.simulatePlants(room, &changes)

	changes.Resources = room.GetResources()

	return changes
}

func (s *SimulationService) simulatePlants(room *domain.Room, changes *domain.UpdatePayload) {
	snapshot := room.Snapshot()

	for _, tile := range snapshot.Tiles {
		if tile.Terrain == domain.TerrainSapling {
			// 10% chance to grow per tick
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
			x := s.rng.Intn(32)
			y := s.rng.Intn(32)

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

	// Gopher Behavior Logic
	for _, g := range gophers {
		// 1. Delivery
		if ok, updatedGopher := s.handleDelivery(g, room, changes); ok {
			changes.Gophers = append(changes.Gophers, updatedGopher)
			continue
		}

		// 2. Harvesting
		if ok, updatedGopher := s.handleHarvesting(g, room, changes); ok {
			changes.Gophers = append(changes.Gophers, updatedGopher)
			continue
		}

		// 3. Wandering
		if ok, updatedGopher := s.handleWandering(g, room); ok {
			changes.Gophers = append(changes.Gophers, updatedGopher)
		}
	}
}

func (s *SimulationService) handleDelivery(g domain.Gopher, room *domain.Room, changes *domain.UpdatePayload) (bool, domain.Gopher) {
	if g.Inventory.Wood == 0 {
		return false, g
	}

	offsets := [][]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}

	// Check if adjacent to chest
	for _, offset := range offsets {
		tX, tY := g.X+offset[0], g.Y+offset[1]
		tile, ok := room.GetTile(tX, tY)
		if ok && tile.Terrain == domain.TerrainChest {
			// Deposit
			amount := g.Inventory.Wood
			updatedGopher := g
			updatedGopher.Inventory.Wood = 0
			updatedGopher.State = domain.GopherStateIdle // Reset state after deposit

			room.DepositResource("wood", amount)
			room.UpdateGopher(&updatedGopher)

			return true, updatedGopher
		}
	}

	// Move towards chest if inventory full
	if g.Inventory.Wood >= 10 {
		chestX, chestY := 16, 16 // Target center
		dx, dy := 0, 0
		if g.X < chestX {
			dx = 1
		} else if g.X > chestX {
			dx = -1
		}
		if g.Y < chestY {
			dy = 1
		} else if g.Y > chestY {
			dy = -1
		}

		if dx != 0 || dy != 0 {
			newX, newY := g.X+dx, g.Y+dy
			if room.MoveGopher(g.ID, newX, newY) {
				updatedGopher := g
				updatedGopher.X = newX
				updatedGopher.Y = newY
				updatedGopher.State = domain.GopherStateMoving
				return true, updatedGopher
			}
		}
	}

	return false, g
}

func (s *SimulationService) handleHarvesting(g domain.Gopher, room *domain.Room, changes *domain.UpdatePayload) (bool, domain.Gopher) {
	if g.Inventory.Wood >= 10 {
		return false, g
	}

	offsets := [][]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}

	for _, offset := range offsets {
		tX, tY := g.X+offset[0], g.Y+offset[1]
		tile, ok := room.GetTile(tX, tY)
		if ok && tile.Terrain == domain.TerrainTree {
			if room.SetTile(tX, tY, domain.TerrainSapling) {
				updatedGopher := g
				updatedGopher.Inventory.Wood++
				updatedGopher.State = domain.GopherStateHarvesting

				room.UpdateGopher(&updatedGopher)
				changes.Tiles = append(changes.Tiles, domain.Tile{X: tX, Y: tY, Terrain: domain.TerrainSapling})

				return true, updatedGopher
			}
		}
	}
	return false, g
}

func (s *SimulationService) handleWandering(g domain.Gopher, room *domain.Room) (bool, domain.Gopher) {
	// 20% chance to move randomly
	if s.rng.Float64() < 0.2 {
		dx := s.rng.Intn(3) - 1 // -1, 0, 1
		dy := s.rng.Intn(3) - 1

		if dx == 0 && dy == 0 {
			return false, g
		}

		newX, newY := g.X+dx, g.Y+dy
		if room.MoveGopher(g.ID, newX, newY) {
			updatedGopher := g
			updatedGopher.X = newX
			updatedGopher.Y = newY
			updatedGopher.State = domain.GopherStateMoving
			return true, updatedGopher
		}
	}
	return false, g
}
