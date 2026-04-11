package handlers

import (
	"encoding/json"
	"net/http"
	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TeacherWorkloadHandler struct {
	Queries *db.Queries
}

func NewTeacherWorkloadHandler(q *db.Queries) *TeacherWorkloadHandler {
	return &TeacherWorkloadHandler{Queries: q}
}

func (h *TeacherWorkloadHandler) CreateTeacherWorkload(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TeacherID          string `json:"teacher_id"`
		MaxHoursPerWeek    string `json:"max_hours_per_week"`
		CurrentHoursPerWeek string `json:"current_hours_per_week"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	teacherID, _ := uuid.Parse(req.TeacherID)

	workload, err := h.Queries.CreateTeacherWorkload(r.Context(), db.CreateTeacherWorkloadParams{
		TeacherID:           teacherID,
		MaxHoursPerWeek:     req.MaxHoursPerWeek,
		CurrentHoursPerWeek: req.CurrentHoursPerWeek,
	})

	if err != nil {
		middleware.SendError(w, "Could not create teacher workload", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(workload)
}

func (h *TeacherWorkloadHandler) GetTeacherWorkloads(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID
	teacherIDStr := r.URL.Query().Get("teacherId")

	var teacherID uuid.NullUUID
	if userCtx.RoleName == "Teacher" {
		teacherID = uuid.NullUUID{UUID: userCtx.UserID, Valid: true}
	} else if teacherIDStr != "" {
		if id, err := uuid.Parse(teacherIDStr); err == nil {
			teacherID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	workloads, err := h.Queries.GetTeacherWorkloads(r.Context(), db.GetTeacherWorkloadsParams{
		SchoolID:  uuid.NullUUID{UUID: schoolID, Valid: true},
		TeacherID: teacherID,
	})
	if err != nil {
		middleware.SendError(w, "Could not fetch teacher workloads", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(workloads)
}

func (h *TeacherWorkloadHandler) GetTeacherWorkloadByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	workloadID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())

	workload, err := h.Queries.GetTeacherWorkloadByID(r.Context(), workloadID)
	if err != nil {
		middleware.SendError(w, "Teacher workload not found", http.StatusNotFound)
		return
	}

	if userCtx.RoleName == "Teacher" && workload.TeacherID != userCtx.UserID {
		middleware.SendError(w, "Not authorized to view this workload", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(workload)
}

func (h *TeacherWorkloadHandler) UpdateTeacherWorkload(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	workloadID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())

	var req struct {
		MaxHoursPerWeek    string `json:"max_hours_per_week"`
		CurrentHoursPerWeek string `json:"current_hours_per_week"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	existingWorkload, err := h.Queries.GetTeacherWorkloadByID(r.Context(), workloadID)
	if err != nil {
		middleware.SendError(w, "Teacher workload not found", http.StatusNotFound)
		return
	}

	if userCtx.RoleName == "Teacher" && existingWorkload.TeacherID != userCtx.UserID {
		middleware.SendError(w, "Not authorized to update this workload", http.StatusForbidden)
		return
	}

	params := db.UpdateTeacherWorkloadParams{
		WorkloadID:          workloadID,
		MaxHoursPerWeek:     existingWorkload.MaxHoursPerWeek,
		CurrentHoursPerWeek: existingWorkload.CurrentHoursPerWeek,
	}

	if req.MaxHoursPerWeek != "" {
		params.MaxHoursPerWeek = req.MaxHoursPerWeek
	}
	if req.CurrentHoursPerWeek != "" {
		params.CurrentHoursPerWeek = req.CurrentHoursPerWeek
	}

	updated, err := h.Queries.UpdateTeacherWorkload(r.Context(), params)
	if err != nil {
		middleware.SendError(w, "Could not update teacher workload", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func (h *TeacherWorkloadHandler) DeleteTeacherWorkload(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	workloadID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())

	existingWorkload, err := h.Queries.GetTeacherWorkloadByID(r.Context(), workloadID)
	if err != nil {
		middleware.SendError(w, "Teacher workload not found", http.StatusNotFound)
		return
	}

	if userCtx.RoleName == "Teacher" && existingWorkload.TeacherID != userCtx.UserID {
		middleware.SendError(w, "Not authorized to delete this workload", http.StatusForbidden)
		return
	}

	err = h.Queries.DeleteTeacherWorkload(r.Context(), workloadID)
	if err != nil {
		middleware.SendError(w, "Could not delete teacher workload", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
