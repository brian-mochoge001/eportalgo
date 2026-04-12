package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
)

type QuizService struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewQuizService(q *db.Queries, d *sql.DB) *QuizService {
	return &QuizService{Queries: q, DB: d}
}

type QuizAnswerRequest struct {
	QuestionID        uuid.UUID
	StudentAnswerText string
	SelectedOptionID  uuid.NullUUID
}

type QuizSubmissionRequest struct {
	QuizID    uuid.UUID
	StudentID uuid.UUID
	Answers   []QuizAnswerRequest
}

func (s *QuizService) SubmitQuiz(ctx context.Context, req QuizSubmissionRequest) (db.QuizSubmission, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return db.QuizSubmission{}, err
	}
	defer tx.Rollback()

	qtx := s.Queries.WithTx(tx)

	// Create Submission
	submission, err := qtx.CreateQuizSubmission(ctx, db.CreateQuizSubmissionParams{
		QuizID:    req.QuizID,
		StudentID: req.StudentID,
		Status:    "completed",
	})
	if err != nil {
		return db.QuizSubmission{}, fmt.Errorf("could not create submission: %w", err)
	}

	// Create Answers
	for _, a := range req.Answers {
		_, err = qtx.CreateQuizAnswer(ctx, db.CreateQuizAnswerParams{
			QuizSubmissionID:  submission.SubmissionID,
			QuestionID:        a.QuestionID,
			StudentAnswerText: sql.NullString{String: a.StudentAnswerText, Valid: a.StudentAnswerText != ""},
			SelectedOptionID:  a.SelectedOptionID,
		})
		if err != nil {
			return db.QuizSubmission{}, fmt.Errorf("could not create answer for question %s: %w", a.QuestionID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return db.QuizSubmission{}, err
	}

	return submission, nil
}
