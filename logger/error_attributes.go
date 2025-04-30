package logger

import (
	"errors"
)

func ErrWithAttributes(err error, attrs ...any) error {
	if err == nil {
		return nil
	}

	return errWithAttributes{
		wrapped: err,
		attrs:   append(getAttributesFromErr(err), attrs...),
	}
}

func getAttributesFromErr(err error) []any {
	var (
		attrs        []any
		errWithAttrs interface{ GetSlogAttributes() []any }
	)

	if errors.As(err, &errWithAttrs) {
		attrs = errWithAttrs.GetSlogAttributes()
	}

	return attrs
}

type errWithAttributes struct { //nolint:errname // .
	wrapped error
	attrs   []any
}

func (e errWithAttributes) Error() string {
	return e.wrapped.Error()
}

func (e errWithAttributes) Unwrap() error {
	return e.wrapped
}

func (e errWithAttributes) GetSlogAttributes() []any {
	return e.attrs
}
