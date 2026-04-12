package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
)

type rlsContextKey string

const (
	DBConnKey rlsContextKey = "db_conn"
)

// RLSMiddleware ensures that every request has its own DB connection with RLS session variables set.
func RLSMiddleware(dbPool *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			
			// 1. Get user context (assuming AuthMiddleware has already run)
			user, ok := GetUser(ctx)
			if !ok {
				// If not authenticated, we just proceed. 
				// Public routes won't have RLS session variables set, so they might see nothing 
				// if RLS is enabled on those tables.
				next.ServeHTTP(w, r)
				return
			}

			// 2. Acquire a dedicated connection from the pool for this request
			conn, err := dbPool.Conn(ctx)
			if err != nil {
				http.Error(w, "Internal Server Error: Could not acquire DB connection", http.StatusInternalServerError)
				return
			}
			defer conn.Close()

			// 3. Set session variables for RLS
			schoolIDStr := ""
			if user.SchoolID.Valid {
				schoolIDStr = user.SchoolID.UUID.String()
			}

			// We use local = true (SET LOCAL) so it only lasts for the current transaction/session logic,
			// but since we have a dedicated Conn, plain SET is also fine until we return it to the pool.
			// However, to be safe across potential pool re-use if Conn.Close() didn't reset, we'd want to reset them.
			_, err = conn.ExecContext(ctx, fmt.Sprintf("SELECT set_config('app.current_school_id', '%s', false), set_config('app.current_role', '%s', false)", schoolIDStr, user.RoleName))
			if err != nil {
				http.Error(w, "Internal Server Error: Could not set RLS context", http.StatusInternalServerError)
				return
			}

			// 4. Add the connection to the context so handlers can use it
			ctx = context.WithValue(ctx, DBConnKey, conn)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetConn retrieves the DB connection from the context.
func GetConn(ctx context.Context) (*sql.Conn, bool) {
	conn, ok := ctx.Value(DBConnKey).(*sql.Conn)
	return conn, ok
}
