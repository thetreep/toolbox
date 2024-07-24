package logger

import (
	"errors"
	"log/slog"
)

func ErrWithAttributes(err error, attrs ...slog.Attr) error {
	if err == nil {
		return nil
	}

	return errWithAttributes{
		wrapped: err,
		attrs:   append(getAttributesFromErr(err), attrs...),
	}
}

func getAttributesFromErr(err error) []slog.Attr {
	var (
		attrs        []slog.Attr
		errWithAttrs interface{ GetSlogAttributes() []slog.Attr }
	)

	if errors.As(err, &errWithAttrs) {
		attrs = errWithAttrs.GetSlogAttributes()
	}

	return attrs
}

type errWithAttributes struct { //nolint:errname // .
	wrapped error
	attrs   []slog.Attr
}

func (e errWithAttributes) Error() string {
	return e.wrapped.Error()
}

func (e errWithAttributes) Unwrap() error {
	return e.wrapped
}

func (e errWithAttributes) GetSlogAttributes() []slog.Attr {
	return e.attrs
}
