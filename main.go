package main

import (
	"bytes"
	"context"
	"fmt"
	"go.yaml.in/yaml/v3"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/nats-io/nats.go"
)

// ------------------ Config ------------------

type NatsConfig struct {
	Servers []string `yaml:"servers"`
}

type JobConfig struct {
	Topic      string `yaml:"topic"`
	Format     string `yaml:"format"` // "json" or "text"
	Path       string `yaml:"path"`
	BufferSize int    `yaml:"buffer_size"`
}

type Config struct {
	NATS NatsConfig  `yaml:"nats"`
	Jobs []JobConfig `yaml:"jobs"`
}

// ------------------ Writer ------------------

type HourlyWriter struct {
	job      JobConfig
	mu       sync.Mutex
	file     *os.File
	currHour string
	tmpl     *template.Template
}

type LogMessage struct {
	Time time.Time
	Body string
}

func NewHourlyWriter(job JobConfig) (*HourlyWriter, error) {
	tmpl, err := template.New("path").Parse(job.Path)
	if err != nil {
		return nil, err
	}
	return &HourlyWriter{
		job:  job,
		tmpl: tmpl,
	}, nil
}

func (w *HourlyWriter) getFilename(ts time.Time) (string, error) {
	data := map[string]string{
		"Date": ts.Format("2006-01-02"),
		"Hour": ts.Format("15"),
	}
	var buf bytes.Buffer
	if err := w.tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (w *HourlyWriter) Write(msg LogMessage) {
	hourKey := msg.Time.Format("2006-01-02_15")

	w.mu.Lock()
	defer w.mu.Unlock()

	// rotate file each hour
	if hourKey != w.currHour {
		if w.file != nil {
			_ = w.file.Close()
		}
		path, err := w.getFilename(msg.Time)
		if err != nil {
			fmt.Printf("template error: %v\n", err)
			return
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			fmt.Printf("mkdir error: %v\n", err)
			return
		}
		f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("open file error: %v\n", err)
			return
		}
		w.file = f
		w.currHour = hourKey
	}

	// choose handler based on job.Format
	var handler slog.Handler
	switch strings.ToLower(w.job.Format) {
	case "text":
		handler = slog.NewTextHandler(w.file, nil)
	default: // fallback to json
		handler = slog.NewJSONHandler(w.file, nil)
	}

	logger := slog.New(handler)
	logger.Info(msg.Body, "topic", w.job.Topic)
}

func (w *HourlyWriter) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		_ = w.file.Close()
	}
}

// ------------------ Main ------------------

func main() {
	// load config
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	// connect NATS with config servers
	nc, err := nats.Connect(strings.Join(cfg.NATS.Servers, ","))
	if err != nil {
		panic(err)
	}
	defer nc.Close()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	for _, job := range cfg.Jobs {
		writer, err := NewHourlyWriter(job)
		if err != nil {
			panic(err)
		}

		ch := make(chan LogMessage, job.BufferSize)
		wg.Add(1)
		go func(job JobConfig, writer *HourlyWriter) {
			defer wg.Done()
			defer writer.Close()
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-ch:
					writer.Write(msg)
				}
			}
		}(job, writer)

		_, err = nc.Subscribe(job.Topic, func(m *nats.Msg) {
			select {
			case ch <- LogMessage{Time: time.Now(), Body: string(m.Data)}:
			default:
				// drop if buffer full
				fmt.Printf("buffer full for topic=%s, dropping message\n", job.Topic)
			}
		})
		if err != nil {
			panic(err)
		}
		fmt.Printf("listening on NATS subject: %s (format=%s)\n", job.Topic, job.Format)
	}

	<-ctx.Done()
	wg.Wait()
	fmt.Println("shutdown gracefully")
}
