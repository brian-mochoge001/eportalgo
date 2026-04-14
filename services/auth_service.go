package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
)

type AuthService struct {
	Queries *db.Queries
}

func NewAuthService(q *db.Queries) *AuthService {
	return &AuthService{Queries: q}
}

type RegisterUserRequest struct {
	Email     string
	FirstName string
	LastName  string
	RoleName  string
	SchoolID  string
}

func (s *AuthService) RegisterUser(ctx context.Context, req RegisterUserRequest) (db.User, error) {
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

	// Create User in local DB
	// BetterAuth handles the auth-side user creation and password management
	user, err := s.Queries.CreateUser(ctx, db.CreateUserParams{
		FirebaseUid: sql.NullString{Valid: false}, // No longer used
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

	return user, nil
}
