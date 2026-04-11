package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TeacherAvailabilityHandler struct {
	Queries *db.Queries
}

func NewTeacherAvailabilityHandler(q *db.Queries) *TeacherAvailabilityHandler {
	return &TeacherAvailabilityHandler{Queries: q}
}

func (h *TeacherAvailabilityHandler) CreateTeacherAvailability(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	teacherID := userCtx.UserID

	var req struct {
		DayOfWeek int32  `json:"day_of_week"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		IsRecurring bool `json:"is_recurring"`
		Notes     string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	if req.DayOfWeek < 0 || req.DayOfWeek > 6 {
		middleware.ValidationError(w, "Day of week must be between 0 (Sunday) and 6 (Saturday)", nil)
		return
	}

	startTime, err := time.Parse("15:04:05", req.StartTime)
	if err != nil {
		middleware.ValidationError(w, "Invalid start time format. Use HH:MM:SS", err)
		return
	}
	endTime, err := time.Parse("15:04:05", req.EndTime)
	if err != nil {
		middleware.ValidationError(w, "Invalid end time format. Use HH:MM:SS", err)
		return
	}

	availability, err := h.Queries.CreateTeacherAvailability(r.Context(), db.CreateTeacherAvailabilityParams{
		TeacherID:   teacherID,
		DayOfWeek:   req.DayOfWeek,
		StartTime:   startTime.UTC(),
		EndTime:     endTime.UTC(),
		IsRecurring: req.IsRecurring,
		Notes:       toNullString(req.Notes),
	})

	if err != nil {
		middleware.InternalError(w, "Could not create teacher availability", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(availability)
}

func (h *TeacherAvailabilityHandler) GetTeacherAvailabilities(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	teacherIDStr := r.URL.Query().Get("teacherId")

	var teacherID uuid.NullUUID
	if userCtx.RoleName == "Teacher" {
		teacherID = uuid.NullUUID{UUID: userCtx.UserID, Valid: true}
	} else if teacherIDStr != "" {
		if id, err := uuid.Parse(teacherIDStr); err == nil {
			teacherID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	availabilities, err := h.Queries.GetTeacherAvailabilities(r.Context(), teacherID)
	if err != nil {
		middleware.InternalError(w, "Could not fetch teacher availabilities", err)
		return
	}

	json.NewEncoder(w).Encode(availabilities)
}

func (h *TeacherAvailabilityHandler) GetTeacherAvailabilityByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	availabilityID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())

	availability, err := h.Queries.GetTeacherAvailabilityByID(r.Context(), availabilityID)
	if err != nil {
		middleware.NotFoundError(w, "Teacher availability not found", err)
		return
	}

	if userCtx.RoleName == "Teacher" && availability.TeacherID != userCtx.UserID {
		middleware.ForbiddenError(w, "Not authorized to view this availability", nil)
		return
	}

	json.NewEncoder(w).Encode(availability)
}

func (h *TeacherAvailabilityHandler) UpdateTeacherAvailability(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	availabilityID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())

	var req struct {
		DayOfWeek   int32  `json:"day_of_week"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`
		IsRecurring bool   `json:"is_recurring"`
		Notes       string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	existingAvailability, err := h.Queries.GetTeacherAvailabilityByID(r.Context(), availabilityID)
	if err != nil {
		middleware.NotFoundError(w, "Teacher availability not found", err)
		return
	}

	if userCtx.RoleName == "Teacher" && existingAvailability.TeacherID != userCtx.UserID {
		middleware.ForbiddenError(w, "Not authorized to update this availability", nil)
		return
	}

	startTime, err := time.Parse("15:04:05", req.StartTime)
	if err != nil {
		middleware.ValidationError(w, "Invalid start time format. Use HH:MM:SS", err)
		return
	}
	endTime, err := time.Parse("15:04:05", req.EndTime)
	if err != nil {
		middleware.ValidationError(w, "Invalid end time format. Use HH:MM:SS", err)
		return
	}

	updated, err := h.Queries.UpdateTeacherAvailability(r.Context(), db.UpdateTeacherAvailabilityParams{
		AvailabilityID: availabilityID,
		TeacherID:      existingAvailability.TeacherID,
		DayOfWeek:      req.DayOfWeek,
		StartTime:      startTime.UTC(),
		EndTime:        endTime.UTC(),
		IsRecurring:    req.IsRecurring,
		Notes:          toNullString(req.Notes),
	})
	if err != nil {
		middleware.InternalError(w, "Could not update teacher availability", err)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func (h *TeacherAvailabilityHandler) DeleteTeacherAvailability(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	availabilityID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())

	existingAvailability, err := h.Queries.GetTeacherAvailabilityByID(r.Context(), availabilityID)
	if err != nil {
		middleware.NotFoundError(w, "Teacher availability not found", err)
		return
	}

	if userCtx.RoleName == "Teacher" && existingAvailability.TeacherID != userCtx.UserID {
		middleware.ForbiddenError(w, "Not authorized to delete this availability", nil)
		return
	}

	err = h.Queries.DeleteTeacherAvailability(r.Context(), db.DeleteTeacherAvailabilityParams{
		AvailabilityID: availabilityID,
		TeacherID:      existingAvailability.TeacherID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not delete teacher availability", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


