package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
)

type LessonPlanService struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewLessonPlanService(q *db.Queries, d *sql.DB) *LessonPlanService {
	return &LessonPlanService{Queries: q, DB: d}
}

type CreateLessonPlanParams struct {
	SchoolID    uuid.UUID
	TeacherID   uuid.UUID
	ClassID     uuid.NullUUID
	Title       string
	Content     string
	DateCovered *time.Time
}

func (s *LessonPlanService) CreateLessonPlan(ctx context.Context, p CreateLessonPlanParams) (db.LessonPlan, error) {
	var dateCovered sql.NullTime
	if p.DateCovered != nil {
		dateCovered = sql.NullTime{Time: *p.DateCovered, Valid: true}
	}

	return s.Queries.CreateLessonPlan(ctx, db.CreateLessonPlanParams{
		SchoolID:    p.SchoolID,
		TeacherID:   p.TeacherID,
		ClassID:     p.ClassID,
		Title:       p.Title,
		Content:     sql.NullString{String: p.Content, Valid: p.Content != ""},
		DateCovered: dateCovered,
	})
}

type UpdateLessonPlanParams struct {
	LessonPlanID uuid.UUID
	SchoolID     uuid.UUID
	TeacherID    uuid.UUID
	RoleName     string
	Title        string
	Content      string
	ClassID      *uuid.UUID
	DateCovered  *time.Time
}

func (s *LessonPlanService) UpdateLessonPlan(ctx context.Context, p UpdateLessonPlanParams) (db.LessonPlan, error) {
	existing, err := s.Queries.GetLessonPlanByID(ctx, db.GetLessonPlanByIDParams{
		LessonPlanID: p.LessonPlanID,
		SchoolID:     p.SchoolID,
	})
	if err != nil {
		return db.LessonPlan{}, fmt.Errorf("lesson plan not found: %w", err)
	}

	// Ownership check for teachers
	if p.RoleName == "Teacher" && existing.TeacherID != p.TeacherID {
		return db.LessonPlan{}, fmt.Errorf("not authorized to update this lesson plan")
	}

	params := db.UpdateLessonPlanParams{
		LessonPlanID: p.LessonPlanID,
		SchoolID:     p.SchoolID,
		TeacherID:    existing.TeacherID,
		Title:        existing.Title,
		Content:      existing.Content,
		ClassID:      existing.ClassID,
		DateCovered:  existing.DateCovered,
	}

	if p.Title != "" {
		params.Title = p.Title
	}
	if p.Content != "" {
		params.Content = sql.NullString{String: p.Content, Valid: true}
	}
	if p.ClassID != nil {
		params.ClassID = uuid.NullUUID{UUID: *p.ClassID, Valid: true}
	}
	if p.DateCovered != nil {
		params.DateCovered = sql.NullTime{Time: *p.DateCovered, Valid: true}
	}

	return s.Queries.UpdateLessonPlan(ctx, params)
}

func (s *LessonPlanService) DeleteLessonPlan(ctx context.Context, lessonPlanID, schoolID, teacherID uuid.UUID, roleName string) error {
	existing, err := s.Queries.GetLessonPlanByID(ctx, db.GetLessonPlanByIDParams{
		LessonPlanID: lessonPlanID,
		SchoolID:     schoolID,
	})
	if err != nil {
		return fmt.Errorf("lesson plan not found: %w", err)
	}

	// Ownership check for teachers
	if roleName == "Teacher" && existing.TeacherID != teacherID {
		return fmt.Errorf("not authorized to delete this lesson plan")
	}

	return s.Queries.DeleteLessonPlan(ctx, db.DeleteLessonPlanParams{
		LessonPlanID: lessonPlanID,
		SchoolID:     schoolID,
		TeacherID:    existing.TeacherID,
	})
}
