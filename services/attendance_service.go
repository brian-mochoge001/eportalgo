package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
)

type AttendanceService struct {
	Queries *db.Queries
}

func NewAttendanceService(q *db.Queries) *AttendanceService {
	return &AttendanceService{Queries: q}
}

type StudentAttendance struct {
	StudentID uuid.UUID
	Status    string
	Notes     string
}

type MarkAttendanceParams struct {
	SchoolID           uuid.UUID
	ClassID            uuid.UUID
	TeacherID          uuid.UUID
	AttendanceDate     time.Time
	StudentsAttendance []StudentAttendance
}

func (s *AttendanceService) MarkAttendance(ctx context.Context, p MarkAttendanceParams) ([]db.AttendanceRecord, error) {
	// Verify teacher
	academicClass, err := s.Queries.GetClassByID(ctx, db.GetClassByIDParams{
		ClassID:  p.ClassID,
		SchoolID: p.SchoolID,
	})
	if err != nil {
		return nil, err
	}
	if academicClass.TeacherID != p.TeacherID {
		return nil, fmt.Errorf("not authorized to mark attendance for this class")
	}

	var results []db.AttendanceRecord
	for _, sa := range p.StudentsAttendance {
		// Check existing
		existing, err := s.Queries.GetAttendanceRecordByUnique(ctx, db.GetAttendanceRecordByUniqueParams{
			SchoolID:       p.SchoolID,
			StudentID:      sa.StudentID,
			ClassID:        p.ClassID,
			AttendanceDate: p.AttendanceDate,
		})

		var record db.AttendanceRecord
		if err == nil {
			record, err = s.Queries.UpdateAttendanceRecord(ctx, db.UpdateAttendanceRecordParams{
				AttendanceID: existing.AttendanceID,
				Status:       sa.Status,
				Notes:        sql.NullString{String: sa.Notes, Valid: sa.Notes != ""},
				SchoolID:     p.SchoolID,
			})
		} else {
			record, err = s.Queries.CreateAttendanceRecord(ctx, db.CreateAttendanceRecordParams{
				SchoolID:       p.SchoolID,
				StudentID:      sa.StudentID,
				ClassID:        p.ClassID,
				AttendanceDate: p.AttendanceDate,
				Status:         sa.Status,
				Notes:          sql.NullString{String: sa.Notes, Valid: sa.Notes != ""},
			})
		}
		if err != nil {
			return nil, fmt.Errorf("failed to mark attendance for student %s: %w", sa.StudentID, err)
		}
		results = append(results, record)
	}

	return results, nil
}
