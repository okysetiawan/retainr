package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/okysetiawan/retainr/cmd"
	"github.com/okysetiawan/retainr/internal/logger"
)

func main() {
	// TODO: refactor to accept multiple command
	logger := logger.NewLogger()
	arhiver, err := cmd.NewArchive(logger, "config.yaml")
	if err != nil {
		logger.With("error", err).Error("Failed to create archiver")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	arhiver.Run(ctx)
	logger.Info("Shutting down...")
}
