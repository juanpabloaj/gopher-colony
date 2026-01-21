package domain

// TerrainType represents the type of terrain on a tile.
type TerrainType string

const (
	TerrainGrass TerrainType = "grass"
	TerrainWater TerrainType = "water"
	TerrainStone TerrainType = "stone"
)

// Tile represents a single cell in the grid.
type Tile struct {
	X, Y    int
	Terrain TerrainType
}

// World represents the game map.
type World struct {
	Width  int
	Height int
	Grid   [][]*Tile
}

// NewWorld creates a new empty world of given dimensions.
func NewWorld(width, height int) *World {
	grid := make([][]*Tile, height)
	for y := 0; y < height; y++ {
		grid[y] = make([]*Tile, width)
		for x := 0; x < width; x++ {
			grid[y][x] = &Tile{
				X:       x,
				Y:       y,
				Terrain: TerrainGrass, // Default
			}
		}
	}
	return &World{
		Width:  width,
		Height: height,
		Grid:   grid,
	}
}
