package logger

import (
	"errors"
	"log/slog"
)

func ErrWithLogLevel(err error, level slog.Level) error {
	if err == nil {
		return nil
	}

	return errWithLogLevel{
		wrapped: err,
		level:   level,
	}
}

func getErrorLevel(err error) slog.Level {
	level := slog.LevelError

	var errWithLevel interface{ GetLogLevel() slog.Level }

	if errors.As(err, &errWithLevel) {
		level = errWithLevel.GetLogLevel()
	}

	return level
}

type errWithLogLevel struct { //nolint:errname // .
	wrapped error
	level   slog.Level
}

func (e errWithLogLevel) Error() string {
	return e.wrapped.Error()
}

func (e errWithLogLevel) Unwrap() error {
	return e.wrapped
}

func (e errWithLogLevel) GetLogLevel() slog.Level {
	return e.level
}
