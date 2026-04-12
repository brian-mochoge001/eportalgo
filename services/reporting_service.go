package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
)

type ReportingService struct {
	Queries *db.Queries
}

func NewReportingService(q *db.Queries) *ReportingService {
	return &ReportingService{Queries: q}
}

type CreateTranscriptParams struct {
	SchoolID       uuid.UUID
	StudentID      uuid.UUID
	AcademicYear   string
	CumulativeGPA  string
	TranscriptData string
	IssuedByUserID uuid.UUID
}

func (s *ReportingService) CreateTranscript(ctx context.Context, p CreateTranscriptParams) (db.Transcript, error) {
	return s.Queries.CreateTranscript(ctx, db.CreateTranscriptParams{
		SchoolID:       p.SchoolID,
		StudentID:      p.StudentID,
		AcademicYear:   p.AcademicYear,
		CumulativeGpa:  sql.NullString{String: p.CumulativeGPA, Valid: p.CumulativeGPA != ""},
		TranscriptData: json.RawMessage(p.TranscriptData),
		IssuedByUserID: uuid.NullUUID{UUID: p.IssuedByUserID, Valid: true},
	})
}

type UpdateTranscriptParams struct {
	TranscriptID   uuid.UUID
	SchoolID       uuid.UUID
	AcademicYear   string
	CumulativeGPA  string
	TranscriptData string
}

func (s *ReportingService) UpdateTranscript(ctx context.Context, p UpdateTranscriptParams) (db.Transcript, error) {
	// Fetch existing to preserve fields
	existing, err := s.Queries.GetTranscriptByID(ctx, db.GetTranscriptByIDParams{
		TranscriptID: p.TranscriptID,
		SchoolID:     p.SchoolID,
	})
	if err != nil {
		return db.Transcript{}, fmt.Errorf("transcript not found: %w", err)
	}

	params := db.UpdateTranscriptParams{
		TranscriptID:   p.TranscriptID,
		SchoolID:       p.SchoolID,
		AcademicYear:   existing.AcademicYear,
		CumulativeGpa:  existing.CumulativeGpa,
		TranscriptData: existing.TranscriptData,
	}

	if p.AcademicYear != "" {
		params.AcademicYear = p.AcademicYear
	}
	if p.CumulativeGPA != "" {
		params.CumulativeGpa = sql.NullString{String: p.CumulativeGPA, Valid: true}
	}
	if p.TranscriptData != "" {
		params.TranscriptData = json.RawMessage(p.TranscriptData)
	}

	return s.Queries.UpdateTranscript(ctx, params)
}
