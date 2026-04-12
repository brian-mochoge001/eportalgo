package handlers

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/google/uuid"
)

type AuthHandler struct {
	Queries      *db.Queries
	FirebaseAuth *auth.Client
}

func NewAuthHandler(q *db.Queries, fb *auth.Client) *AuthHandler {
	return &AuthHandler{Queries: q, FirebaseAuth: fb}
}

func (h *AuthHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	q := GetQueries(r.Context(), h.Queries)
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

	// Check if user exists
	existing, _ := q.GetUserByFirebaseUID(r.Context(), sql.NullString{String: req.FirebaseUID, Valid: true})
	if existing.UserID != uuid.Nil {
		middleware.SendError(w, "User already registered", http.StatusConflict, "CONFLICT", nil)
		return
	}

	// Find role
	role, err := q.GetRoleByName(r.Context(), req.RoleName)
	if err != nil {
		middleware.ValidationError(w, "Invalid role specified", err)
		return
	}

	var assignedSchoolID uuid.NullUUID
	if role.IsSchoolRole {
		if req.SchoolID == "" {
			middleware.ValidationError(w, "School ID is required for school roles", err)
			return
		}
		sid, err := uuid.Parse(req.SchoolID)
		if err != nil {
			middleware.ValidationError(w, "Invalid school ID format", err)
			return
		}
		// Verify school exists
		_, err = q.GetSchool(r.Context(), sid)
		if err != nil {
			middleware.ValidationError(w, "School not found", err)
			return
		}
		assignedSchoolID = uuid.NullUUID{UUID: sid, Valid: true}
	}

	// Create User
	user, err := q.CreateUser(r.Context(), db.CreateUserParams{
		FirebaseUid: sql.NullString{String: req.FirebaseUID, Valid: true},
		Email:       req.Email,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		RoleID:      role.RoleID,
		SchoolID:    assignedSchoolID,
		IsActive:    true,
	})
	if err != nil {
		slog.Error("Failed to register user", "error", err)
		middleware.InternalError(w, "Internal Server Error", err)
		return
	}

	// Set Firebase claims
	claims := map[string]interface{}{
		"role":     role.RoleName,
		"schoolId": req.SchoolID,
	}
	h.FirebaseAuth.SetCustomUserClaims(r.Context(), req.FirebaseUID, claims)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"user": map[string]interface{}{
			"id":           user.UserID,
			"email":        user.Email,
			"firstName":    user.FirstName,
			"lastName":     user.LastName,
			"role":         role.RoleName,
			"schoolId":     user.SchoolID,
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



