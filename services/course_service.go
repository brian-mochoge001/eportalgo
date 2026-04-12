package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
)

type CourseService struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewCourseService(q *db.Queries, d *sql.DB) *CourseService {
	return &CourseService{Queries: q, DB: d}
}

type CreateCourseParams struct {
	SchoolID              uuid.UUID
	CourseCode            string
	CourseName            string
	Description           string
	IsShortCourse         bool
	Price                 string
	IsGradedIndependently bool
}

func (s *CourseService) CreateCourse(ctx context.Context, p CreateCourseParams) (db.Course, error) {
	return s.Queries.CreateCourse(ctx, db.CreateCourseParams{
		SchoolID:              p.SchoolID,
		CourseCode:            p.CourseCode,
		CourseName:            p.CourseName,
		Description:           sql.NullString{String: p.Description, Valid: p.Description != ""},
		IsShortCourse:         p.IsShortCourse,
		Price:                 sql.NullString{String: p.Price, Valid: p.Price != ""},
		IsGradedIndependently: p.IsGradedIndependently,
	})
}

func (s *CourseService) EnrollShortCourse(ctx context.Context, courseID, studentID, schoolID uuid.UUID) (db.ShortCourseEnrollment, error) {
	// Verify course
	course, err := s.Queries.GetCourseByID(ctx, db.GetCourseByIDParams{
		CourseID: courseID,
		SchoolID: schoolID,
	})
	if err != nil {
		return db.ShortCourseEnrollment{}, fmt.Errorf("course not found: %w", err)
	}
	if !course.IsShortCourse {
		return db.ShortCourseEnrollment{}, fmt.Errorf("course is not a short course")
	}

	// Check existing enrollment
	_, err = s.Queries.CheckShortCourseEnrollment(ctx, db.CheckShortCourseEnrollmentParams{
		StudentID: studentID,
		CourseID:  courseID,
	})
	if err == nil {
		return db.ShortCourseEnrollment{}, fmt.Errorf("student is already enrolled")
	}

	return s.Queries.EnrollShortCourse(ctx, db.EnrollShortCourseParams{
		StudentID: studentID,
		CourseID:  courseID,
		SchoolID:  schoolID,
		Status:    "Enrolled",
	})
}

func (s *CourseService) UnenrollShortCourse(ctx context.Context, courseID, studentID, schoolID uuid.UUID) error {
	return s.Queries.UnenrollShortCourse(ctx, db.UnenrollShortCourseParams{
		StudentID: studentID,
		CourseID:  courseID,
		SchoolID:  schoolID,
	})
}
