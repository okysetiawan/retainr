package archive

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/okysetiawan/retainr/internal/fs"
)

type WorkerHourly struct {
	logger     *slog.Logger
	msgCh      chan Message
	mu         sync.Locker
	prefixFile string

	running       atomic.Bool
	currentBucket atomic.Pointer[string]
	currentFile   atomic.Pointer[fs.FileSystem]

	filesystems chan *fs.FileSystem
}

func (w *WorkerHourly) Start(ctx context.Context) {
	if w.running.Load() {
		w.logger.Info("[worker] already started")
		return
	}
	w.logger.Info("[worker] started")
	w.running.Store(true)
	go func() {
		select {
		case msg := <-w.msgCh:
			w.send(msg)
		case <-ctx.Done():
			w.logger.With("error", ctx.Err()).Info("[worker] context done")
			w.Stop()
		default:

		}
	}()
}

func (w *WorkerHourly) Listen() chan *fs.FileSystem {
	return w.filesystems
}

func (w *WorkerHourly) relayFile(fs *fs.FileSystem) {
	if !w.running.Load() {
		return
	}

	if fs == nil {
		return
	}

	w.logger.With("file", fs.File.Name()).Info("[worker] relaying file has finished")
	w.filesystems <- fs
}

func (w *WorkerHourly) RelayFile(timestamp time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()
	newBucket := timestamp.UTC().Format("2006-01-02-15")
	if w.currentBucket.Load() == nil {
		w.currentBucket.Store(&newBucket)
	}
	currBucket := *w.currentBucket.Load()
	if currBucket == newBucket && w.currentFile.Load() != nil {
		return
	}

	w.relayFile(w.currentFile.Load())

	// create new file and store
	date := timestamp.Format(time.DateOnly)
	hour := timestamp.Hour()
	path := fmt.Sprintf("%s/%s/%d_archive.txt", w.prefixFile, date, hour)
	filesystem, err := fs.New(path)
	if err != nil {
		w.logger.With("error", err).Error("[worker] unable to create filesystem")
	} else {
		w.currentFile.Store(filesystem)
	}
}

func (w *WorkerHourly) send(msg Message) {
	w.RelayFile(msg.Timestamp)
	writer := w.currentFile.Load()
	msgToBeWrite := fmt.Sprintf("%s - %s\n", msg.Timestamp.Format(time.RFC3339Nano), msg.Msg)
	if _, err := writer.WriteString(msgToBeWrite); err != nil {
		w.logger.With("error", err).Error("[worker] unable to write to file")
	}
}

func (w *WorkerHourly) Stop() {
	if !w.running.Load() {
		return
	}

	w.logger.Info("[worker] stopping...")
	w.running.Store(false)

	w.mu.Lock()
	defer w.mu.Unlock()
	w.relayFile(w.currentFile.Load()) // make sure all file is stored properly
	close(w.msgCh)
	close(w.filesystems)
	w.logger.Info("[worker] stopped")
}

func (w *WorkerHourly) Enqueue(msg Message) {
	w.msgCh <- msg
}

func NewWorker(logger *slog.Logger, prefix string) *WorkerHourly {
	return &WorkerHourly{
		logger:        logger,
		msgCh:         make(chan Message, 1000),
		running:       atomic.Bool{},
		mu:            new(sync.Mutex),
		prefixFile:    prefix,
		currentBucket: atomic.Pointer[string]{},
		currentFile:   atomic.Pointer[fs.FileSystem]{},
		filesystems:   make(chan *fs.FileSystem, 10),
	}
}
