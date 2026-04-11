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

type ChatHandler struct {
	Queries *db.Queries
}

func NewChatHandler(q *db.Queries) *ChatHandler {
	return &ChatHandler{Queries: q}
}

func (h *ChatHandler) GetChatRooms(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())

	rooms, err := h.Queries.GetChatRoomsByUser(r.Context(), userCtx.UserID)
	if err != nil {
		middleware.SendError(w, "Could not fetch chat rooms", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(rooms)
}

func (h *ChatHandler) GetChatMessages(w http.ResponseWriter, r *http.Request) {
	roomIDStr := chi.URLParam(r, "chat_room_id")
	roomID, _ := uuid.Parse(roomIDStr)

	messages, err := h.Queries.GetChatMessagesByRoom(r.Context(), roomID)
	if err != nil {
		middleware.SendError(w, "Could not fetch messages", http.StatusInternalServerError)
		return
	}

	// In a real app, we would decrypt here. For now, sending as is.
	json.NewEncoder(w).Encode(messages)
}

func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		ChatRoomID      string `json:"chat_room_id"`
		MessageText     string `json:"message_text"`
		AttachmentURL   string `json:"attachment_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	roomID, _ := uuid.Parse(req.ChatRoomID)

	// Verify membership
	participants, err := h.Queries.GetChatParticipants(r.Context(), roomID)
	isMember := false
	for _, p := range participants {
		if p.UserID == userCtx.UserID {
			isMember = true
			break
		}
	}

	if !isMember {
		middleware.SendError(w, "Not a member of this chat room", http.StatusForbidden)
		return
	}

	room, _ := h.Queries.GetChatRoom(r.Context(), roomID)

	// File type check
	if req.AttachmentURL != "" {
		parts := strings.Split(req.AttachmentURL, ".")
		ext := parts[len(parts)-1]
		allowed := false
		for _, a := range room.AllowedFileTypes {
			if a == ext {
				allowed = true
				break
			}
		}
		if !allowed && len(room.AllowedFileTypes) > 0 {
			middleware.SendError(w, "File type not allowed", http.StatusBadRequest)
			return
		}
	}

	// In a real app, encrypt MessageText here.
	message, err := h.Queries.CreateChatMessage(r.Context(), db.CreateChatMessageParams{
		ChatRoomID:    roomID,
		SenderID:      userCtx.UserID,
		SchoolID:      schoolID,
		MessageText:   req.MessageText,
		AttachmentUrl: sql.NullString{String: req.AttachmentURL, Valid: req.AttachmentURL != ""},
	})

	if err != nil {
		middleware.SendError(w, "Could not send message", http.StatusInternalServerError)
		return
	}

	// Real-time broadcast would go here (e.g. via NATS, Redis PubSub, or WebSockets)

	json.NewEncoder(w).Encode(message)
}

func isTeacherOrDeptHead(role string) bool {
	return role == "Teacher" || role == "Department Head"
}
