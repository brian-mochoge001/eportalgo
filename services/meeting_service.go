package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
)

type MeetingService struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewMeetingService(q *db.Queries, d *sql.DB) *MeetingService {
	return &MeetingService{Queries: q, DB: d}
}

type CreateMeetingParams struct {
	SchoolID        uuid.UUID
	Title           string
	Agenda          string
	MeetingDate     time.Time
	DurationMinutes int32
	Location        string
	MeetingType     string
	OrganizerID     uuid.UUID
	AttendeeIDs     []uuid.UUID
}

func (s *MeetingService) CreateMeeting(ctx context.Context, p CreateMeetingParams) (db.Meeting, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return db.Meeting{}, err
	}
	defer tx.Rollback()
	qtx := s.Queries.WithTx(tx)

	meeting, err := qtx.CreateMeeting(ctx, db.CreateMeetingParams{
		SchoolID:        p.SchoolID,
		Title:           p.Title,
		Agenda:          sql.NullString{String: p.Agenda, Valid: p.Agenda != ""},
		MeetingDate:     p.MeetingDate,
		DurationMinutes: sql.NullInt32{Int32: p.DurationMinutes, Valid: p.DurationMinutes > 0},
		Location:        sql.NullString{String: p.Location, Valid: p.Location != ""},
		MeetingType:     p.MeetingType,
		OrganizerID:     uuid.NullUUID{UUID: p.OrganizerID, Valid: true},
	})
	if err != nil {
		return db.Meeting{}, fmt.Errorf("could not create meeting: %w", err)
	}

	for _, attendeeID := range p.AttendeeIDs {
		err := qtx.AddMeetingAttendee(ctx, db.AddMeetingAttendeeParams{
			MeetingID: meeting.MeetingID,
			UserID:    attendeeID,
			SchoolID:  p.SchoolID,
		})
		if err != nil {
			return db.Meeting{}, fmt.Errorf("could not add attendee %s: %w", attendeeID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return db.Meeting{}, err
	}

	return meeting, nil
}

type UpdateMeetingParams struct {
	MeetingID       uuid.UUID
	SchoolID        uuid.UUID
	Title           string
	Agenda          string
	MeetingDate     *time.Time
	DurationMinutes *int32
	Location        string
	MeetingType     string
	OrganizerID     *uuid.UUID
}

func (s *MeetingService) UpdateMeeting(ctx context.Context, p UpdateMeetingParams) (db.Meeting, error) {
	existing, err := s.Queries.GetMeetingByID(ctx, db.GetMeetingByIDParams{
		MeetingID: p.MeetingID,
		SchoolID:  p.SchoolID,
	})
	if err != nil {
		return db.Meeting{}, fmt.Errorf("meeting not found: %w", err)
	}

	params := db.UpdateMeetingParams{
		MeetingID:       p.MeetingID,
		SchoolID:        p.SchoolID,
		Title:           existing.Title,
		Agenda:          existing.Agenda,
		MeetingDate:     existing.MeetingDate,
		DurationMinutes: existing.DurationMinutes,
		Location:        existing.Location,
		MeetingType:     existing.MeetingType,
		OrganizerID:     existing.OrganizerID,
	}

	if p.Title != "" {
		params.Title = p.Title
	}
	if p.Agenda != "" {
		params.Agenda = sql.NullString{String: p.Agenda, Valid: true}
	}
	if p.MeetingDate != nil {
		params.MeetingDate = *p.MeetingDate
	}
	if p.DurationMinutes != nil {
		params.DurationMinutes = sql.NullInt32{Int32: *p.DurationMinutes, Valid: true}
	}
	if p.Location != "" {
		params.Location = sql.NullString{String: p.Location, Valid: true}
	}
	if p.MeetingType != "" {
		params.MeetingType = p.MeetingType
	}
	if p.OrganizerID != nil {
		params.OrganizerID = uuid.NullUUID{UUID: *p.OrganizerID, Valid: true}
	}

	return s.Queries.UpdateMeeting(ctx, params)
}
