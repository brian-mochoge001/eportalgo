package middleware

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

// AuditMiddleware logs data-modifying actions
func AuditMiddleware(queries *db.Queries) func(http.Handler) http.Handler {
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

				// Execute the request first to see if it succeeds? 
				// The Node.js version calls next() after starting the async create call.
				// In Go, we'll run it in a goroutine to not block the request.
				go func(schoolID uuid.NullUUID, userID uuid.UUID, action string, newValue pqtype.NullRawMessage, ip string, ua string) {
					_, err := queries.CreateAuditLog(context.Background(), db.CreateAuditLogParams{
						SchoolID:   schoolID,
						UserID:     uuid.NullUUID{UUID: userID, Valid: true},
						Action:     action,
						EntityType: "Unknown", // Placeholder as in Node.js version
						EntityID:   uuid.NullUUID{Valid: false},
						OldValue:   pqtype.NullRawMessage{Valid: false},
						NewValue:   newValue,
						IpAddress:  sql.NullString{String: ip, Valid: true},
						UserAgent:  sql.NullString{String: ua, Valid: true},
					})
					if err != nil {
						slog.Error("Failed to create audit log", "error", err)
					}
				}(user.SchoolID, user.UserID, action, newValue, ipAddress, userAgent)
			}

			next.ServeHTTP(w, r)
		})
	}
}
