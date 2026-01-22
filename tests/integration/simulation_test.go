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

func TestSimulationDelivery(t *testing.T) {
	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	rng := rand.New(rand.NewSource(42)) // Deterministic
	simService := services.NewSimulationService(logger, services.WithRNG(rng))

	mapGen := services.NewMapGenerator()
	roomRepo := memsockets.NewRoomManager(mapGen)
	room, _ := roomRepo.CreateRoom("delivery_test")

	// Setup: Gopher with Wood at 5,5. Chest at 5,6.
	gopher := &domain.Gopher{
		ID: "carrier",
		X:  5, Y: 5,
		Inventory: domain.Inventory{Wood: 5},
	}
	room.AddGopher(gopher)
	room.SetTile(5, 6, domain.TerrainChest)

	// Tick
	// Delivery priority: Check adjacent chest -> Deposit
	changes := simService.Tick(room)

	// Verify Room Resources
	if room.Resources["wood"] != 5 {
		t.Errorf("Expected Room to have 5 wood, got %d", room.Resources["wood"])
	}

	// Verify Gopher Inventory
	updatedGophers := room.GetGophers()
	if updatedGophers[0].Inventory.Wood != 0 {
		t.Errorf("Expected Gopher to have 0 wood, got %d", updatedGophers[0].Inventory.Wood)
	}

	// Verify Changes payload contains resources
	if val, ok := changes.Resources["wood"]; !ok || val != 5 {
		t.Errorf("Expected changes payload to include resources: %v", changes.Resources)
	}
}

func TestSimulationDeliveryMovement(t *testing.T) {
	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	rng := rand.New(rand.NewSource(42))
	simService := services.NewSimulationService(logger, services.WithRNG(rng))

	mapGen := services.NewMapGenerator()
	roomRepo := memsockets.NewRoomManager(mapGen)
	room, _ := roomRepo.CreateRoom("delivery_movement_test")

	// Ensure path is clear (remove random trees/stones)
	for x := range 32 {
		for y := range 32 {
			room.SetTile(x, y, domain.TerrainGrass)
		}
	}

	// Setup: Chest at 16,16 (Target)
	room.SetTile(16, 16, domain.TerrainChest)

	// Setup: Gopher at 6,16 (10 tiles to the left) with FULL Wood
	gopher := &domain.Gopher{
		ID:        "carrier_long_dist",
		X:         6,
		Y:         16,
		Inventory: domain.Inventory{Wood: 10},
	}
	room.AddGopher(gopher)

	// Simulate Ticks
	// Distance is 10.
	// Should take ~10 ticks to arrive + 1 to deposit.
	// Give it 20 ticks to be safe.
	deposited := false
	for range 20 {
		changes := simService.Tick(room)

		// Check if wood was deposited in this tick
		if val, ok := changes.Resources["wood"]; ok && val == 10 {
			deposited = true
			break
		}
	}

	if !deposited {
		t.Fatalf("Gopher failed to deposit wood within 40 ticks")
	}

	// Verify Final State
	if room.Resources["wood"] != 10 {
		t.Errorf("Expected Room to have 10 wood, got %d", room.Resources["wood"])
	}

	storedGophers := room.GetGophers()
	if storedGophers[0].Inventory.Wood != 0 {
		t.Errorf("Expected Gopher inventory to be empty, got %d", storedGophers[0].Inventory.Wood)
	}

	// Verify position is adjacent (should be at 15,16 or stayed there)
	finalGopher := storedGophers[0]
	// Calculate distance to chest
	dist := (finalGopher.X-16)*(finalGopher.X-16) + (finalGopher.Y-16)*(finalGopher.Y-16)
	if dist > 2 { // Adjacent means dist 1 (cardinal) or 2 (diagonal)
		t.Errorf("Expected Gopher to be adjacent to chest (16,16), got at %d,%d", finalGopher.X, finalGopher.Y)
	}
}

func TestSimulationDiagonalPath(t *testing.T) {
	// Setup
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	rng := rand.New(rand.NewSource(42))
	simService := services.NewSimulationService(logger, services.WithRNG(rng))

	mapGen := services.NewMapGenerator()
	roomRepo := memsockets.NewRoomManager(mapGen)
	room, _ := roomRepo.CreateRoom("diagonal_path_test")

	// Chest at 10,10
	room.SetTile(10, 10, domain.TerrainChest)

	// Gopher at 14,14 (Diagonal distance)
	gopher := &domain.Gopher{
		ID: "diagonal_walker",
		X:  14, Y: 14,
		Inventory: domain.Inventory{Wood: 10},
	}
	room.AddGopher(gopher)

	// Clear area around
	for y := 8; y < 16; y++ {
		for x := 8; x < 16; x++ {
			room.SetTile(x, y, domain.TerrainGrass)
		}
	}

	// Place an obstacle that BLOCKS diagonal "greedy" path
	// If path is 14,14 -> 13,13 -> 12,12...
	// Block 13,13 with Stone.
	room.SetTile(13, 13, domain.TerrainStone)

	// In a greedy algorithm, Gopher at 14,14 wants to go to 10,10.
	// dx = -1, dy = -1. Target: 13,13.
	// If 13,13 is Stone, MoveGopher returns false.
	// Gopher stays at 14,14. STUCK.

	// Run 1 tick.
	simService.Tick(room)

	updated := room.GetGophers()[0]
	if updated.X == 14 && updated.Y == 14 {
		t.Logf("Confirmed: Gopher got stuck on diagonal obstacle as expected with greedy logic.")
	} else {
		t.Logf("Gopher moved to %d,%d", updated.X, updated.Y)
	}
}
