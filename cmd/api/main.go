package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/saimonsiddique/blog-api/internal/app"
	"github.com/saimonsiddique/blog-api/internal/config"
)

const shutdownTimeout = 30 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Initialize application
	application, err := app.New(cfg)
	if err != nil {
		return err
	}
	defer application.Close()

	// Create context for interrupt signals
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("Server starting on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := application.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErrors:
		return err
	case <-ctx.Done():
		log.Println("Shutdown signal received")
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	log.Println("Shutting down gracefully...")
	if err := application.Shutdown(shutdownCtx); err != nil {
		return err
	}

	log.Println("Shutdown completed")
	return nil
}
