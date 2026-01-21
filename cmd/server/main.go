package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/juanpabloaj/gophercolony/internal/adapters/primary/http"
	"github.com/juanpabloaj/gophercolony/internal/adapters/secondary/memsockets"
	"github.com/juanpabloaj/gophercolony/internal/core/services"
)

func main() {
	// 1. Setup Logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	// 2. Setup Services (Core)
	mapGenerator := services.NewMapGenerator()
	roomRepo := memsockets.NewRoomManager(mapGenerator)

	// Inject Repo into ConnectionManager
	connManager := services.NewConnectionManager(logger, roomRepo)

	// 3. Setup Adapters (Primary)
	srv := http.NewServer(8080, connManager, logger)

	// 4. Start Server
	go func() {
		if err := srv.Start(); err != nil {
			logger.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	// 5. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exiting")
}
