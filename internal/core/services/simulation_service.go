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

	// Add current resources state to payload (simple)
	// We could optimize to only send if changed, but map is small.
	// Room.Resources read requires lock.
	// SimulateGophers might have updated it.
	// Let's add a Safe GetResources method or just peek Snapshot?
	// Snapshot is heavy.
	// Since Room.DepositResource locks, we can't read it concurrently easily without lock.
	// Let's rely on changes? No, DepositResource doesn't return changes.
	// We can explicitly read it here.
	changes.Resources = room.GetResources()

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
		didAction := false

		// 0. Try to Deliver Logic (Priority over Harvest)
		if g.Inventory.Wood > 0 {
			// Seek Chest
			// Heuristic: Check if adjacent to chest.
			// Or move towards chest.
			// Ideally we know chest location. For now scan offsets or assume center?
			// Scanning whole map is expensive.
			// Let's interact if adjacent to Chest.

			offsets := [][]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}

			deposited := false
			for _, offset := range offsets {
				tX, tY := g.X+offset[0], g.Y+offset[1]
				tile, ok := room.GetTile(tX, tY)
				if ok && tile.Terrain == domain.TerrainChest {
					// Deposit
					amount := g.Inventory.Wood

					// Update Gopher
					updatedGopher := g
					updatedGopher.Inventory.Wood = 0

					// Update Room Resources
					// We need a method on Room to AddResource?
					// Or we lock and map access. Room.Resources is exported but we need lock.
					// Let's add room.AddResource(type, amount) to types.go?
					// For now, assume we can add a method or do it roughly here if we dare (we can't, race).
					// Let's add AddResource to Room in next step.
					// We will Assume Room.Deposit(resource, amount) exists.

					room.DepositResource("wood", amount)
					room.UpdateGopher(&updatedGopher)

					changes.Gophers = append(changes.Gophers, updatedGopher)
					// changes.Resources? We need to send resource update.
					// UpdatePayload needs Resources map?
					// We added Resources to GameStatePayload.
					// Protocol UpdatePayload needs it too.
					// Let's add it to protocol in next step.

					deposited = true
					didAction = true
					break
				}
			}

			// If not adjacent, we should move towards chest?
			// For simplicity: Random walk for now?
			// Gophers will stumble upon chest eventually.
			// Or we implement basic homing: Center is target.
			if !deposited && g.Inventory.Wood >= 10 {
				// If full, SEEK chest (simple vector)
				chestX, chestY := 16, 16 // Hardcoded or find it?
				// Simple step towards center
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

				// Apply randomized component to avoid getting stuck?
				// Try direct move first
				if dx != 0 || dy != 0 {
					newX, newY := g.X+dx, g.Y+dy
					if room.MoveGopher(g.ID, newX, newY) {
						updatedGopher := g
						updatedGopher.X = newX
						updatedGopher.Y = newY
						updatedGopher.State = domain.GopherStateMoving
						changes.Gophers = append(changes.Gophers, updatedGopher)
						continue // Next gopher
					}
				}
			}

			if deposited {
				continue
			}
		}

		// 1. Try to Harvest Logic
		// Capacity check (limit 10 for now)
		if g.Inventory.Wood < 10 {
			offsets := [][]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
			// Randomize offsets order? Or deterministic? Deterministic is fine.

			for _, offset := range offsets {
				tX, tY := g.X+offset[0], g.Y+offset[1]
				tile, ok := room.GetTile(tX, tY)
				if ok && tile.Terrain == domain.TerrainTree {
					// Attempt Harvest
					// Tree -> Sapling (Sustainable forestry)
					if room.SetTile(tX, tY, domain.TerrainSapling) {
						// Success
						// Update Gopher (Need to update in Room too?
						// Room.Gophers is a map of pointers. If we update local 'g', it's a COPY from GetGophers() value.
						// Wait, GetGophers() returns []Gopher (copies).
						// So we need a way to UpdateGopher in Room.
						// Room.MoveGopher updates position. We need Room.UpdateGopherState/Inventory?
						// Or expose Room.GetGopher(id) *Gopher?
						// Room.Gophers is private.
						// Let's add UpdateGopher method or use a callback?
						// Actually Room.Gophers map holds pointers.
						// But GetGophers returns copies.
						// We need to write back using a new method Room.UpdateGopher(gopher).

						updatedGopher := g
						updatedGopher.Inventory.Wood++
						updatedGopher.State = domain.GopherStateHarvesting

						room.UpdateGopher(&updatedGopher)
						changes.Gophers = append(changes.Gophers, updatedGopher)
						changes.Tiles = append(changes.Tiles, domain.Tile{X: tX, Y: tY, Terrain: domain.TerrainSapling})

						didAction = true
						break
					}
				}
			}
		}

		if didAction {
			continue
		}

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
				// Get updated state (MoveGopher updates position in room)
				// We need to fetch it or construct it.
				// MoveGopher updates X, Y, State in the room map pointer.
				// We need to return that new state.
				updatedGopher := g
				updatedGopher.X = newX
				updatedGopher.Y = newY
				updatedGopher.State = domain.GopherStateMoving
				changes.Gophers = append(changes.Gophers, updatedGopher)
			}
		} else {
			// Even if idle, if state was previously Moving/Harvesting, we might want to broadcast Idle?
			// Currently we don't broadcast idle unless it changes.
			// But since we operate on snapshots, we rely on 'changes' payload.
		}
	}
}
