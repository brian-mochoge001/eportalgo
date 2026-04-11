package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
)

type UserContext struct {
	UserID      uuid.UUID
	SchoolID    uuid.NullUUID
	RoleID      uuid.UUID
	RoleName    string
	FirebaseUID string
	Email       string
}

// AuthMiddleware handles Firebase ID token verification and user loading from DB
func AuthMiddleware(firebaseAuth *auth.Client, queries *db.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := firebaseAuth.VerifyIDToken(r.Context(), tokenStr)
			if err != nil {
				http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
				return
			}

			firebaseUID := token.UID
			userRow, err := queries.GetUserByFirebaseUID(r.Context(), sql.NullString{String: firebaseUID, Valid: true})
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "Unauthorized: User not found in database", http.StatusUnauthorized)
					return
				}
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			userCtx := &UserContext{
				UserID:      userRow.UserID,
				SchoolID:    userRow.SchoolID,
				RoleID:      userRow.RoleID,
				RoleName:    userRow.RoleName,
				FirebaseUID: userRow.FirebaseUid.String,
				Email:       userRow.Email,
			}

			ctx := context.WithValue(r.Context(), UserKey, userCtx)
			ctx = context.WithValue(ctx, RoleKey, userRow.RoleName)
			
			// Also set school_id if it's not already set or override it from user
			if userRow.SchoolID.Valid {
				ctx = context.WithValue(ctx, SchoolIDKey, userRow.SchoolID.UUID)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Authorize middleware checks if the user has one of the allowed roles
func Authorize(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roleName, ok := r.Context().Value(RoleKey).(string)
			if !ok {
				http.Error(w, "Forbidden: User role not found", http.StatusForbidden)
				return
			}

			authorized := false
			for _, role := range allowedRoles {
				if role == roleName {
					authorized = true
					break
				}
			}

			if !authorized {
				http.Error(w, "Forbidden: Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUser retrieves the user context from the request context
func GetUser(ctx context.Context) (*UserContext, bool) {
	user, ok := ctx.Value(UserKey).(*UserContext)
	return user, ok
}

// IsAdmin checks if a role name corresponds to an administrative role
func IsAdmin(roleName string) bool {
	admins := map[string]bool{
		"Developer":               true,
		"DB Manager":              true,
		"Executive Administrator": true,
		"Academic Administrator":  true,
		"Finance Administrator":   true,
		"IT Administrator":        true,
	}
	return admins[roleName]
}
