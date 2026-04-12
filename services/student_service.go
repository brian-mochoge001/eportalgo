package services

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type StudentService struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewStudentService(q *db.Queries, d *sql.DB) *StudentService {
	return &StudentService{Queries: q, DB: d}
}

type OnboardStudentRequest struct {
	SchoolID          uuid.UUID
	StudentFirstName  string
	StudentLastName   string
	StudentDob        time.Time
	StudentGender     string
	ParentFirstName   string
	ParentLastName    string
	ParentEmail       string
	ParentPhoneNumber string
	ClassID           uuid.UUID
	EnrollmentDate    time.Time
}

func (s *StudentService) OnboardStudent(ctx context.Context, req OnboardStudentRequest) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	qtx := s.Queries.WithTx(tx)

	// Roles
	studentRole, err := qtx.GetRoleByName(ctx, "Student")
	if err != nil {
		return fmt.Errorf("student role not found: %w", err)
	}
	parentRole, err := qtx.GetRoleByName(ctx, "Parent")
	if err != nil {
		return fmt.Errorf("parent role not found: %w", err)
	}

	// Create Student
	studentEmail := fmt.Sprintf("%s.%s@student.edu", strings.ToLower(req.StudentFirstName), strings.ToLower(req.StudentLastName))
	pass, _ := bcrypt.GenerateFromPassword([]byte("password123"), 10)

	studentUser, err := qtx.CreateUser(ctx, db.CreateUserParams{
		SchoolID:     uuid.NullUUID{UUID: req.SchoolID, Valid: true},
		RoleID:       studentRole.RoleID,
		FirstName:    req.StudentFirstName,
		LastName:     req.StudentLastName,
		Email:        studentEmail,
		PasswordHash: sql.NullString{String: string(pass), Valid: true},
		DateOfBirth:  sql.NullTime{Time: req.StudentDob, Valid: true},
		Gender:       sql.NullString{String: req.StudentGender, Valid: true},
		IsActive:     true,
	})
	if err != nil {
		return fmt.Errorf("failed to create student user: %w", err)
	}

	_, err = qtx.CreateStudentProfile(ctx, db.CreateStudentProfileParams{
		UserID:           studentUser.UserID,
		SchoolID:         req.SchoolID,
		EnrollmentNumber: fmt.Sprintf("ENR-%d", time.Now().Unix()),
		AdmissionDate:    req.EnrollmentDate,
		CurrentClassID:   uuid.NullUUID{UUID: req.ClassID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("failed to create student profile: %w", err)
	}

	// Parent
	parentUser, err := qtx.GetUserByEmail(ctx, db.GetUserByEmailParams{
		Email:    req.ParentEmail,
		SchoolID: uuid.NullUUID{UUID: req.SchoolID, Valid: true},
	})
	if err != nil {
		if err == sql.ErrNoRows {
			parentUser, err = qtx.CreateUser(ctx, db.CreateUserParams{
				SchoolID:     uuid.NullUUID{UUID: req.SchoolID, Valid: true},
				RoleID:       parentRole.RoleID,
				FirstName:    req.ParentFirstName,
				LastName:     req.ParentLastName,
				Email:        req.ParentEmail,
				PasswordHash: sql.NullString{String: string(pass), Valid: true},
				PhoneNumber:  sql.NullString{String: req.ParentPhoneNumber, Valid: true},
				IsActive:     true,
			})
			if err != nil {
				return fmt.Errorf("failed to create parent user: %w", err)
			}
			_, err = qtx.CreateParentProfile(ctx, db.CreateParentProfileParams{
				UserID:   parentUser.UserID,
				SchoolID: req.SchoolID,
			})
			if err != nil {
				return fmt.Errorf("failed to create parent profile: %w", err)
			}
		} else {
			return fmt.Errorf("failed to check parent user: %w", err)
		}
	}

	_, err = qtx.CreateEnrollment(ctx, db.CreateEnrollmentParams{
		SchoolID:       req.SchoolID,
		StudentID:      studentUser.UserID,
		ClassID:        req.ClassID,
		EnrollmentDate: req.EnrollmentDate,
		Status:         "Enrolled",
	})
	if err != nil {
		return fmt.Errorf("failed to create enrollment: %w", err)
	}

	return tx.Commit()
}

func (s *StudentService) InitiateTransfer(ctx context.Context, studentID, sourceSchoolID, destSchoolID, initiatorID uuid.UUID) (db.TransferRequest, error) {
	return s.Queries.CreateTransferRequest(ctx, db.CreateTransferRequestParams{
		EntityType:          "Student",
		EntityID:            studentID,
		SourceSchoolID:      sourceSchoolID,
		DestinationSchoolID: destSchoolID,
		InitiatedByUserID:   initiatorID,
	})
}

func (s *StudentService) ProcessTransfer(ctx context.Context, transferID uuid.UUID, status, notes string, schoolID uuid.UUID) (db.TransferRequest, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return db.TransferRequest{}, err
	}
	defer tx.Rollback()

	qtx := s.Queries.WithTx(tx)

	tr, err := qtx.GetTransferRequestByID(ctx, transferID)
	if err != nil {
		return db.TransferRequest{}, err
	}

	if status == "approved" {
		_, err = qtx.CreateEnrollment(ctx, db.CreateEnrollmentParams{
			SchoolID:       schoolID,
			StudentID:      tr.EntityID,
			EnrollmentDate: time.Now(),
			Status:         "Enrolled",
		})
		if err != nil {
			return db.TransferRequest{}, err
		}
	}

	updated, err := qtx.UpdateTransferRequestStatus(ctx, db.UpdateTransferRequestStatusParams{
		TransferID:     transferID,
		Status:         status,
		CompletionDate: sql.NullTime{Time: time.Now(), Valid: true},
		Notes:          sql.NullString{String: notes, Valid: true},
	})
	if err != nil {
		return db.TransferRequest{}, err
	}

	if err := tx.Commit(); err != nil {
		return db.TransferRequest{}, err
	}

	return updated, nil
}
