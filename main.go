package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"silentrig/internal/api"
	"silentrig/internal/config"
	"silentrig/internal/database"
	"silentrig/internal/logger"
	"silentrig/internal/registry"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// Initialize logger
	log := logger.New()
	log.Info("Starting SilentRig server", "address", cfg.Server.Address, "port", cfg.Server.Port)

	// Initialize database
	db, err := database.New(cfg.Database.Path)
	if err != nil {
		log.Fatal("Failed to initialize database", "error", err)
	}
	defer db.Close()

	// Initialize registry
	reg := registry.New(db, log)

	// Initialize API server
	server := api.New(cfg, reg, log)

	// Start server in background
	go func() {
		if err := server.Start(); err != nil {
			log.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Error during shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("Server shutdown complete")
} 