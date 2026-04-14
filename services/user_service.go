package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
)

type UserService struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewUserService(q *db.Queries, d *sql.DB) *UserService {
	return &UserService{Queries: q, DB: d}
}

type AddUserParams struct {
	SchoolID  uuid.UUID
	Email     string
	FirstName string
	LastName  string
	RoleName  string
}

func (s *UserService) AddUser(ctx context.Context, p AddUserParams) (db.User, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return db.User{}, err
	}
	defer tx.Rollback()
	qtx := s.Queries.WithTx(tx)

	// Find the role_id based on roleName
	role, err := qtx.GetRoleByName(ctx, p.RoleName)
	if err != nil {
		return db.User{}, fmt.Errorf("invalid role specified: %w", err)
	}

	if !role.IsSchoolRole {
		return db.User{}, fmt.Errorf("cannot add parent company role (%s) via this endpoint", p.RoleName)
	}

	// Check if user already exists in DB
	existingUser, err := qtx.GetUserByEmail(ctx, db.GetUserByEmailParams{
		Email:    p.Email,
		SchoolID: uuid.NullUUID{UUID: p.SchoolID, Valid: true},
	})
	if err == nil && existingUser.UserID != uuid.Nil {
		return db.User{}, fmt.Errorf("user with this email already exists")
	}

	// Create in local DB
	// BetterAuth handles the auth-side user creation separately
	newUser, err := qtx.CreateUser(ctx, db.CreateUserParams{
		SchoolID:    uuid.NullUUID{UUID: p.SchoolID, Valid: true},
		RoleID:      role.RoleID,
		FirstName:   p.FirstName,
		LastName:    p.LastName,
		Email:       p.Email,
		FirebaseUid: sql.NullString{Valid: false}, // No longer used
		IsActive:    true,
	})
	if err != nil {
		return db.User{}, fmt.Errorf("failed to create database user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return db.User{}, err
	}

	return newUser, nil
}

type CreateStudentProfileParams struct {
	UserID            uuid.UUID
	SchoolID          uuid.UUID
	EnrollmentNumber  string
	CurrentGradeLevel string
	AdmissionDate     time.Time
	CurrentClassID    uuid.NullUUID
}

func (s *UserService) CreateStudentProfile(ctx context.Context, p CreateStudentProfileParams) (db.StudentProfile, error) {
	// Check if user exists and is a student
	user, err := s.Queries.GetUser(ctx, db.GetUserParams{
		UserID:   p.UserID,
		SchoolID: uuid.NullUUID{UUID: p.SchoolID, Valid: true},
	})
	if err != nil {
		return db.StudentProfile{}, fmt.Errorf("user not found: %w", err)
	}

	studentRole, err := s.Queries.GetRoleByName(ctx, "Student")
	if err != nil || user.RoleID != studentRole.RoleID {
		return db.StudentProfile{}, fmt.Errorf("user must have the Student role")
	}

	// Check if profile exists
	if _, err := s.Queries.GetStudentProfileByUserID(ctx, db.GetStudentProfileByUserIDParams{
		UserID:   p.UserID,
		SchoolID: p.SchoolID,
	}); err == nil {
		return db.StudentProfile{}, fmt.Errorf("student profile already exists")
	}

	return s.Queries.CreateStudentProfile(ctx, db.CreateStudentProfileParams{
		UserID:            p.UserID,
		SchoolID:          p.SchoolID,
		EnrollmentNumber:  p.EnrollmentNumber,
		CurrentGradeLevel: sql.NullString{String: p.CurrentGradeLevel, Valid: p.CurrentGradeLevel != ""},
		AdmissionDate:     p.AdmissionDate,
		CurrentClassID:    p.CurrentClassID,
	})
}

type CreateParentProfileParams struct {
	UserID                uuid.UUID
	SchoolID              uuid.UUID
	HomeAddress           string
	Occupation            string
	EmergencyContactName  string
	EmergencyContactPhone string
}

func (s *UserService) CreateParentProfile(ctx context.Context, p CreateParentProfileParams) (db.ParentProfile, error) {
	// Check if user exists and is a parent
	user, err := s.Queries.GetUser(ctx, db.GetUserParams{
		UserID:   p.UserID,
		SchoolID: uuid.NullUUID{UUID: p.SchoolID, Valid: true},
	})
	if err != nil {
		return db.ParentProfile{}, fmt.Errorf("user not found: %w", err)
	}

	parentRole, err := s.Queries.GetRoleByName(ctx, "Parent")
	if err != nil || user.RoleID != parentRole.RoleID {
		return db.ParentProfile{}, fmt.Errorf("user must have the Parent role")
	}

	// Check if profile exists
	if _, err := s.Queries.GetParentProfileByUserID(ctx, db.GetParentProfileByUserIDParams{
		UserID:   p.UserID,
		SchoolID: p.SchoolID,
	}); err == nil {
		return db.ParentProfile{}, fmt.Errorf("parent profile already exists")
	}

	return s.Queries.CreateParentProfile(ctx, db.CreateParentProfileParams{
		UserID:                p.UserID,
		SchoolID:              p.SchoolID,
		HomeAddress:           sql.NullString{String: p.HomeAddress, Valid: p.HomeAddress != ""},
		Occupation:            sql.NullString{String: p.Occupation, Valid: p.Occupation != ""},
		EmergencyContactName:  sql.NullString{String: p.EmergencyContactName, Valid: p.EmergencyContactName != ""},
		EmergencyContactPhone: sql.NullString{String: p.EmergencyContactPhone, Valid: p.EmergencyContactPhone != ""},
	})
}
