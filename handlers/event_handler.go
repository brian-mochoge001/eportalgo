package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type EventHandler struct {
	Queries *db.Queries
}

func NewEventHandler(q *db.Queries) *EventHandler {
	return &EventHandler{Queries: q}
}

func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		EventDate   string    `json:"event_date"`
		EndDate     string    `json:"end_date"`
		Location    string    `json:"location"`
		EventType   string    `json:"event_type"`
		OrganizerID string    `json:"organizer_id"`
		IsPublic    *bool     `json:"is_public"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.EventDate == "" || req.EventType == "" {
		middleware.SendError(w, "Title, event date, and event type are required", http.StatusBadRequest)
		return
	}

	eventDate, err := time.Parse(time.RFC3339, req.EventDate)
	if err != nil {
		middleware.SendError(w, "Invalid event date format", http.StatusBadRequest)
		return
	}

	var endDate sql.NullTime
	if req.EndDate != "" {
		t, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			middleware.SendError(w, "Invalid end date format", http.StatusBadRequest)
			return
		}
		endDate = sql.NullTime{Time: t, Valid: true}
	}

	organizerID := uuid.NullUUID{UUID: userCtx.UserID, Valid: true}
	if req.OrganizerID != "" {
		parsedID, err := uuid.Parse(req.OrganizerID)
		if err == nil {
			organizerID = uuid.NullUUID{UUID: parsedID, Valid: true}
		}
	}

	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	event, err := h.Queries.CreateEvent(r.Context(), db.CreateEventParams{
		SchoolID:    schoolID,
		Title:       req.Title,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		EventDate:   eventDate,
		EndDate:     endDate,
		Location:    sql.NullString{String: req.Location, Valid: req.Location != ""},
		EventType:   req.EventType,
		OrganizerID: organizerID,
		IsPublic:    isPublic,
	})

	if err != nil {
		middleware.SendError(w, "Could not create event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(event)
}

func (h *EventHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	events, err := h.Queries.GetEventsBySchool(r.Context(), schoolID)
	if err != nil {
		middleware.SendError(w, "Could not fetch events", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(events)
}

func (h *EventHandler) GetEventByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.SendError(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	event, err := h.Queries.GetEventByID(r.Context(), db.GetEventByIDParams{
		EventID:  eventID,
		SchoolID: schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Event not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(event)
}

func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.SendError(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		EventDate   string `json:"event_date"`
		EndDate     string `json:"end_date"`
		Location    string `json:"location"`
		EventType   string `json:"event_type"`
		OrganizerID string `json:"organizer_id"`
		IsPublic    *bool  `json:"is_public"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Fetch existing event to get current values for partial updates if needed, 
	// but UpdateEvent query in SQL seems to expect all fields.
	existingEvent, err := h.Queries.GetEventByID(r.Context(), db.GetEventByIDParams{
		EventID:  eventID,
		SchoolID: schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Event not found", http.StatusNotFound)
		return
	}

	params := db.UpdateEventParams{
		EventID:     eventID,
		SchoolID:    schoolID,
		Title:       existingEvent.Title,
		Description: existingEvent.Description,
		EventDate:   existingEvent.EventDate,
		EndDate:     existingEvent.EndDate,
		Location:    existingEvent.Location,
		EventType:   existingEvent.EventType,
		OrganizerID: existingEvent.OrganizerID,
		IsPublic:    existingEvent.IsPublic,
	}

	if req.Title != "" {
		params.Title = req.Title
	}
	if req.Description != "" {
		params.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.EventDate != "" {
		t, err := time.Parse(time.RFC3339, req.EventDate)
		if err == nil {
			params.EventDate = t
		}
	}
	if req.EndDate != "" {
		t, err := time.Parse(time.RFC3339, req.EndDate)
		if err == nil {
			params.EndDate = sql.NullTime{Time: t, Valid: true}
		}
	}
	if req.Location != "" {
		params.Location = sql.NullString{String: req.Location, Valid: true}
	}
	if req.EventType != "" {
		params.EventType = req.EventType
	}
	if req.OrganizerID != "" {
		parsedID, err := uuid.Parse(req.OrganizerID)
		if err == nil {
			params.OrganizerID = uuid.NullUUID{UUID: parsedID, Valid: true}
		}
	}
	if req.IsPublic != nil {
		params.IsPublic = *req.IsPublic
	}

	updatedEvent, err := h.Queries.UpdateEvent(r.Context(), params)
	if err != nil {
		middleware.SendError(w, "Could not update event", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updatedEvent)
}

func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	eventID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.SendError(w, "Invalid event ID", http.StatusBadRequest)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	err = h.Queries.DeleteEvent(r.Context(), db.DeleteEventParams{
		EventID:  eventID,
		SchoolID: schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Could not delete event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
