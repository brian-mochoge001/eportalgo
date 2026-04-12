package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/brian-mochoge001/eportalgo/worker"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

type AssignmentHandler struct {
	Queries *db.Queries
	Asynq   *asynq.Client
}

func NewAssignmentHandler(q *db.Queries, asynqClient *asynq.Client) *AssignmentHandler {
	return &AssignmentHandler{Queries: q, Asynq: asynqClient}
}

func (h *AssignmentHandler) GetAssignments(w http.ResponseWriter, r *http.Request) {
	q := GetQueries(r.Context(), h.Queries)
	classIDStr := chi.URLParam(r, "class_id")
	classID, _ := uuid.Parse(classIDStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	assignments, err := q.GetAssignmentsByClass(r.Context(), db.GetAssignmentsByClassParams{
		ClassID:  classID,
		SchoolID: schoolID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not fetch assignments", err)
		return
	}

	json.NewEncoder(w).Encode(assignments)
}

func (h *AssignmentHandler) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	q := GetQueries(r.Context(), h.Queries)
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		ClassID        string `json:"class_id"`
		Title          string `json:"title"`
		Description    string `json:"description"`
		DueDate        string `json:"due_date"`
		MaxScore       string `json:"max_score"`
		AssignmentType string `json:"assignment_type"`
		FileURL        string `json:"file_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	classID, _ := uuid.Parse(req.ClassID)
	dueDate, _ := time.Parse("2006-01-02", req.DueDate)

	// Verify teacher
	academicClass, err := q.GetClassByID(r.Context(), db.GetClassByIDParams{
		ClassID:  classID,
		SchoolID: schoolID,
	})
	if err != nil || academicClass.TeacherID != userCtx.UserID {
		middleware.ForbiddenError(w, "Not authorized to post assignments to this class", err)
		return
	}

	assignment, err := q.CreateAssignment(r.Context(), db.CreateAssignmentParams{
		SchoolID:       schoolID,
		ClassID:        classID,
		TeacherID:      userCtx.UserID,
		Title:          req.Title,
		Description:    sql.NullString{String: req.Description, Valid: req.Description != ""},
		DueDate:        sql.NullTime{Time: dueDate, Valid: req.DueDate != ""},
		MaxScore:       req.MaxScore,
		AssignmentType: req.AssignmentType,
		FileUrl:        sql.NullString{String: req.FileURL, Valid: req.FileURL != ""},
	})

	if err != nil {
		middleware.InternalError(w, "Could not create assignment", err)
		return
	}

	// Notify students (Async via Asynq)
	payload, _ := json.Marshal(worker.AssignmentNotificationPayload{
		SchoolID:     schoolID,
		ClassID:      classID,
		TeacherID:    userCtx.UserID,
		Title:        assignment.Title,
		DueDate:      dueDate.Format("2006-01-02"),
		AssignmentID: assignment.AssignmentID,
	})
	task := asynq.NewTask(worker.TypeAssignmentNotification, payload)
	if _, err := h.Asynq.Enqueue(task); err != nil {
		log.Printf("could not enqueue task: %v", err)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(assignment)
}

func (h *AssignmentHandler) UpdateAssignment(w http.ResponseWriter, r *http.Request) {
	q := GetQueries(r.Context(), h.Queries)
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		Title          string `json:"title"`
		Description    string `json:"description"`
		DueDate        string `json:"due_date"`
		MaxScore       string `json:"max_score"`
		AssignmentType string `json:"assignment_type"`
		FileURL        string `json:"file_url"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	dueDate, _ := time.Parse("2006-01-02", req.DueDate)

	updated, err := q.UpdateAssignment(r.Context(), db.UpdateAssignmentParams{
		AssignmentID:   id,
		Title:          req.Title,
		Description:    sql.NullString{String: req.Description, Valid: req.Description != ""},
		DueDate:        sql.NullTime{Time: dueDate, Valid: req.DueDate != ""},
		MaxScore:       req.MaxScore,
		AssignmentType: req.AssignmentType,
		FileUrl:        sql.NullString{String: req.FileURL, Valid: req.FileURL != ""},
		SchoolID:       schoolID,
	})

	if err != nil {
		middleware.InternalError(w, "Could not update assignment", err)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func (h *AssignmentHandler) DeleteAssignment(w http.ResponseWriter, r *http.Request) {
	q := GetQueries(r.Context(), h.Queries)
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	err := q.DeleteAssignment(r.Context(), db.DeleteAssignmentParams{
		AssignmentID: id,
		SchoolID:     schoolID,
		TeacherID:    userCtx.UserID,
	})

	if err != nil {
		middleware.InternalError(w, "Could not delete assignment", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}


