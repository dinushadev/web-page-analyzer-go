package analyzer

import (
	"log/slog"
	"test-project-go/internal/util"
)

func logInfo(message string, args ...any) {
	util.EnsureLogger().Info(message, append(args, slog.String("component", "analyzer"))...)
}

func logError(message string, args ...any) {
	util.EnsureLogger().Error(message, append(args, slog.String("component", "analyzer"))...)
}
