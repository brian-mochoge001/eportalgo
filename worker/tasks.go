package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/sqlc-dev/pqtype"
)

const (
	TypeAssignmentNotification = "notification:assignment"
	TypeAuditLog               = "audit:log"
)

type AssignmentNotificationPayload struct {
	SchoolID     uuid.UUID
	ClassID      uuid.UUID
	TeacherID    uuid.UUID
	Title        string
	DueDate      string
	AssignmentID uuid.UUID
}

type AuditLogPayload struct {
	SchoolID   uuid.NullUUID
	UserID     uuid.UUID
	Action     string
	NewValue   pqtype.NullRawMessage
	IpAddress  string
	UserAgent  string
}

type TaskHandler struct {
	Queries *db.Queries
}

func NewTaskHandler(q *db.Queries) *TaskHandler {
	return &TaskHandler{Queries: q}
}

func (h *TaskHandler) HandleAssignmentNotification(ctx context.Context, t *asynq.Task) error {
	var p AssignmentNotificationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("Processing assignment notification for class %s", p.ClassID)

	students, err := h.Queries.GetEnrollmentsByClass(ctx, p.ClassID)
	if err != nil {
		return err
	}

	notification, err := h.Queries.CreateNotification(ctx, db.CreateNotificationParams{
		SchoolID:         p.SchoolID,
		SenderID:         uuid.NullUUID{UUID: p.TeacherID, Valid: true},
		NotificationType: db.NotificationTypeANNOUNCEMENT,
		Title:            "New Assignment: " + p.Title,
		Message:          fmt.Sprintf("%s: due on %s", p.Title, p.DueDate),
		LinkUrl:          sql.NullString{String: "/assignments/" + p.AssignmentID.String(), Valid: true},
	})
	if err != nil {
		return err
	}

	for _, studentID := range students {
		h.Queries.CreateNotificationRecipient(ctx, db.CreateNotificationRecipientParams{
			NotificationID: notification.NotificationID,
			RecipientID:    studentID,
		})
	}

	return nil
}

func (h *TaskHandler) HandleAuditLog(ctx context.Context, t *asynq.Task) error {
	var p AuditLogPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	_, err := h.Queries.CreateAuditLog(ctx, db.CreateAuditLogParams{
		SchoolID:   p.SchoolID,
		UserID:     uuid.NullUUID{UUID: p.UserID, Valid: true},
		Action:     p.Action,
		EntityType: "Unknown",
		EntityID:   uuid.NullUUID{Valid: false},
		OldValue:   pqtype.NullRawMessage{Valid: false},
		NewValue:   p.NewValue,
		IpAddress:  sql.NullString{String: p.IpAddress, Valid: true},
		UserAgent:  sql.NullString{String: p.UserAgent, Valid: true},
	})
	return err
}
