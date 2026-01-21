package services

import (
	"math/rand"
	"time"

	"github.com/juanpabloaj/gophercolony/internal/core/domain"
)

type MapGenerator struct {
	seed int64
}

func NewMapGenerator() *MapGenerator {
	return &MapGenerator{
		seed: time.Now().UnixNano(),
	}
}

func NewSeededMapGenerator(seed int64) *MapGenerator {
	return &MapGenerator{
		seed: seed,
	}
}

func (g *MapGenerator) Generate(width, height int) *domain.World {
	// For Phase 2, we use a simple local random.
	// In production/Phase 3+, we might use Perlin noise or seeded rand.
	rnd := rand.New(rand.NewSource(g.seed)) // nolint:gosec

	world := domain.NewWorld(width, height)

	// Simple generation: Scatter water
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := rnd.Float64()
			if r < 0.2 {
				world.Grid[y][x].Terrain = domain.TerrainWater
			} else if r < 0.25 {
				world.Grid[y][x].Terrain = domain.TerrainStone
			} else {
				world.Grid[y][x].Terrain = domain.TerrainGrass
			}
		}
	}

	return world
}
