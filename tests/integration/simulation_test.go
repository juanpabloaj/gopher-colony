package integration

import (
	"log/slog"
	"math/rand"
	"os"
	"testing"

	"github.com/juanpabloaj/gophercolony/internal/adapters/secondary/memsockets"
	"github.com/juanpabloaj/gophercolony/internal/core/domain"
	"github.com/juanpabloaj/gophercolony/internal/core/services"
)

func TestSimulationGrowth(t *testing.T) {
	// Setup Core Logic without Ticker (Manual Ticking for Determinism)
	// We want to verify that the SimulationService correctly transforms Saplings to Trees.

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	// Use fixed seed for deterministic behavior
	rng := rand.New(rand.NewSource(42))
	simService := services.NewSimulationService(logger, services.WithRNG(rng))

	// Create a room with a specific tile as Sapling
	mapGen := services.NewMapGenerator()
	roomRepo := memsockets.NewRoomManager(mapGen)
	room, _ := roomRepo.CreateRoom("sim_test")

	// Force set a sapling
	room.SetTile(5, 5, domain.TerrainSapling)

	// Verify initial state
	tile, _ := room.World.GetTile(5, 5)
	if tile.Terrain != domain.TerrainSapling {
		t.Fatalf("Failed to setup sapling")
	}

	// Run Tick multiple times until it grows.
	// Probability is 10%. 100 ticks should be enough to be almost certain.
	// If it never grows, test fails (or logic is broken).
	grew := false
	for i := 0; i < 200; i++ {
		changes := simService.Tick(room)
		for _, c := range changes.Tiles {
			if c.X == 5 && c.Y == 5 && c.Terrain == domain.TerrainTree {
				grew = true
				break
			}
		}
		if grew {
			break
		}
	}

	if !grew {
		t.Errorf("Sapling did not grow into a Tree after 200 ticks")
	}

	// Verify Room State
	tile, _ = room.World.GetTile(5, 5)
	if tile.Terrain != domain.TerrainTree {
		t.Errorf("Room state mismatch. Expected Tree (4), got %d", tile.Terrain)
	}
}

func TestSimulationHarvesting(t *testing.T) {
	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	rng := rand.New(rand.NewSource(42)) // Deterministic
	simService := services.NewSimulationService(logger, services.WithRNG(rng))

	mapGen := services.NewMapGenerator()
	roomRepo := memsockets.NewRoomManager(mapGen)
	room, _ := roomRepo.CreateRoom("harvest_test")

	// Setup Condition: Gopher at 5,5. Tree at 5,6.
	gopher := &domain.Gopher{ID: "worker", X: 5, Y: 5}
	room.AddGopher(gopher)
	room.SetTile(5, 6, domain.TerrainTree)

	// Tick
	// Gopher logic checks neighbors.
	// offset {0,1} -> 5,6.
	changes := simService.Tick(room)

	// Verify Changes
	foundChange := false
	for _, tile := range changes.Tiles {
		if tile.X == 5 && tile.Y == 6 && tile.Terrain == domain.TerrainSapling {
			foundChange = true
			break
		}
	}
	if !foundChange {
		t.Errorf("Expected Tree at 5,6 to become Sapling")
	}

	// Verify Inventory (Need to refetch gopher from room as changes only contain DELTA)
	// But changes.Gophers also has the updated gopher state.
	updatedGophers := room.GetGophers()
	if len(updatedGophers) != 1 {
		t.Fatalf("Expected 1 gopher")
	}
	if updatedGophers[0].Inventory.Wood != 1 {
		t.Errorf("Expected 1 wood, got %d", updatedGophers[0].Inventory.Wood)
	}
}
