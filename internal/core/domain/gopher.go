package domain

// GopherState represents the current activity of a Gopher.
type GopherState int

const (
	GopherStateIdle       GopherState = iota // 0
	GopherStateMoving                        // 1
	GopherStateWorking                       // 2
	GopherStateHarvesting                    // 3
)

// Inventory holds resources carried by the Gopher.
type Inventory struct {
	Wood int `json:"wood"`
	Food int `json:"food"`
}

// Gopher represents an autonomous agent in the colony.
type Gopher struct {
	ID        string      `json:"id"`
	X         int         `json:"x"`
	Y         int         `json:"y"`
	State     GopherState `json:"state"`
	Inventory Inventory   `json:"inventory"`
}
