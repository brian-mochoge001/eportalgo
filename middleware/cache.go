package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// responseBuffer captures the response body for caching
type responseBuffer struct {
	http.ResponseWriter
	body []byte
}

func (rb *responseBuffer) Write(b []byte) (int, error) {
	rb.body = append(rb.body, b...)
	return rb.ResponseWriter.Write(b)
}

// CacheMiddleware caches responses in Redis
func CacheMiddleware(redisClient *redis.Client, duration time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				next.ServeHTTP(w, r)
				return
			}

			key := r.URL.String()
			cachedBody, err := redisClient.Get(r.Context(), key).Result()
			if err == nil {
				slog.Info("Cache hit", "url", key)
				w.Header().Set("Content-Type", "application/json")
				w.Write([]byte(cachedBody))
				return
			}

			if err != redis.Nil {
				slog.Error("Redis cache error", "error", err)
			}

			slog.Info("Cache miss", "url", key)
			rb := &responseBuffer{ResponseWriter: w}
			next.ServeHTTP(rb, r)

			if len(rb.body) > 0 {
				err := redisClient.Set(r.Context(), key, rb.body, duration).Err()
				if err != nil {
					slog.Error("Failed to cache response", "error", err)
				}
			}
		})
	}
}
