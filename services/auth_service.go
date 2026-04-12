package services

import (
	"context"
	"database/sql"
	"fmt"

	"firebase.google.com/go/v4/auth"
	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
)

type AuthService struct {
	Queries      *db.Queries
	FirebaseAuth *auth.Client
}

func NewAuthService(q *db.Queries, fb *auth.Client) *AuthService {
	return &AuthService{Queries: q, FirebaseAuth: fb}
}

type RegisterUserRequest struct {
	FirebaseUID string
	Email       string
	FirstName   string
	LastName    string
	RoleName    string
	SchoolID    string
}

func (s *AuthService) RegisterUser(ctx context.Context, req RegisterUserRequest) (db.User, error) {
	// Check if user exists
	existing, _ := s.Queries.GetUserByFirebaseUID(ctx, sql.NullString{String: req.FirebaseUID, Valid: true})
	if existing.UserID != uuid.Nil {
		return db.User{}, fmt.Errorf("user already registered")
	}

	// Find role
	role, err := s.Queries.GetRoleByName(ctx, req.RoleName)
	if err != nil {
		return db.User{}, fmt.Errorf("invalid role specified: %w", err)
	}

	var assignedSchoolID uuid.NullUUID
	if role.IsSchoolRole {
		if req.SchoolID == "" {
			return db.User{}, fmt.Errorf("school ID is required for school roles")
		}
		sid, err := uuid.Parse(req.SchoolID)
		if err != nil {
			return db.User{}, fmt.Errorf("invalid school ID format: %w", err)
		}
		// Verify school exists
		_, err = s.Queries.GetSchool(ctx, sid)
		if err != nil {
			return db.User{}, fmt.Errorf("school not found: %w", err)
		}
		assignedSchoolID = uuid.NullUUID{UUID: sid, Valid: true}
	}

	// Create User
	user, err := s.Queries.CreateUser(ctx, db.CreateUserParams{
		FirebaseUid: sql.NullString{String: req.FirebaseUID, Valid: true},
		Email:       req.Email,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		RoleID:      role.RoleID,
		SchoolID:    assignedSchoolID,
		IsActive:    true,
	})
	if err != nil {
		return db.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	// Set Firebase claims
	claims := map[string]interface{}{
		"role":     role.RoleName,
		"schoolId": req.SchoolID,
	}
	err = s.FirebaseAuth.SetCustomUserClaims(ctx, req.FirebaseUID, claims)
	if err != nil {
		// Log but don't fail? Or fail? Let's return error but user is created.
		return user, fmt.Errorf("failed to set firebase claims: %w", err)
	}

	return user, nil
}
