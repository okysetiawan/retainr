package archive

import (
	"context"
	"log/slog"
	"time"

	"github.com/okysetiawan/retainr/internal/ports"
)

type Message struct {
	Msg       []byte
	Timestamp time.Time
}

type Service struct {
	logger *slog.Logger
}

type Request struct {
	Topic       string
	Name        string
	Source      ports.ArchiveSource
	Destination ports.ArchiveDestination
}

func (s *Service) ArchiveHourly(ctx context.Context, req *Request) error {
	worker := NewWorker(s.logger, req.Name)
	worker.Start(ctx)
	defer worker.Stop()

	callback := func(msg []byte) {
		message := Message{Msg: msg, Timestamp: time.Now()}
		worker.Enqueue(message)
	}
	if err := req.Source.Subscribe(req.Topic, callback); err != nil {
		return err
	}

	done := make(chan bool)

	go func() {
		uploadContext := context.Background()
		for file := range worker.Listen() {
			if err := req.Destination.PutReader(uploadContext, file.Path(), file); err != nil {
				s.logger.With("error", err, "path", file.Path()).Warn("failed to put reader")
			}
			_ = file.Close()
		}
		close(done)
	}()

	worker.Stop()
	<-done

	return nil
}

func NewArchiveService(logger *slog.Logger) (*Service, error) {
	return &Service{
		logger: logger,
	}, nil
}
