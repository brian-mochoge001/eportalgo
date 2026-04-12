package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/worker"
	"github.com/hibiken/asynq"
	"github.com/sqlc-dev/pqtype"
)

// AuditMiddleware logs data-modifying actions
func AuditMiddleware(asynqClient *asynq.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := GetUser(r.Context())
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			// Only log for specific methods that modify data
			if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" || r.Method == "DELETE" {
				var bodyBytes []byte
				if r.Body != nil {
					bodyBytes, _ = io.ReadAll(r.Body)
					// Restore the request body for the next handler
					r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}

				// Prepare for audit log
				action := r.Method + " " + r.URL.Path
				ipAddress := r.RemoteAddr
				userAgent := r.UserAgent()

				var newValue pqtype.NullRawMessage
				if len(bodyBytes) > 0 {
					newValue = pqtype.NullRawMessage{RawMessage: json.RawMessage(bodyBytes), Valid: true}
				}

				// Notify audit logger (Async via Asynq)
				payload, _ := json.Marshal(worker.AuditLogPayload{
					SchoolID:  user.SchoolID,
					UserID:    user.UserID,
					Action:    action,
					NewValue:  newValue,
					IpAddress: ipAddress,
					UserAgent: userAgent,
				})
				task := asynq.NewTask(worker.TypeAuditLog, payload)
				if _, err := asynqClient.Enqueue(task); err != nil {
					slog.Error("could not enqueue audit task", "error", err)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
