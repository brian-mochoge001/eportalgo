package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
)

type EventService struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewEventService(q *db.Queries, d *sql.DB) *EventService {
	return &EventService{Queries: q, DB: d}
}

type CreateEventParams struct {
	SchoolID    uuid.UUID
	Title       string
	Description string
	EventDate   time.Time
	EndDate     *time.Time
	Location    string
	EventType   string
	OrganizerID uuid.UUID
	IsPublic    bool
}

func (s *EventService) CreateEvent(ctx context.Context, p CreateEventParams) (db.Event, error) {
	var endDate sql.NullTime
	if p.EndDate != nil {
		endDate = sql.NullTime{Time: *p.EndDate, Valid: true}
	}

	return s.Queries.CreateEvent(ctx, db.CreateEventParams{
		SchoolID:    p.SchoolID,
		Title:       p.Title,
		Description: sql.NullString{String: p.Description, Valid: p.Description != ""},
		EventDate:   p.EventDate,
		EndDate:     endDate,
		Location:    sql.NullString{String: p.Location, Valid: p.Location != ""},
		EventType:   p.EventType,
		OrganizerID: uuid.NullUUID{UUID: p.OrganizerID, Valid: true},
		IsPublic:    p.IsPublic,
	})
}

type UpdateEventParams struct {
	EventID     uuid.UUID
	SchoolID    uuid.UUID
	Title       string
	Description string
	EventDate   *time.Time
	EndDate     *time.Time
	Location    string
	EventType   string
	OrganizerID *uuid.UUID
	IsPublic    *bool
}

func (s *EventService) UpdateEvent(ctx context.Context, p UpdateEventParams) (db.Event, error) {
	existing, err := s.Queries.GetEventByID(ctx, db.GetEventByIDParams{
		EventID:  p.EventID,
		SchoolID: p.SchoolID,
	})
	if err != nil {
		return db.Event{}, fmt.Errorf("event not found: %w", err)
	}

	params := db.UpdateEventParams{
		EventID:     p.EventID,
		SchoolID:    p.SchoolID,
		Title:       existing.Title,
		Description: existing.Description,
		EventDate:   existing.EventDate,
		EndDate:     existing.EndDate,
		Location:    existing.Location,
		EventType:   existing.EventType,
		OrganizerID: existing.OrganizerID,
		IsPublic:    existing.IsPublic,
	}

	if p.Title != "" {
		params.Title = p.Title
	}
	if p.Description != "" {
		params.Description = sql.NullString{String: p.Description, Valid: true}
	}
	if p.EventDate != nil {
		params.EventDate = *p.EventDate
	}
	if p.EndDate != nil {
		params.EndDate = sql.NullTime{Time: *p.EndDate, Valid: true}
	}
	if p.Location != "" {
		params.Location = sql.NullString{String: p.Location, Valid: true}
	}
	if p.EventType != "" {
		params.EventType = p.EventType
	}
	if p.OrganizerID != nil {
		params.OrganizerID = uuid.NullUUID{UUID: *p.OrganizerID, Valid: true}
	}
	if p.IsPublic != nil {
		params.IsPublic = *p.IsPublic
	}

	return s.Queries.UpdateEvent(ctx, params)
}
