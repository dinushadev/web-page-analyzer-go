package util

import (
	"log/slog"
	"os"
	"sync"
)

var (
	Logger   *slog.Logger
	initOnce sync.Once
)

func InitLogger() {
	Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

func EnsureLogger() *slog.Logger {
	initOnce.Do(func() {
		if Logger == nil {
			Logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
		}
	})
	return Logger
}
