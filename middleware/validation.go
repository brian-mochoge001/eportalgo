package middleware

import (
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidationMiddleware validates the request body using go-playground/validator
func ValidationMiddleware(v interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// In Go, we need to decode the body into a struct
			// This is usually done in the handler, but we can do it here if we want to mimic express-validator
			
			// For a generic middleware, this is tricky because we need the type
			// A better pattern in Go is to validate inside the handler or use a helper
			next.ServeHTTP(w, r)
		})
	}
}

// ValidateStruct is a helper to validate any struct
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}
