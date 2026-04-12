package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
)

type ClassService struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewClassService(q *db.Queries, d *sql.DB) *ClassService {
	return &ClassService{Queries: q, DB: d}
}

type BulkEnrollResult struct {
	NewlyEnrolledCount   int
	AlreadyEnrolledCount int
}

func (s *ClassService) BulkEnrollStudents(ctx context.Context, classID, schoolID uuid.UUID, studentIDs []uuid.UUID) (BulkEnrollResult, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return BulkEnrollResult{}, err
	}
	defer tx.Rollback()

	qtx := s.Queries.WithTx(tx)

	// Verify class
	_, err = qtx.GetClassByID(ctx, db.GetClassByIDParams{
		ClassID:  classID,
		SchoolID: schoolID,
	})
	if err != nil {
		return BulkEnrollResult{}, fmt.Errorf("class not found: %w", err)
	}

	result := BulkEnrollResult{}

	for _, sid := range studentIDs {
		// Check if already enrolled
		_, err = qtx.GetEnrollmentByStudentAndClass(ctx, db.GetEnrollmentByStudentAndClassParams{
			StudentID: sid,
			ClassID:   classID,
		})
		if err == nil {
			result.AlreadyEnrolledCount++
			continue
		}

		// Verify student
		_, err = qtx.GetUser(ctx, db.GetUserParams{
			UserID:   sid,
			SchoolID: uuid.NullUUID{UUID: schoolID, Valid: true},
		})
		if err != nil {
			continue // Skip non-existent students or different school
		}

		// Create enrollment
		_, err = qtx.CreateEnrollment(ctx, db.CreateEnrollmentParams{
			SchoolID:       schoolID,
			StudentID:      sid,
			ClassID:        classID,
			EnrollmentDate: time.Now(),
			Status:         "Enrolled",
		})
		if err == nil {
			result.NewlyEnrolledCount++
		}
	}

	if err := tx.Commit(); err != nil {
		return BulkEnrollResult{}, err
	}

	return result, nil
}
