package middleware

import (
	"errors"
	"log/slog"
	"net/http"
)

type ServiceError struct {
	Code    string
	Message string
	Err     error
}

func (e *ServiceError) Error() string { return e.Message }
func (e *ServiceError) Unwrap() error { return e.Err }

func NewServiceError(code, message string) *ServiceError {
	return &ServiceError{Code: code, Message: message}
}

func WrapServiceError(code, message string, err error) *ServiceError {
	return &ServiceError{Code: code, Message: message, Err: err}
}

var statusMap = map[string]int{
	"validation_error": http.StatusBadRequest,
	"unauthenticated":  http.StatusUnauthorized,
	"forbidden":        http.StatusForbidden,
	"not_found":        http.StatusNotFound,
	"conflict":         http.StatusConflict,
	"unprocessable":    http.StatusUnprocessableEntity,
	"rate_limited":     http.StatusTooManyRequests,
	"internal_error":   http.StatusInternalServerError,
}

func HandleServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var svcErr *ServiceError
	if errors.As(err, &svcErr) {
		status, ok := statusMap[svcErr.Code]
		if !ok {
			status = http.StatusInternalServerError
		}
		RespondError(w, status, svcErr.Code, svcErr.Message)
		return
	}

	slog.Error("unexpected error", "err", err, "method", r.Method, "path", r.URL.Path)
	RespondError(w, http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
}
