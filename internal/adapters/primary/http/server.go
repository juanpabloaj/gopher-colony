package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/juanpabloaj/gophercolony/internal/core/ports"
)

type Server struct {
	server *http.Server
	logger *slog.Logger
}

func NewServer(port int, connService ports.ConnectionService, logger *slog.Logger) *Server {
	mux := http.NewServeMux()

	// Register basic health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Register WebSocket endpoint
	mux.HandleFunc("/ws", connService.HandleConnection)

	// Serve Static Files
	// We assume "web/static" is relative to the working directory where the binary is run.
	fileServer := http.FileServer(http.Dir("web/static"))
	mux.Handle("/", fileServer)

	return &Server{
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		logger: logger,
	}
}

func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server", "addr", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping HTTP server")
	return s.server.Shutdown(ctx)
}
