package cmd

import (
	"context"
	"log/slog"
	"sync"

	"github.com/okysetiawan/retainr/config"
	"github.com/okysetiawan/retainr/ext/minio"
	"github.com/okysetiawan/retainr/ext/nats"
	"github.com/okysetiawan/retainr/internal/ports"
	"github.com/okysetiawan/retainr/internal/service/archive"
)

// TODO: refactor this

type Service interface {
	ArchiveHourly(ctx context.Context, req *archive.Request) error
}

type Archive struct {
	logger      *slog.Logger
	source      ports.ArchiveSource
	destination ports.ArchiveDestination
	jobs        []config.JobConfig
	service     Service
}

func (cmd *Archive) Run(ctx context.Context) {
	wg := new(sync.WaitGroup)
	wg.Add(len(cmd.jobs))

	for _, job := range cmd.jobs {
		go func(todo config.JobConfig) {
			defer wg.Done()
			req := &archive.Request{
				Topic:       todo.Topic,
				Name:        todo.Name,
				Source:      cmd.source,
				Destination: cmd.destination,
			}
			if err := cmd.service.ArchiveHourly(ctx, req); err != nil {
				cmd.logger.With("error", err).Error("failed to archive hourly")
			}
		}(job)
	}

	wg.Wait()
}

func NewArchive(logger *slog.Logger, configFilePath string) (*Archive, error) {
	type Config struct {
		Nats  config.NatsConfig  `mapstructure:"nats"`
		Minio config.MinioConfig `mapstructure:"minio"`
		Jobs  []config.JobConfig `mapstructure:"jobs"`
	}

	archiveConfig, err := config.LoadConfig[Config](configFilePath)
	if err != nil {
		return nil, err
	}

	source, err := nats.New(archiveConfig.Nats, logger)
	if err != nil {
		return nil, err
	}

	destination, err := minio.NewClient(archiveConfig.Minio)
	if err != nil {
		return nil, err
	}

	service, err := archive.NewArchiveService(logger)
	if err != nil {
		return nil, err
	}

	return &Archive{
		logger:      logger,
		source:      source,
		destination: destination,
		service:     service,
		jobs:        archiveConfig.Jobs,
	}, nil
}
