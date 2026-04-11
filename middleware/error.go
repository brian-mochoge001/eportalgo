package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
)

// AppError represents a structured error in the application
type AppError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
	ErrorCode  string `json:"error_code,omitempty"`
	Internal   error  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Internal != nil {
		return fmt.Sprintf("[%s] %s: %v", e.ErrorCode, e.Message, e.Internal)
	}
	return fmt.Sprintf("[%s] %s", e.ErrorCode, e.Message)
}

type ErrorResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	ErrorCode string `json:"error_code,omitempty"`
	Stack     string `json:"stack,omitempty"`
}

// NewAppError creates a new structured error
func NewAppError(statusCode int, message string, errorCode string, internal error) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Message:    message,
		ErrorCode:  errorCode,
		Internal:   internal,
	}
}

// ErrorHandler middleware catches panics
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("Panic recovered", 
					"error", err, 
					"path", r.URL.Path,
					"stack", string(debug.Stack()))

				SendError(w, "An unexpected error occurred", http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", fmt.Errorf("%v", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// SendError helper function to send formatted errors and log internal details
func SendError(w http.ResponseWriter, message string, statusCode int, errorCode string, internal error) {
	// Log the error with internal details
	logLevel := slog.LevelError
	if statusCode < 500 {
		logLevel = slog.LevelWarn
	}

	slog.Log(context.Background(), logLevel, message, 
		"status_code", statusCode, 
		"error_code", errorCode, 
		"internal_err", internal)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	resp := ErrorResponse{
		Status:    "error",
		Message:   message,
		ErrorCode: errorCode,
	}

	// Include stack trace in non-production environments for 5xx errors
	if statusCode >= 500 && os.Getenv("NODE_ENV") != "production" {
		resp.Stack = string(debug.Stack())
	}

	json.NewEncoder(w).Encode(resp)
}

// ValidationError is a helper for 400 Bad Request
func ValidationError(w http.ResponseWriter, message string, internal error) {
	SendError(w, message, http.StatusBadRequest, "VALIDATION_ERROR", internal)
}

// UnauthorizedError is a helper for 401 Unauthorized
func UnauthorizedError(w http.ResponseWriter, message string, internal error) {
	SendError(w, message, http.StatusUnauthorized, "UNAUTHORIZED", internal)
}

// ForbiddenError is a helper for 403 Forbidden
func ForbiddenError(w http.ResponseWriter, message string, internal error) {
	SendError(w, message, http.StatusForbidden, "FORBIDDEN", internal)
}

// NotFoundError is a helper for 404 Not Found
func NotFoundError(w http.ResponseWriter, message string, internal error) {
	SendError(w, message, http.StatusNotFound, "NOT_FOUND", internal)
}

// InternalError is a helper for 500 Internal Server Error
func InternalError(w http.ResponseWriter, message string, internal error) {
	SendError(w, message, http.StatusInternalServerError, "INTERNAL_ERROR", internal)
}

// IsAppError checks if an error is of type *AppError
func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
