package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SubjectHandler struct {
	Queries *db.Queries
}

func NewSubjectHandler(q *db.Queries) *SubjectHandler {
	return &SubjectHandler{Queries: q}
}

func (h *SubjectHandler) GetSubjects(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	subjects, err := h.Queries.GetSubjectsBySchool(r.Context(), schoolID)
	if err != nil {
		middleware.SendError(w, "Could not fetch subjects", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"results": len(subjects),
		"data": map[string]interface{}{
			"subjects": subjects,
		},
	})
}

func (h *SubjectHandler) GetSubjectByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		middleware.SendError(w, "Invalid subject ID", http.StatusBadRequest)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	subject, err := h.Queries.GetSubjectByID(r.Context(), db.GetSubjectByIDParams{
		SubjectID: id,
		SchoolID:  schoolID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.SendError(w, "Subject not found", http.StatusNotFound)
			return
		}
		middleware.SendError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"subject": subject,
		},
	})
}

func (h *SubjectHandler) CreateSubject(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		SubjectName          string `json:"subject_name"`
		Description          string `json:"description"`
		DoublePeriodRequired bool   `json:"double_period_required"`
		LabPeriodRequired    bool   `json:"lab_period_required"`
		MaxOnlinePercentage  string `json:"max_online_percentage"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	subject, err := h.Queries.CreateSubject(r.Context(), db.CreateSubjectParams{
		SchoolID:             schoolID,
		SubjectName:          req.SubjectName,
		Description:          sql.NullString{String: req.Description, Valid: req.Description != ""},
		DoublePeriodRequired: req.DoublePeriodRequired,
		LabPeriodRequired:    req.LabPeriodRequired,
		MaxOnlinePercentage:  sql.NullString{String: req.MaxOnlinePercentage, Valid: req.MaxOnlinePercentage != ""},
	})

	if err != nil {
		middleware.SendError(w, "Could not create subject", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"subject": subject,
		},
	})
}

func (h *SubjectHandler) UpdateSubject(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		SubjectName          string `json:"subject_name"`
		Description          string `json:"description"`
		DoublePeriodRequired bool   `json:"double_period_required"`
		LabPeriodRequired    bool   `json:"lab_period_required"`
		MaxOnlinePercentage  string `json:"max_online_percentage"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	subject, err := h.Queries.UpdateSubject(r.Context(), db.UpdateSubjectParams{
		SubjectID:            id,
		SchoolID:             schoolID,
		SubjectName:          req.SubjectName,
		Description:          sql.NullString{String: req.Description, Valid: req.Description != ""},
		DoublePeriodRequired: req.DoublePeriodRequired,
		LabPeriodRequired:    req.LabPeriodRequired,
		MaxOnlinePercentage:  sql.NullString{String: req.MaxOnlinePercentage, Valid: req.MaxOnlinePercentage != ""},
	})

	if err != nil {
		middleware.SendError(w, "Could not update subject", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{
			"subject": subject,
		},
	})
}

func (h *SubjectHandler) DeleteSubject(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	err := h.Queries.DeleteSubject(r.Context(), db.DeleteSubjectParams{
		SubjectID: id,
		SchoolID:  schoolID,
	})

	if err != nil {
		middleware.SendError(w, "Could not delete subject", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
