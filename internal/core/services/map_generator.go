package services

import (
	"math/rand"

	"github.com/juanpabloaj/gophercolony/internal/core/domain"
)

type MapGenerator struct {
}

func NewMapGenerator() *MapGenerator {
	return &MapGenerator{}
}

func (g *MapGenerator) Generate(width, height int, seed int64) *domain.World {
	// For Phase 2, we use a simple local random.
	// In production/Phase 3+, we might use Perlin noise or seeded rand.
	rnd := rand.New(rand.NewSource(seed)) // nolint:gosec

	world := domain.NewWorld(width, height)

	// Simple generation: Scatter water
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Perlin noise or simple random for now
			// We already use seeded randomness via rnd (type *rand.Rand)

			val := rnd.Float64()
			var tType domain.TerrainType

			if val < 0.1 {
				tType = domain.TerrainWater
			} else if val < 0.15 {
				tType = domain.TerrainStone
			} else if val < 0.20 {
				// 5% Chance for Sapling
				tType = domain.TerrainSapling
			} else {
				tType = domain.TerrainGrass
			}

			// Our domain.Tile struct has Terrain TerrainType.
			world.Grid[y][x].Terrain = tType
		}
	}

	// Phase 5 Stage 3: Place a Storage Chest at the center
	cx, cy := width/2, height/2
	if cx >= 0 && cx < width && cy >= 0 && cy < height {
		world.Grid[cy][cx].Terrain = domain.TerrainChest
	}

	return world
}
