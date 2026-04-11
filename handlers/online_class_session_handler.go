package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type OnlineClassSessionHandler struct {
	Queries *db.Queries
}

func NewOnlineClassSessionHandler(q *db.Queries) *OnlineClassSessionHandler {
	return &OnlineClassSessionHandler{Queries: q}
}

func (h *OnlineClassSessionHandler) CreateOnlineClassSession(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID
	userID := userCtx.UserID

	var req struct {
		ClassID        string `json:"class_id"`
		SessionTitle   string `json:"session_title"`
		StartTime      string `json:"start_time"`
		EndTime        string `json:"end_time"`
		MeetingLink    string `json:"meeting_link"`
		Description    string `json:"description"`
		RecordingLink  string `json:"recording_link"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ClassID == "" || req.SessionTitle == "" || req.StartTime == "" || req.EndTime == "" || req.MeetingLink == "" {
		middleware.SendError(w, "Class ID, session title, start time, end time, and meeting link are required", http.StatusBadRequest)
		return
	}

	classID, err := uuid.Parse(req.ClassID)
	if err != nil {
		middleware.SendError(w, "Invalid class ID", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		middleware.SendError(w, "Invalid start time format", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		middleware.SendError(w, "Invalid end time format", http.StatusBadRequest)
		return
	}

	session, err := h.Queries.CreateOnlineClassSession(r.Context(), db.CreateOnlineClassSessionParams{
		SchoolID:      schoolID,
		ClassID:       classID,
		TeacherID:     userID,
		SessionTitle:  req.SessionTitle,
		StartTime:     startTime,
		EndTime:       endTime,
		MeetingLink:   req.MeetingLink,
		Description:   sql.NullString{String: req.Description, Valid: req.Description != ""},
		RecordingLink: sql.NullString{String: req.RecordingLink, Valid: req.RecordingLink != ""},
	})

	if err != nil {
		middleware.SendError(w, "Could not create online class session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(session)
}

func (h *OnlineClassSessionHandler) GetOnlineClassSessions(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	classIDStr := r.URL.Query().Get("classId")
	teacherIDStr := r.URL.Query().Get("teacherId")

	var classID uuid.NullUUID
	if classIDStr != "" {
		if id, err := uuid.Parse(classIDStr); err == nil {
			classID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	var teacherID uuid.NullUUID
	if teacherIDStr != "" {
		if id, err := uuid.Parse(teacherIDStr); err == nil {
			teacherID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	// Teachers can only see their own sessions unless they are an admin
	if userCtx.RoleName == "Teacher" {
		teacherID = uuid.NullUUID{UUID: userCtx.UserID, Valid: true}
	}

	sessions, err := h.Queries.GetOnlineClassSessions(r.Context(), db.GetOnlineClassSessionsParams{
		SchoolID:  schoolID,
		ClassID:   classID,
		TeacherID: teacherID,
	})
	if err != nil {
		middleware.SendError(w, "Could not fetch online class sessions", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(sessions)
}

func (h *OnlineClassSessionHandler) GetOnlineClassSessionByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.SendError(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	session, err := h.Queries.GetOnlineClassSessionByID(r.Context(), db.GetOnlineClassSessionByIDParams{
		SessionID: sessionID,
		SchoolID:  schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Online class session not found", http.StatusNotFound)
		return
	}

	// Ensure teacher can only access their own sessions unless they are an admin
	if userCtx.RoleName == "Teacher" && session.TeacherID != userCtx.UserID {
		middleware.SendError(w, "Not authorized to view this online class session", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(session)
}

func (h *OnlineClassSessionHandler) UpdateOnlineClassSession(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.SendError(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		SessionTitle  string `json:"session_title"`
		StartTime     string `json:"start_time"`
		EndTime       string `json:"end_time"`
		MeetingLink   string `json:"meeting_link"`
		Description   string `json:"description"`
		RecordingLink string `json:"recording_link"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	existingSession, err := h.Queries.GetOnlineClassSessionByID(r.Context(), db.GetOnlineClassSessionByIDParams{
		SessionID: sessionID,
		SchoolID:  schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Online class session not found", http.StatusNotFound)
		return
	}

	// Ensure teacher can only update their own sessions unless they are an admin
	if userCtx.RoleName == "Teacher" && existingSession.TeacherID != userCtx.UserID {
		middleware.SendError(w, "Not authorized to update this online class session", http.StatusForbidden)
		return
	}

	params := db.UpdateOnlineClassSessionParams{
		SessionID:     sessionID,
		SchoolID:      schoolID,
		TeacherID:     existingSession.TeacherID,
		SessionTitle:  existingSession.SessionTitle,
		StartTime:     existingSession.StartTime,
		EndTime:       existingSession.EndTime,
		MeetingLink:   existingSession.MeetingLink,
		Description:   existingSession.Description,
		RecordingLink: existingSession.RecordingLink,
	}

	if req.SessionTitle != "" {
		params.SessionTitle = req.SessionTitle
	}
	if req.StartTime != "" {
		if t, err := time.Parse(time.RFC3339, req.StartTime); err == nil {
			params.StartTime = t
		}
	}
	if req.EndTime != "" {
		if t, err := time.Parse(time.RFC3339, req.EndTime); err == nil {
			params.EndTime = t
		}
	}
	if req.MeetingLink != "" {
		params.MeetingLink = req.MeetingLink
	}
	if req.Description != "" {
		params.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.RecordingLink != "" {
		params.RecordingLink = sql.NullString{String: req.RecordingLink, Valid: true}
	}

	updatedSession, err := h.Queries.UpdateOnlineClassSession(r.Context(), params)
	if err != nil {
		middleware.SendError(w, "Could not update online class session", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updatedSession)
}

func (h *OnlineClassSessionHandler) DeleteOnlineClassSession(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.SendError(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	existingSession, err := h.Queries.GetOnlineClassSessionByID(r.Context(), db.GetOnlineClassSessionByIDParams{
		SessionID: sessionID,
		SchoolID:  schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Online class session not found", http.StatusNotFound)
		return
	}

	// Ensure teacher can only delete their own sessions unless they are an admin
	if userCtx.RoleName == "Teacher" && existingSession.TeacherID != userCtx.UserID {
		middleware.SendError(w, "Not authorized to delete this online class session", http.StatusForbidden)
		return
	}

	err = h.Queries.DeleteOnlineClassSession(r.Context(), db.DeleteOnlineClassSessionParams{
		SessionID: sessionID,
		SchoolID:  schoolID,
		TeacherID: existingSession.TeacherID,
	})
	if err != nil {
		middleware.SendError(w, "Could not delete online class session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
