package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type SchoolService struct {
	Queries      *db.Queries
	FirebaseAuth *auth.Client
	Redis        *redis.Client
}

func NewSchoolService(q *db.Queries, fb *auth.Client, r *redis.Client) *SchoolService {
	return &SchoolService{Queries: q, FirebaseAuth: fb, Redis: r}
}

func generateSchoolInitial(name string) string {
	words := strings.Fields(name)
	var initial string
	if len(words) > 1 {
		for _, word := range words {
			if len(word) > 0 {
				initial += strings.ToUpper(string(word[0]))
			}
		}
	} else if len(name) >= 3 {
		initial = strings.ToUpper(name[:3])
	} else {
		initial = strings.ToUpper(name)
	}

	b := make([]byte, 2)
	rand.Read(b)
	return initial + strings.ToUpper(hex.EncodeToString(b))
}

type RegisterSchoolRequest struct {
	SchoolName       string
	Subdomain        string
	Address          string
	PhoneNumber      string
	Email            string
	AdminFirstName   string
	AdminLastName    string
	AdminEmail       string
	AdminFirebaseUid string
	AdminRoleName    string
}

type RegisterSchoolResponse struct {
	School    db.School
	AdminUser db.User
	RoleName  string
}

func (s *SchoolService) RegisterSchool(ctx context.Context, req RegisterSchoolRequest) (RegisterSchoolResponse, error) {
	if req.AdminRoleName == "" {
		req.AdminRoleName = "Executive Administrator"
	}

	// Check if school exists
	existing, _ := s.Queries.GetSchoolByNameOrSubdomain(ctx, db.GetSchoolByNameOrSubdomainParams{
		SchoolName: req.SchoolName,
		Subdomain:  sql.NullString{String: req.Subdomain, Valid: req.Subdomain != ""},
	})
	if existing.SchoolID != uuid.Nil {
		return RegisterSchoolResponse{}, fmt.Errorf("school with this name or subdomain already exists")
	}

	// Generate unique initial
	schoolInitial := generateSchoolInitial(req.SchoolName)
	for {
		exists, _ := s.Queries.GetSchoolByInitial(ctx, sql.NullString{String: schoolInitial, Valid: true})
		if exists.SchoolID == uuid.Nil {
			break
		}
		schoolInitial = generateSchoolInitial(req.SchoolName)
	}

	// Create School
	school, err := s.Queries.CreateSchool(ctx, db.CreateSchoolParams{
		SchoolName:    req.SchoolName,
		Subdomain:     sql.NullString{String: req.Subdomain, Valid: req.Subdomain != ""},
		Status:        "pending",
		SchoolInitial: sql.NullString{String: schoolInitial, Valid: true},
		Address:       sql.NullString{String: req.Address, Valid: req.Address != ""},
		PhoneNumber:   sql.NullString{String: req.PhoneNumber, Valid: req.PhoneNumber != ""},
		Email:         sql.NullString{String: req.Email, Valid: req.Email != ""},
	})
	if err != nil {
		return RegisterSchoolResponse{}, fmt.Errorf("failed to create school: %w", err)
	}

	// Find Role
	role, err := s.Queries.GetRoleByName(ctx, req.AdminRoleName)
	if err != nil {
		return RegisterSchoolResponse{}, fmt.Errorf("role '%s' not found: %w", req.AdminRoleName, err)
	}

	// Create Admin User
	adminUser, err := s.Queries.CreateUser(ctx, db.CreateUserParams{
		SchoolID:    uuid.NullUUID{UUID: school.SchoolID, Valid: true},
		RoleID:      role.RoleID,
		FirstName:   req.AdminFirstName,
		LastName:    req.AdminLastName,
		Email:       req.AdminEmail,
		FirebaseUid: sql.NullString{String: req.AdminFirebaseUid, Valid: true},
		IsActive:    true,
	})
	if err != nil {
		return RegisterSchoolResponse{}, fmt.Errorf("failed to create admin user: %w", err)
	}

	// Create Default School Settings
	_, _ = s.Queries.CreateSchoolSetting(ctx, school.SchoolID)

	// Set Firebase Custom Claims
	claims := map[string]interface{}{
		"role":         role.RoleName,
		"schoolId":     school.SchoolID.String(),
		"schoolStatus": school.Status,
	}
	_ = s.FirebaseAuth.SetCustomUserClaims(ctx, req.AdminFirebaseUid, claims)

	return RegisterSchoolResponse{
		School:    school,
		AdminUser: adminUser,
		RoleName:  role.RoleName,
	}, nil
}

func (s *SchoolService) VerifySchool(ctx context.Context, schoolID uuid.UUID, status string) (db.School, error) {
	school, err := s.Queries.GetSchoolWithAdmin(ctx, schoolID)
	if err != nil {
		return db.School{}, err
	}

	updatedSchool, err := s.Queries.UpdateSchoolStatus(ctx, db.UpdateSchoolStatusParams{
		SchoolID: schoolID,
		Status:   status,
	})
	if err != nil {
		return db.School{}, err
	}

	// Update Firebase claims for admin
	if school.AdminFirebaseUid.Valid {
		claims := map[string]interface{}{
			"schoolStatus": updatedSchool.Status,
			"schoolId":     updatedSchool.SchoolID.String(),
			"role":         school.AdminRoleName,
		}
		_ = s.FirebaseAuth.SetCustomUserClaims(ctx, school.AdminFirebaseUid.String, claims)
	}

	return updatedSchool, nil
}
