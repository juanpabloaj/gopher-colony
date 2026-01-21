package domain

// TerrainType represents the type of terrain on a tile.
type TerrainType int

const (
	TerrainGrass   TerrainType = iota // 0
	TerrainWater                      // 1
	TerrainStone                      // 2
	TerrainSapling                    // 3
	TerrainTree                       // 4
)

// Tile represents a single cell in the grid.
type Tile struct {
	X       int         `json:"x"`
	Y       int         `json:"y"`
	Terrain TerrainType `json:"type,omitempty"`
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

// GetTile returns the tile at the given coordinates, or nil/false if out of bounds.
func (w *World) GetTile(x, y int) (*Tile, bool) {
	if x < 0 || y < 0 || x >= w.Width || y >= w.Height {
		return nil, false
	}
	return w.Grid[y][x], true
}
