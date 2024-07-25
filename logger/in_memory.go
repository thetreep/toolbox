package logger

import (
	"context"
	"log/slog"
	"slices"
	"sync"
)

func InMemoryHandler(level slog.Level) *inMemoryHandler {
	return &inMemoryHandler{
		level: level,
	}
}

type inMemoryHandler struct {
	level slog.Level
	logs  []slog.Record
	lock  sync.Mutex
}

func (h *inMemoryHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *inMemoryHandler) Handle(_ context.Context, record slog.Record) error {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.logs = append(h.logs, record)
	return nil
}

func (h *inMemoryHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return h // FIXME
}

func (h *inMemoryHandler) WithGroup(_ string) slog.Handler {
	return h // FIXME
}

func (h *inMemoryHandler) GetLogs() []slog.Record {
	return slices.Clone(h.logs)
}
