package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/brian-mochoge001/eportalgo/services"
)

type AuthHandler struct {
	Queries      *db.Queries
	AuthService  *services.AuthService
	FirebaseAuth *auth.Client
}

func NewAuthHandler(q *db.Queries, s *services.AuthService, fb *auth.Client) *AuthHandler {
	return &AuthHandler{Queries: q, AuthService: s, FirebaseAuth: fb}
}

func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FirebaseUID string `json:"firebase_uid"`
		Email       string `json:"email"`
		FirstName   string `json:"firstName"`
		LastName    string `json:"lastName"`
		RoleName    string `json:"roleName"`
		SchoolID    string `json:"schoolId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	user, err := h.AuthService.RegisterUser(r.Context(), services.RegisterUserRequest{
		FirebaseUID: req.FirebaseUID,
		Email:       req.Email,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		RoleName:    req.RoleName,
		SchoolID:    req.SchoolID,
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
			"id":           user.UserID,
			"email":        user.Email,
			"firstName":    user.FirstName,
			"lastName":     user.LastName,
			"firebase_uid": user.FirebaseUid.String,
		},
	})
}

func (h *AuthHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	q := GetQueries(r.Context(), h.Queries)
	var req struct {
		IDToken string `json:"idToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	token, err := h.FirebaseAuth.VerifyIDToken(r.Context(), req.IDToken)
	if err != nil {
		middleware.UnauthorizedError(w, "Authentication failed", err)
		return
	}

	userRow, err := q.GetUserByFirebaseUID(r.Context(), sql.NullString{String: token.UID, Valid: true})
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.NotFoundError(w, "User not found in database", err)
			return
		}
		middleware.InternalError(w, "Internal Server Error", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"user": map[string]interface{}{
			"id":           userRow.UserID,
			"email":        userRow.Email,
			"firstName":    userRow.FirstName,
			"lastName":     userRow.LastName,
			"role":         userRow.RoleName,
			"schoolId":     userRow.SchoolID,
			"firebase_uid": userRow.FirebaseUid.String,
		},
	})
}



