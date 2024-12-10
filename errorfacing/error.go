package errorfacing

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/thetreep/toolbox/logger"
)

// ConstantError see https://dave.cheney.net/2016/04/07/constant-errors
// usage: `const ErrSomething = tools.ConstantError("error message")`
type ConstantError string

var _ error = ConstantError("")

func (e ConstantError) Error() string {
	return string(e)
}

type userFacingMessageError struct {
	message string
	wrapped error
}

func (e userFacingMessageError) Error() string {
	return e.wrapped.Error()
}

func (e userFacingMessageError) Unwrap() error {
	return e.wrapped
}

func WrapWithUserFacingMessage(message string, err error) error {
	if err == nil {
		return nil
	}

	return userFacingMessageError{
		message: message,
		wrapped: err,
	}
}

func GetUserFacingMessage(err error) string {
	var typedErr userFacingMessageError
	if errors.As(err, &typedErr) {
		return typedErr.message
	}

	return ""
}

type ErrorWithStatusCode interface {
	error
	GetErrorStatusCode() int
}

type statusCodeError struct {
	statusCode int
	wrapped    error
}

var _ ErrorWithStatusCode = statusCodeError{}

func (e statusCodeError) GetErrorStatusCode() int {
	return e.statusCode
}

func (e statusCodeError) Error() string {
	return e.wrapped.Error()
}

func (e statusCodeError) Unwrap() error {
	return e.wrapped
}

func BadRequestError(err error) error {
	return statusCodeError{
		statusCode: http.StatusBadRequest,
		wrapped:    logger.ErrWithLogLevel(err, slog.LevelWarn),
	}
}

func NotFoundError(err error) error {
	return statusCodeError{
		statusCode: http.StatusNotFound,
		wrapped:    logger.ErrWithLogLevel(err, slog.LevelWarn),
	}
}

func ForbiddenError(err error) error {
	return statusCodeError{
		statusCode: http.StatusForbidden,
		wrapped:    logger.ErrWithLogLevel(err, slog.LevelWarn),
	}
}

func UnauthorizedError(err error) error {
	return statusCodeError{
		statusCode: http.StatusUnauthorized,
		wrapped:    logger.ErrWithLogLevel(err, slog.LevelWarn),
	}
}

func ConflictError(err error) error {
	return statusCodeError{
		statusCode: http.StatusConflict,
		wrapped:    logger.ErrWithLogLevel(err, slog.LevelWarn),
	}
}

func GoneError(err error) error {
	return statusCodeError{
		statusCode: http.StatusGone,
		wrapped:    logger.ErrWithLogLevel(err, slog.LevelWarn),
	}
}
