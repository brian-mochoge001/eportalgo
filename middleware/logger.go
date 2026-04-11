package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// responseWriter is a wrapper around http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestLogger middleware logs details of each HTTP request
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		// Get user info from context if available (added by Auth middleware)
		userID := "N/A"
		roleName := "N/A"
		if user, ok := GetUser(r.Context()); ok {
			userID = user.UserID.String()
			roleName = user.RoleName
		}

		slog.Info("HTTP Request",
			"method", r.Method,
			"url", r.URL.String(),
			"ip", r.RemoteAddr,
			"status", rw.statusCode,
			"duration", duration.String(),
			"user_agent", r.UserAgent(),
			"user_id", userID,
			"role", roleName,
		)
	})
}
