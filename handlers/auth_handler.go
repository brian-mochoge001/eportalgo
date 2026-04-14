package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/brian-mochoge001/eportalgo/services"
)

type AuthHandler struct {
	Queries     *db.Queries
	AuthService *services.AuthService
}

func NewAuthHandler(q *db.Queries, s *services.AuthService) *AuthHandler {
	return &AuthHandler{Queries: q, AuthService: s}
}

// RegisterUser creates a local DB user record after BetterAuth has created the auth-side user.
// The mobile app calls this endpoint after successful BetterAuth signup.
func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		RoleName  string `json:"roleName"`
		SchoolID  string `json:"schoolId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	user, err := h.AuthService.RegisterUser(r.Context(), services.RegisterUserRequest{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		RoleName:  req.RoleName,
		SchoolID:  req.SchoolID,
	})

	if err != nil {
		if err.Error() == "user already registered" {
			middleware.SendError(w, "User already registered", http.StatusConflict, "CONFLICT", nil)
			return
		}
		slog.Error("Failed to register user", "error", err)
		middleware.InternalError(w, "Internal Server Error", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"user": map[string]interface{}{
			"id":        user.UserID,
			"email":     user.Email,
			"firstName": user.FirstName,
			"lastName":  user.LastName,
		},
	})
}

// LoginUser is a no-op since BetterAuth handles authentication.
// This endpoint exists for backwards compatibility — it returns the user profile
// for a valid JWT token (the token is already verified by AuthMiddleware).
func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUser(r.Context())
	if !ok {
		middleware.UnauthorizedError(w, "Not authenticated", nil)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"user": map[string]interface{}{
			"id":       userCtx.UserID,
			"email":    userCtx.Email,
			"role":     userCtx.RoleName,
			"schoolId": userCtx.SchoolID,
		},
	})
}

// GetMe returns the current authenticated user's profile and role.
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := middleware.GetUser(r.Context())
	if !ok {
		middleware.UnauthorizedError(w, "Not authenticated", nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": map[string]interface{}{
			"id":       userCtx.UserID,
			"email":    userCtx.Email,
			"role":     userCtx.RoleName,
			"schoolId": userCtx.SchoolID,
			"roleId":   userCtx.RoleID,
		},
	})
}
