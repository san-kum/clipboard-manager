package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/san-kum/clipboard-manager/internal/clipboard"
	"github.com/san-kum/clipboard-manager/internal/config"
)

func main() {
	// Setup logging
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg := config.DefaultConfig()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Create clipboard store
	store := clipboard.NewClipboardStore(
		cfg.StoragePath,
		cfg.MaxEntries,
		log,
	)

	// Load existing history
	if err := store.Load(); err != nil {
		log.Errorf("Failed to load clipboard history: %v", err)
	}

	// Create clipboard manager
	clipboardManager := clipboard.NewClipboardManager(store, log)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start clipboard listener
	if err := clipboardManager.Start(ctx); err != nil {
		log.Fatalf("Failed to start clipboard listener: %v", err)
	}

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Info("Received shutdown signal, stopping clipboard manager...")
	clipboardManager.Stop()
}
