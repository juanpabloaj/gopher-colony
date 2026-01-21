package services_test

import (
	"testing"

	"github.com/juanpabloaj/gophercolony/internal/core/domain"
	"github.com/juanpabloaj/gophercolony/internal/core/services"
)

func TestMapGenerator_Generate(t *testing.T) {
	width := 20
	height := 20
	seed := int64(12345)

	// Use fixed seed for deterministic testing
	gen := services.NewMapGenerator()
	world := gen.Generate(width, height, seed)

	// 1. Verify Dimensions
	if world.Width != width {
		t.Errorf("Expected width %d, got %d", width, world.Width)
	}
	if world.Height != height {
		t.Errorf("Expected height %d, got %d", height, world.Height)
	}

	// 2. Verify Grid Integrity
	if len(world.Grid) != height {
		t.Errorf("Expected grid height %d, got %d", height, len(world.Grid))
	}
	if len(world.Grid[0]) != width {
		t.Errorf("Expected grid width %d, got %d", width, len(world.Grid[0]))
	}

	// 3. Verify Terrain Distribution (Basic check that it's not all one type)
	typeCount := make(map[domain.TerrainType]int)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			tile := world.Grid[y][x]
			// Basic coordinate check
			if tile.X != x || tile.Y != y {
				t.Errorf("Tile at [%d][%d] has malformed coordinates: %+v", x, y, tile)
			}
			typeCount[tile.Terrain]++
		}
	}

	if typeCount[domain.TerrainGrass] == 0 {
		t.Error("Expected some Grass")
	}
	if typeCount[domain.TerrainWater] == 0 {
		t.Error("Expected some Water (random chance)")
	}
}
