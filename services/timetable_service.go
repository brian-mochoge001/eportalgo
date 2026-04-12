package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/services/scheduler"
	"github.com/google/uuid"
)

type TimetableService struct {
	Queries   *db.Queries
	DB        *sql.DB
	Scheduler *scheduler.Scheduler
}

func NewTimetableService(q *db.Queries, d *sql.DB) *TimetableService {
	s := scheduler.NewScheduler(q, scheduler.Config{})
	return &TimetableService{Queries: q, DB: d, Scheduler: s}
}

func (s *TimetableService) GenerateAndSaveTimetable(ctx context.Context, timetableID, schoolID uuid.UUID) (float64, error) {
	// 1. Get timetable details
	timetable, err := s.Queries.GetTimetableByID(ctx, db.GetTimetableByIDParams{
		TimetableID: timetableID,
		SchoolID:    schoolID,
	})
	if err != nil {
		return 0, fmt.Errorf("could not find timetable: %w", err)
	}

	// 2. Generate
	result, err := s.Scheduler.Generate(ctx, schoolID, timetable.AcademicYear, timetable.Semester.String)
	if err != nil {
		return 0, fmt.Errorf("scheduling failed: %w", err)
	}

	// 3. Save results within a transaction
	tx, err := s.DB.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	qtx := s.Queries.WithTx(tx)

	// Delete old entries
	err = qtx.DeleteTimetableEntriesByTimetable(ctx, timetableID)
	if err != nil {
		return 0, fmt.Errorf("could not clear old entries: %w", err)
	}

	for _, gene := range result.Genes {
		_, err = qtx.CreateTimetableEntry(ctx, db.CreateTimetableEntryParams{
			TimetableID: timetableID,
			ClassID:     gene.ClassID,
			SubjectID:   gene.SubjectID,
			TeacherID:   gene.TeacherID,
			RoomID:      gene.RoomID,
			DayOfWeek:   int32(gene.DayOfWeek),
			StartTime:   gene.StartTime,
			EndTime:     gene.EndTime,
		})
		if err != nil {
			return 0, fmt.Errorf("could not save timetable entry: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return result.Fitness, nil
}
