package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/worker"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type AssignmentService struct {
	Queries *db.Queries
	Asynq   *asynq.Client
}

func NewAssignmentService(q *db.Queries, asynqClient *asynq.Client) *AssignmentService {
	return &AssignmentService{Queries: q, Asynq: asynqClient}
}

type CreateAssignmentParams struct {
	SchoolID       uuid.UUID
	ClassID        uuid.UUID
	TeacherID      uuid.UUID
	Title          string
	Description    string
	DueDate        time.Time
	MaxScore       string
	AssignmentType string
	FileURL        string
}

func (s *AssignmentService) CreateAssignment(ctx context.Context, p CreateAssignmentParams) (db.Assignment, error) {
	// Verify teacher
	academicClass, err := s.Queries.GetClassByID(ctx, db.GetClassByIDParams{
		ClassID:  p.ClassID,
		SchoolID: p.SchoolID,
	})
	if err != nil {
		return db.Assignment{}, err
	}
	if academicClass.TeacherID != p.TeacherID {
		return db.Assignment{}, fmt.Errorf("not authorized to post assignments to this class")
	}

	assignment, err := s.Queries.CreateAssignment(ctx, db.CreateAssignmentParams{
		SchoolID:       p.SchoolID,
		ClassID:        p.ClassID,
		TeacherID:      p.TeacherID,
		Title:          p.Title,
		Description:    sql.NullString{String: p.Description, Valid: p.Description != ""},
		DueDate:        sql.NullTime{Time: p.DueDate, Valid: !p.DueDate.IsZero()},
		MaxScore:       p.MaxScore,
		AssignmentType: p.AssignmentType,
		FileUrl:        sql.NullString{String: p.FileURL, Valid: p.FileURL != ""},
	})
	if err != nil {
		return db.Assignment{}, err
	}

	// Notify students
	payload, _ := json.Marshal(worker.AssignmentNotificationPayload{
		SchoolID:     p.SchoolID,
		ClassID:      p.ClassID,
		TeacherID:    p.TeacherID,
		Title:        assignment.Title,
		DueDate:      p.DueDate.Format("2006-01-02"),
		AssignmentID: assignment.AssignmentID,
	})
	task := asynq.NewTask(worker.TypeAssignmentNotification, payload)
	if _, err := s.Asynq.Enqueue(task); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("could not enqueue notification task: %v\n", err)
	}

	return assignment, nil
}
