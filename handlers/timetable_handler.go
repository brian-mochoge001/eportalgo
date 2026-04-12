package handlers

import (
	"encoding/json"
	"net/http"
	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/brian-mochoge001/eportalgo/services/scheduler"
	"github.com/google/uuid"
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

	// 1. Get timetable details
	timetable, err := h.Queries.GetTimetableByID(r.Context(), db.GetTimetableByIDParams{
		TimetableID: req.TimetableID,
		SchoolID:    schoolID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not find timetable", err)
		return
	}

	// 2. Initialize scheduler
	s := scheduler.NewScheduler(h.Queries, scheduler.Config{})

	// 3. Generate
	result, err := s.Generate(r.Context(), schoolID, timetable.AcademicYear, timetable.Semester.String)
	if err != nil {
		middleware.InternalError(w, "Scheduling failed", err)
		return
	}

	// 4. Save results (delete old entries first)
	err = h.Queries.DeleteTimetableEntriesByTimetable(r.Context(), req.TimetableID)
	if err != nil {
		middleware.InternalError(w, "Could not clear old entries", err)
		return
	}

	for _, gene := range result.Genes {
		_, err = h.Queries.CreateTimetableEntry(r.Context(), db.CreateTimetableEntryParams{
			TimetableID: req.TimetableID,
			ClassID:     gene.ClassID,
			SubjectID:   gene.SubjectID,
			TeacherID:   gene.TeacherID,
			RoomID:      gene.RoomID,
			DayOfWeek:   int32(gene.DayOfWeek),
			StartTime:   gene.StartTime,
			EndTime:     gene.EndTime,
		})
		if err != nil {
			middleware.InternalError(w, "Could not save timetable entry", err)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"fitness": result.Fitness,
		"message": "Timetable generated successfully",
	})
}


