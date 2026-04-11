package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SubmissionHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewSubmissionHandler(q *db.Queries, d *sql.DB) *SubmissionHandler {
	return &SubmissionHandler{Queries: q, DB: d}
}

func (h *SubmissionHandler) CreateSubmission(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID
	userID := userCtx.UserID

	var req struct {
		AssignmentID      string `json:"assignment_id"`
		SubmissionContent string `json:"submission_content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AssignmentID == "" {
		middleware.SendError(w, "Assignment ID is required", http.StatusBadRequest)
		return
	}

	assignmentID, _ := uuid.Parse(req.AssignmentID)

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		middleware.SendError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	qtx := h.Queries.WithTx(tx)

	// Verify assignment exists and belongs to the school
	assignment, err := qtx.GetAssignmentByID(r.Context(), db.GetAssignmentByIDParams{
		AssignmentID: assignmentID,
		SchoolID:     schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Assignment not found or does not belong to your school", http.StatusNotFound)
		return
	}

	// Check if student is enrolled in the class associated with the assignment
	_, err = qtx.GetEnrollmentByStudentAndClass(r.Context(), db.GetEnrollmentByStudentAndClassParams{
		StudentID: userID,
		ClassID:   assignment.ClassID,
	})
	if err != nil {
		middleware.SendError(w, "You are not enrolled in the class for this assignment", http.StatusForbidden)
		return
	}

	// Check if student already submitted for this assignment (optional: allow multiple submissions)
	existingSubmission, err := qtx.GetSubmissionByStudentAndAssignment(r.Context(), db.GetSubmissionByStudentAndAssignmentParams{
		StudentID:    userID,
		AssignmentID: assignmentID,
	})
	if err == nil && existingSubmission.SubmissionID != uuid.Nil { // If found
		middleware.SendError(w, "You have already submitted for this assignment. Please update your existing submission.", http.StatusConflict)
		return
	}

	submission, err := qtx.CreateSubmission(r.Context(), db.CreateSubmissionParams{
		SchoolID:          schoolID,
		StudentID:         userID,
		AssignmentID:      assignmentID,
		SubmissionContent: toNullString(req.SubmissionContent),
		Status:            "Submitted",
	})
	if err != nil {
		middleware.SendError(w, "Could not submit assignment", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		middleware.SendError(w, "Could not commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(submission)
}

func (h *SubmissionHandler) GetSubmissions(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	assignmentIDStr := r.URL.Query().Get("assignmentId")
	studentIDStr := r.URL.Query().Get("studentId")

	var assignmentID uuid.NullUUID
	if assignmentIDStr != "" {
		if id, err := uuid.Parse(assignmentIDStr); err == nil {
			assignmentID = uuid.NullUUID{UUID: id, Valid: true}
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

	submissions, err := h.Queries.GetSubmissions(r.Context(), db.GetSubmissionsParams{
		SchoolID:     schoolID,
		AssignmentID: assignmentID,
		StudentID:    studentID,
	})
	if err != nil {
		middleware.SendError(w, "Could not fetch submissions", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(submissions)
}

func (h *SubmissionHandler) GetSubmissionByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	submissionID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	submission, err := h.Queries.GetSubmissionByID(r.Context(), db.GetSubmissionByIDParams{
		SubmissionID: submissionID,
		SchoolID:     schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Submission not found", http.StatusNotFound)
		return
	}

	// Authorization check
	isTeacherOrAdmin := strings.Contains(userCtx.RoleName, "Teacher") || middleware.IsAdmin(userCtx.RoleName)
	if !isTeacherOrAdmin && submission.StudentID != userCtx.UserID {
		middleware.SendError(w, "Not authorized to view this submission", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(submission)
}

func (h *SubmissionHandler) UpdateSubmissionStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	submissionID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	submission, err := h.Queries.GetSubmissionByID(r.Context(), db.GetSubmissionByIDParams{
		SubmissionID: submissionID,
		SchoolID:     schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Submission not found", http.StatusNotFound)
		return
	}

	// Verify teacher is authorized to update this submission
	isTeacherOrAdmin := strings.Contains(userCtx.RoleName, "Teacher") || middleware.IsAdmin(userCtx.RoleName)
	if !isTeacherOrAdmin || submission.TeacherID != userCtx.UserID {
		middleware.SendError(w, "Not authorized to update this submission", http.StatusForbidden)
		return
	}

	updatedSubmission, err := h.Queries.UpdateSubmissionStatus(r.Context(), db.UpdateSubmissionStatusParams{
		SubmissionID: submissionID,
		Status:       req.Status,
		SchoolID:     schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Could not update submission status", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updatedSubmission)
}
