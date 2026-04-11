package handlers

import (
	"encoding/json"
	"net/http"
	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
)

type TimetableHandler struct {
	Queries *db.Queries
}

func NewTimetableHandler(q *db.Queries) *TimetableHandler {
	return &TimetableHandler{Queries: q}
}

func (h *TimetableHandler) GetTimetables(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	timetables, err := h.Queries.GetTimetables(r.Context(), schoolID)
	if err != nil {
		middleware.SendError(w, "Could not fetch timetables", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(timetables)
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
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
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
		middleware.SendError(w, "Could not create timetable", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(timetable)
}
