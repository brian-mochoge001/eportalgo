package handlers

import (
	"encoding/json"
	"net/http"
	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/brian-mochoge001/eportalgo/services"
	"github.com/google/uuid"
)

type TimetableHandler struct {
	Queries          *db.Queries
	TimetableService *services.TimetableService
}

func NewTimetableHandler(q *db.Queries, s *services.TimetableService) *TimetableHandler {
	return &TimetableHandler{Queries: q, TimetableService: s}
}

func (h *TimetableHandler) GetTimetables(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	timetables, err := h.Queries.GetTimetables(r.Context(), schoolID)
	if err != nil {
		middleware.InternalError(w, "Could not fetch timetables", err)
		return
	}

	json.NewEncoder(w).Encode(timetables)
}

func (h *TimetableHandler) GetTimetableEntries(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	timetableID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid timetable ID", err)
		return
	}

	entries, err := h.Queries.GetTimetableEntries(r.Context(), timetableID)
	if err != nil {
		middleware.InternalError(w, "Could not fetch timetable entries", err)
		return
	}

	json.NewEncoder(w).Encode(entries)
}

func (h *TimetableHandler) CreateTimetable(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		AcademicYear string `json:"academic_year"`
		Semester     string `json:"semester"`
		Title        string `json:"title"`
		Description  string `json:"description"`
		IsActive     bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	timetable, err := h.Queries.CreateTimetable(r.Context(), db.CreateTimetableParams{
		SchoolID:     schoolID,
		AcademicYear: req.AcademicYear,
		Semester:     toNullString(req.Semester),
		Title:        req.Title,
		Description:  toNullString(req.Description),
		IsActive:     req.IsActive,
	})

	if err != nil {
		middleware.InternalError(w, "Could not create timetable", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(timetable)
}

func (h *TimetableHandler) GenerateTimetable(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		TimetableID uuid.UUID `json:"timetable_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	fitness, err := h.TimetableService.GenerateAndSaveTimetable(r.Context(), req.TimetableID, schoolID)
	if err != nil {
		middleware.InternalError(w, "Scheduling failed", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"fitness": fitness,
		"message": "Timetable generated successfully",
	})
}


