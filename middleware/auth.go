package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type UserContext struct {
	UserID   uuid.UUID
	SchoolID uuid.NullUUID
	RoleID   uuid.UUID
	RoleName string
	Email    string
}

// JWKS key set cache
type jwksCache struct {
	mu       sync.RWMutex
	keys     map[string]interface{} // kid -> public key
	url      string
	lastFetch time.Time
	ttl      time.Duration
}

var globalJWKS *jwksCache

func initJWKSCache(jwksURL string) {
	globalJWKS = &jwksCache{
		url:  jwksURL,
		keys: make(map[string]interface{}),
		ttl:  10 * time.Minute,
	}
}

func (c *jwksCache) getKey(kid string) (interface{}, error) {
	c.mu.RLock()
	if key, ok := c.keys[kid]; ok && time.Since(c.lastFetch) < c.ttl {
		c.mu.RUnlock()
		return key, nil
	}
	c.mu.RUnlock()

	// Fetch fresh keys
	return c.fetchAndGetKey(kid)
}

func (c *jwksCache) fetchAndGetKey(kid string) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check after lock
	if key, ok := c.keys[kid]; ok && time.Since(c.lastFetch) < c.ttl {
		return key, nil
	}

	resp, err := http.Get(c.url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read JWKS response: %w", err)
	}

	var jwks struct {
		Keys []json.RawMessage `json:"keys"`
	}
	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, fmt.Errorf("failed to parse JWKS: %w", err)
	}

	// Parse each key
	newKeys := make(map[string]interface{})
	for _, rawKey := range jwks.Keys {
		var keyMeta struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			Alg string `json:"alg"`
			Crv string `json:"crv"`
		}
		if err := json.Unmarshal(rawKey, &keyMeta); err != nil {
			continue
		}

		// Parse the key using golang-jwt compatible format
		// For EdDSA (OKP) keys, we parse directly
		// For RSA/EC keys, use standard parsers
		pubKey, parseErr := parseJWKPublicKey(rawKey, keyMeta.Kty)
		if parseErr != nil {
			log.Printf("Warning: failed to parse JWK kid=%s: %v", keyMeta.Kid, parseErr)
			continue
		}
		newKeys[keyMeta.Kid] = pubKey
	}

	c.keys = newKeys
	c.lastFetch = time.Now()

	if key, ok := newKeys[kid]; ok {
		return key, nil
	}

	// If no kid match, return first available key (BetterAuth may not set kid in token)
	for _, key := range newKeys {
		return key, nil
	}

	return nil, fmt.Errorf("no matching key found for kid: %s", kid)
}

// AuthMiddleware handles JWT token verification via JWKS and user loading from DB
func AuthMiddleware(jwksURL string, queries *db.Queries) func(http.Handler) http.Handler {
	initJWKSCache(jwksURL)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

			// Parse and validate JWT
			token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
				kid, _ := token.Header["kid"].(string)
				return globalJWKS.getKey(kid)
			})
			if err != nil || !token.Valid {
				http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "Unauthorized: Invalid claims", http.StatusUnauthorized)
				return
			}

			// Extract user info from JWT claims
			sub, _ := claims["sub"].(string)
			email, _ := claims["email"].(string)
			roleName, _ := claims["role"].(string)

			if sub == "" {
				http.Error(w, "Unauthorized: No user ID in token", http.StatusUnauthorized)
				return
			}

			// Look up user in our database by email (BetterAuth user IDs may differ from our UUIDs)
			userRow, err := queries.GetUserByEmailOnly(r.Context(), email)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "Unauthorized: User not found in database", http.StatusUnauthorized)
					return
				}
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// If role from JWT is empty, get it from the DB
			if roleName == "" {
				roleName = userRow.RoleName
			}

			userCtx := &UserContext{
				UserID:   userRow.UserID,
				SchoolID: userRow.SchoolID,
				RoleID:   userRow.RoleID,
				RoleName: roleName,
				Email:    userRow.Email,
			}

			ctx := context.WithValue(r.Context(), UserKey, userCtx)
			ctx = context.WithValue(ctx, RoleKey, roleName)

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
