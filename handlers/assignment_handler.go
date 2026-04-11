package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AssignmentHandler struct {
	Queries *db.Queries
}

func NewAssignmentHandler(q *db.Queries) *AssignmentHandler {
	return &AssignmentHandler{Queries: q}
}

func (h *AssignmentHandler) GetAssignments(w http.ResponseWriter, r *http.Request) {
	classIDStr := chi.URLParam(r, "class_id")
	classID, _ := uuid.Parse(classIDStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	assignments, err := h.Queries.GetAssignmentsByClass(r.Context(), db.GetAssignmentsByClassParams{
		ClassID:  classID,
		SchoolID: schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Could not fetch assignments", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(assignments)
}

func (h *AssignmentHandler) CreateAssignment(w http.ResponseWriter, r *http.Request) {
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
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	classID, _ := uuid.Parse(req.ClassID)
	dueDate, _ := time.Parse("2006-01-02", req.DueDate)

	// Verify teacher
	academicClass, err := h.Queries.GetClassByID(r.Context(), db.GetClassByIDParams{
		ClassID:  classID,
		SchoolID: schoolID,
	})
	if err != nil || academicClass.TeacherID != userCtx.UserID {
		middleware.SendError(w, "Not authorized to post assignments to this class", http.StatusForbidden)
		return
	}

	assignment, err := h.Queries.CreateAssignment(r.Context(), db.CreateAssignmentParams{
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
		middleware.SendError(w, "Could not create assignment", http.StatusInternalServerError)
		return
	}

	// Notify students (Async)
	go func() {
		students, _ := h.Queries.GetEnrollmentsByClass(context.Background(), classID)
		notification, _ := h.Queries.CreateNotification(context.Background(), db.CreateNotificationParams{
			SchoolID:         schoolID,
			SenderID:         uuid.NullUUID{UUID: userCtx.UserID, Valid: true},
			NotificationType: db.NotificationTypeANNOUNCEMENT,
			Title:            "New Assignment: " + assignment.Title,
			Message:          fmt.Sprintf("%s: due on %s", assignment.Title, dueDate.Format("2006-01-02")),
			LinkUrl:          sql.NullString{String: "/assignments/" + assignment.AssignmentID.String(), Valid: true},
		})

		for _, studentID := range students {
			h.Queries.CreateNotificationRecipient(context.Background(), db.CreateNotificationRecipientParams{
				NotificationID: notification.NotificationID,
				RecipientID:    studentID,
			})
		}
	}()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(assignment)
}

func (h *AssignmentHandler) UpdateAssignment(w http.ResponseWriter, r *http.Request) {
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

	updated, err := h.Queries.UpdateAssignment(r.Context(), db.UpdateAssignmentParams{
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
		middleware.SendError(w, "Could not update assignment", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func (h *AssignmentHandler) DeleteAssignment(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	err := h.Queries.DeleteAssignment(r.Context(), db.DeleteAssignmentParams{
		AssignmentID: id,
		SchoolID:     schoolID,
		TeacherID:    userCtx.UserID,
	})

	if err != nil {
		middleware.SendError(w, "Could not delete assignment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
