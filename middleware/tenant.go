package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const (
	SchoolIDKey contextKey = "school_id"
	UserKey     contextKey = "user"
	RoleKey     contextKey = "role"
)

// TenantMiddleware extracts the school_id from headers and adds it to the request context.
// In a real B2B SaaS, you might extract this from a JWT claim or a subdomain.
func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantIDStr := r.Header.Get("X-Tenant-ID") // Use X-Tenant-ID to match Node.js version
		if tenantIDStr == "" {
			// Fallback to X-School-ID if needed
			tenantIDStr = r.Header.Get("X-School-ID")
		}

		if tenantIDStr != "" {
			schoolID, err := uuid.Parse(tenantIDStr)
			if err == nil {
				// Add school_id to context
				ctx := context.WithValue(r.Context(), SchoolIDKey, schoolID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// GetSchoolID retrieves the school_id from the context.
func GetSchoolID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(SchoolIDKey).(uuid.UUID)
	return id, ok
}
