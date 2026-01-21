package lifecycle

import (
	"log/slog"

	"github.com/juanpabloaj/gophercolony/internal/adapters/primary/ticker"
	"github.com/juanpabloaj/gophercolony/internal/core/domain"
	"github.com/juanpabloaj/gophercolony/internal/core/ports"
	"github.com/juanpabloaj/gophercolony/internal/core/services"
)

type RoomDelegate struct {
	Logger      *slog.Logger
	SimService  *services.SimulationService
	ConnService ports.ConnectionService
}

func NewRoomDelegate(logger *slog.Logger, sim *services.SimulationService, conn ports.ConnectionService) *RoomDelegate {
	return &RoomDelegate{
		Logger:      logger,
		SimService:  sim,
		ConnService: conn,
	}
}

func (d *RoomDelegate) OnRoomCreated(room *domain.Room) {
	t := ticker.NewRoomTicker(d.Logger, d.SimService, d.ConnService, room)
	go t.Start()
}
