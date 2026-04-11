package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type NotificationHandler struct {
	Queries *db.Queries
}

func NewNotificationHandler(q *db.Queries) *NotificationHandler {
	return &NotificationHandler{Queries: q}
}

func (h *NotificationHandler) CreateNotification(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		SenderID        string `json:"sender_id"`
		RecipientID     string `json:"recipient_id"`
		NotificationType string `json:"notification_type"`
		Title           string `json:"title"`
		Message         string `json:"message"`
		LinkURL         string `json:"link_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	senderID, _ := uuid.Parse(req.SenderID)
	recipientID, _ := uuid.Parse(req.RecipientID)

	
	notification, err := h.Queries.CreateNotification(r.Context(), db.CreateNotificationParams{
		SchoolID:         schoolID,
		SenderID:         uuid.NullUUID{UUID: senderID, Valid: true},
		NotificationType: db.NotificationType(req.NotificationType),
		Title:            req.Title,
		Message:          req.Message,
		LinkUrl:          toNullString(req.LinkURL),
	})

	if err != nil {
		middleware.InternalError(w, "Could not create notification", err)
		return
	}

	_, err = h.Queries.CreateNotificationRecipient(r.Context(), db.CreateNotificationRecipientParams{
		NotificationID: notification.NotificationID,
		RecipientID:    recipientID,
	})

	if err != nil {
		middleware.InternalError(w, "Could not create notification recipient", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(notification)
}

func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	recipientID := userCtx.UserID

	notifications, err := h.Queries.GetNotificationsByRecipient(r.Context(), recipientID)
	if err != nil {
		middleware.InternalError(w, "Could not fetch notifications", err)
		return
	}

	json.NewEncoder(w).Encode(notifications)
}

func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id") // Assuming URL param is 'id' for notification_id
	notificationID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())

	err := h.Queries.MarkNotificationAsRead(r.Context(), db.MarkNotificationAsReadParams{
		NotificationID: notificationID,
		RecipientID:    userCtx.UserID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not mark notification as read", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Notification marked as read"})
}



