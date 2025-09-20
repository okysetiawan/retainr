package logger

import (
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

// TODO: refactor

func NewLogger() *slog.Logger {
	level := slog.LevelInfo
	lbj := &lumberjack.Logger{
		Filename:   "application.log",
		MaxSize:    500,
		MaxAge:     7,
		MaxBackups: 5,
		LocalTime:  true,
		Compress:   true,
	}
	multiWriter := io.MultiWriter(os.Stdout, lbj)

	return slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{Level: level}))
}
