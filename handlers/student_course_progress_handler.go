package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type StudentCourseProgressHandler struct {
	Queries *db.Queries
}

func NewStudentCourseProgressHandler(q *db.Queries) *StudentCourseProgressHandler {
	return &StudentCourseProgressHandler{Queries: q}
}

func (h *StudentCourseProgressHandler) CreateStudentCourseProgress(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EnrollmentID        string `json:"enrollment_id"`
		ProgressPercentage string `json:"progress_percentage"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	enrollmentID, _ := uuid.Parse(req.EnrollmentID)

	progress, err := h.Queries.CreateStudentCourseProgress(r.Context(), db.CreateStudentCourseProgressParams{
		EnrollmentID:       enrollmentID,
		ProgressPercentage: req.ProgressPercentage,
	})

	if err != nil {
		middleware.InternalError(w, "Could not create student course progress", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(progress)
}

func (h *StudentCourseProgressHandler) GetStudentCourseProgresses(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	enrollmentIDStr := r.URL.Query().Get("enrollmentId")
	studentIDStr := r.URL.Query().Get("studentId")

	var enrollmentID uuid.NullUUID
	if enrollmentIDStr != "" {
		if id, err := uuid.Parse(enrollmentIDStr); err == nil {
			enrollmentID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	var studentID uuid.NullUUID
	if userCtx.RoleName == "Student" {
		studentID = uuid.NullUUID{UUID: userCtx.UserID, Valid: true}
	} else if studentIDStr != "" {
		if id, err := uuid.Parse(studentIDStr); err == nil {
			studentID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	progresses, err := h.Queries.GetStudentCourseProgresses(r.Context(), db.GetStudentCourseProgressesParams{
		SchoolID:     schoolID,
		EnrollmentID: enrollmentID,
		StudentID:    studentID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not fetch student course progresses", err)
		return
	}

	json.NewEncoder(w).Encode(progresses)
}

func (h *StudentCourseProgressHandler) GetStudentCourseProgressByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	progressID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	progress, err := h.Queries.GetStudentCourseProgressByID(r.Context(), progressID)
	if err != nil {
		middleware.NotFoundError(w, "Student course progress not found", err)
		return
	}

	if progress.SchoolID != schoolID {
		middleware.ForbiddenError(w, "Not authorized to view this progress", err)
		return
	}

	if userCtx.RoleName == "Student" && progress.StudentID != userCtx.UserID {
		middleware.ForbiddenError(w, "Not authorized to view this progress", err)
		return
	}

	json.NewEncoder(w).Encode(progress)
}

func (h *StudentCourseProgressHandler) UpdateStudentCourseProgress(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	progressID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		ProgressPercentage string `json:"progress_percentage"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	existingProgress, err := h.Queries.GetStudentCourseProgressByID(r.Context(), progressID)
	if err != nil {
		middleware.NotFoundError(w, "Student course progress not found", err)
		return
	}

	if existingProgress.SchoolID != schoolID {
		middleware.ForbiddenError(w, "Not authorized to update this progress", err)
		return
	}

	updated, err := h.Queries.UpdateStudentCourseProgress(r.Context(), db.UpdateStudentCourseProgressParams{
		ProgressID:         progressID,
		ProgressPercentage: req.ProgressPercentage,
	})
	if err != nil {
		middleware.InternalError(w, "Could not update student course progress", err)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func (h *StudentCourseProgressHandler) DeleteStudentCourseProgress(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	progressID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	existingProgress, err := h.Queries.GetStudentCourseProgressByID(r.Context(), progressID)
	if err != nil {
		middleware.NotFoundError(w, "Student course progress not found", err)
		return
	}

	if existingProgress.SchoolID != schoolID {
		middleware.ForbiddenError(w, "Not authorized to delete this progress", err)
		return
	}

	err = h.Queries.DeleteStudentCourseProgress(r.Context(), progressID)
	if err != nil {
		middleware.InternalError(w, "Could not delete student course progress", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}



