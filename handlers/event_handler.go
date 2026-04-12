package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/brian-mochoge001/eportalgo/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type EventHandler struct {
	Queries      *db.Queries
	EventService *services.EventService
}

func NewEventHandler(q *db.Queries, s *services.EventService) *EventHandler {
	return &EventHandler{Queries: q, EventService: s}
}

func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		StartTime   string `json:"start_time"` // We map this to EventDate
		EndTime     string `json:"end_time"`   // We map this to EndDate
		Location    string `json:"location"`
		EventType   string `json:"event_type"`
		OrganizerID string `json:"organizer_id"`
		IsPublic    bool   `json:"is_public"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	eventDate, _ := time.Parse(time.RFC3339, req.StartTime)
	var endDate *time.Time
	if req.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			endDate = &t
		}
	}
	organizerID, _ := uuid.Parse(req.OrganizerID)
	if organizerID == uuid.Nil {
		organizerID = userCtx.UserID
	}

	event, err := h.EventService.CreateEvent(r.Context(), services.CreateEventParams{
		SchoolID:    schoolID,
		Title:       req.Title,
		Description: req.Description,
		EventDate:   eventDate,
		EndDate:     endDate,
		Location:    req.Location,
		EventType:   req.EventType,
		OrganizerID: organizerID,
		IsPublic:    req.IsPublic,
	})

	if err != nil {
		middleware.InternalError(w, err.Error(), err)
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
		middleware.InternalError(w, "Could not fetch events", err)
		return
	}

	json.NewEncoder(w).Encode(events)
}

func (h *EventHandler) GetEventByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	eventID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	event, err := h.Queries.GetEventByID(r.Context(), db.GetEventByIDParams{
		EventID:  eventID,
		SchoolID: schoolID,
	})
	if err != nil {
		middleware.NotFoundError(w, "Event not found", err)
		return
	}

	json.NewEncoder(w).Encode(event)
}

func (h *EventHandler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	eventID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`
		Location    string `json:"location"`
		EventType   string `json:"event_type"`
		OrganizerID string `json:"organizer_id"`
		IsPublic    *bool  `json:"is_public"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	params := services.UpdateEventParams{
		EventID:     eventID,
		SchoolID:    schoolID,
		Title:       req.Title,
		Description: req.Description,
		Location:    req.Location,
		EventType:   req.EventType,
		IsPublic:    req.IsPublic,
	}

	if req.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
			params.EventDate = &t
		}
	}
	if req.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			params.EndDate = &t
		}
	}
	if req.OrganizerID != "" {
		if id, err := uuid.Parse(req.OrganizerID); err == nil {
			params.OrganizerID = &id
		}
	}

	updated, err := h.EventService.UpdateEvent(r.Context(), params)
	if err != nil {
		middleware.InternalError(w, err.Error(), err)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func (h *EventHandler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	eventID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	err := h.Queries.DeleteEvent(r.Context(), db.DeleteEventParams{
		EventID:  eventID,
		SchoolID: schoolID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not delete event", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
