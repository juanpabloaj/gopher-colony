package domain

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
