package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type DepartmentHandler struct {
	Queries *db.Queries
}

func NewDepartmentHandler(q *db.Queries) *DepartmentHandler {
	return &DepartmentHandler{Queries: q}
}

func (h *DepartmentHandler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		DepartmentName           string   `json:"departmentName"`
		HeadOfDepartmentId       string   `json:"headOfDepartmentId"`
		DeputyHeadOfDepartmentId string   `json:"deputyHeadOfDepartmentId"`
		SubjectIds               []string `json:"subjectIds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	// Check existing
	existing, _ := h.Queries.GetDepartmentByName(r.Context(), db.GetDepartmentByNameParams{
		SchoolID:       schoolID,
		DepartmentName: req.DepartmentName,
	})
	if existing.DepartmentID != uuid.Nil {
		middleware.SendError(w, "Department already exists", http.StatusConflict, "CONFLICT", nil)
		return
	}

	headID, _ := uuid.Parse(req.HeadOfDepartmentId)
	deputyID, _ := uuid.Parse(req.DeputyHeadOfDepartmentId)

	department, err := h.Queries.CreateDepartment(r.Context(), db.CreateDepartmentParams{
		SchoolID:                 schoolID,
		DepartmentName:           req.DepartmentName,
		HeadOfDepartmentID:       uuid.NullUUID{UUID: headID, Valid: req.HeadOfDepartmentId != ""},
		DeputyHeadOfDepartmentID: uuid.NullUUID{UUID: deputyID, Valid: req.DeputyHeadOfDepartmentId != ""},
	})

	if err != nil {
		middleware.InternalError(w, "Could not create department", err)
		return
	}

	// Add subjects
	for _, sidStr := range req.SubjectIds {
		sid, err := uuid.Parse(sidStr)
		if err == nil {
			h.Queries.AddDepartmentSubject(r.Context(), db.AddDepartmentSubjectParams{
				DepartmentID: department.DepartmentID,
				SubjectID:    sid,
			})
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(department)
}

func (h *DepartmentHandler) GetDepartments(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	departments, err := h.Queries.GetDepartmentsBySchool(r.Context(), schoolID)
	if err != nil {
		middleware.InternalError(w, "Could not fetch departments", err)
		return
	}

	json.NewEncoder(w).Encode(departments)
}

func (h *DepartmentHandler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	deptIDStr := chi.URLParam(r, "departmentId")
	deptID, err := uuid.Parse(deptIDStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid department ID", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		DepartmentName           string   `json:"departmentName"`
		HeadOfDepartmentId       string   `json:"headOfDepartmentId"`
		DeputyHeadOfDepartmentId string   `json:"deputyHeadOfDepartmentId"`
		SubjectIds               []string `json:"subjectIds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	_, err = h.Queries.GetDepartmentByID(r.Context(), db.GetDepartmentByIDParams{
		DepartmentID: deptID,
		SchoolID:     schoolID,
	})
	if err != nil {
		middleware.NotFoundError(w, "Department not found", err)
		return
	}

	headID, _ := uuid.Parse(req.HeadOfDepartmentId)
	deputyID, _ := uuid.Parse(req.DeputyHeadOfDepartmentId)

	updated, err := h.Queries.UpdateDepartment(r.Context(), db.UpdateDepartmentParams{
		DepartmentID:             deptID,
		SchoolID:                 schoolID,
		DepartmentName:           req.DepartmentName,
		HeadOfDepartmentID:       uuid.NullUUID{UUID: headID, Valid: req.HeadOfDepartmentId != ""},
		DeputyHeadOfDepartmentID: uuid.NullUUID{UUID: deputyID, Valid: req.DeputyHeadOfDepartmentId != ""},
	})

	if err != nil {
		middleware.InternalError(w, "Could not update department", err)
		return
	}

	// Update subjects
	if req.SubjectIds != nil {
		h.Queries.ClearDepartmentSubjects(r.Context(), deptID)
		for _, sidStr := range req.SubjectIds {
			sid, err := uuid.Parse(sidStr)
			if err == nil {
				h.Queries.AddDepartmentSubject(r.Context(), db.AddDepartmentSubjectParams{
					DepartmentID: deptID,
					SubjectID:    sid,
				})
			}
		}
	}

	json.NewEncoder(w).Encode(updated)
}

func (h *DepartmentHandler) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	deptIDStr := chi.URLParam(r, "departmentId")
	deptID, _ := uuid.Parse(deptIDStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	err := h.Queries.DeleteDepartment(r.Context(), db.DeleteDepartmentParams{
		DepartmentID: deptID,
		SchoolID:     schoolID,
	})

	if err != nil {
		middleware.InternalError(w, "Could not delete department", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Department deleted successfully"})
}



