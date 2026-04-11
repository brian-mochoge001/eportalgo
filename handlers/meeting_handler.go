package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type MeetingHandler struct {
	Queries *db.Queries
}

func NewMeetingHandler(q *db.Queries) *MeetingHandler {
	return &MeetingHandler{Queries: q}
}

func (h *MeetingHandler) CreateMeeting(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		Title           string   `json:"title"`
		Agenda          string   `json:"agenda"`
		MeetingDate     string   `json:"meeting_date"`
		DurationMinutes string   `json:"duration_minutes"`
		Location        string   `json:"location"`
		MeetingType     string   `json:"meeting_type"`
		OrganizerID     string   `json:"organizer_id"`
		Attendees       []string `json:"attendees"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	if req.Title == "" || req.MeetingDate == "" || req.MeetingType == "" {
		middleware.ValidationError(w, "Title, meeting date, and meeting type are required", nil)
		return
	}

	meetingDate, err := time.Parse(time.RFC3339, req.MeetingDate)
	if err != nil {
		middleware.ValidationError(w, "Invalid meeting date format", err)
		return
	}

	var durationMinutes sql.NullInt32
	if req.DurationMinutes != "" {
		if d, err := strconv.Atoi(req.DurationMinutes); err == nil {
			durationMinutes = sql.NullInt32{Int32: int32(d), Valid: true}
		}
	}

	organizerID := uuid.NullUUID{UUID: userCtx.UserID, Valid: true}
	if req.OrganizerID != "" {
		if id, err := uuid.Parse(req.OrganizerID); err == nil {
			organizerID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	meeting, err := h.Queries.CreateMeeting(r.Context(), db.CreateMeetingParams{
		SchoolID:        schoolID,
		Title:           req.Title,
		Agenda:          sql.NullString{String: req.Agenda, Valid: req.Agenda != ""},
		MeetingDate:     meetingDate,
		DurationMinutes: durationMinutes,
		Location:        sql.NullString{String: req.Location, Valid: req.Location != ""},
		MeetingType:     req.MeetingType,
		OrganizerID:     organizerID,
	})

	if err != nil {
		middleware.InternalError(w, "Could not create meeting", err)
		return
	}

	// Add attendees
	for _, attendeeIDStr := range req.Attendees {
		if attendeeID, err := uuid.Parse(attendeeIDStr); err == nil {
			h.Queries.AddMeetingAttendee(r.Context(), db.AddMeetingAttendeeParams{
				MeetingID: meeting.MeetingID,
				UserID:    attendeeID,
				SchoolID:  schoolID,
			})
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(meeting)
}

func (h *MeetingHandler) GetMeetings(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	meetings, err := h.Queries.GetMeetingsBySchool(r.Context(), schoolID)
	if err != nil {
		middleware.InternalError(w, "Could not fetch meetings", err)
		return
	}

	json.NewEncoder(w).Encode(meetings)
}

func (h *MeetingHandler) GetMeetingByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	meetingID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid meeting ID", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	meeting, err := h.Queries.GetMeetingByID(r.Context(), db.GetMeetingByIDParams{
		MeetingID: meetingID,
		SchoolID:  schoolID,
	})
	if err != nil {
		middleware.NotFoundError(w, "Meeting not found", err)
		return
	}

	attendees, _ := h.Queries.GetMeetingAttendees(r.Context(), meetingID)

	response := struct {
		db.GetMeetingByIDRow
		Attendees interface{} `json:"attendees"`
	}{
		GetMeetingByIDRow: meeting,
		Attendees:         attendees,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *MeetingHandler) UpdateMeeting(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	meetingID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid meeting ID", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		Title           string `json:"title"`
		Agenda          string `json:"agenda"`
		MeetingDate     string `json:"meeting_date"`
		DurationMinutes string `json:"duration_minutes"`
		Location        string `json:"location"`
		MeetingType     string `json:"meeting_type"`
		OrganizerID     string `json:"organizer_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	existingMeeting, err := h.Queries.GetMeetingByID(r.Context(), db.GetMeetingByIDParams{
		MeetingID: meetingID,
		SchoolID:  schoolID,
	})
	if err != nil {
		middleware.NotFoundError(w, "Meeting not found", err)
		return
	}

	params := db.UpdateMeetingParams{
		MeetingID:       meetingID,
		SchoolID:        schoolID,
		Title:           existingMeeting.Title,
		Agenda:          existingMeeting.Agenda,
		MeetingDate:     existingMeeting.MeetingDate,
		DurationMinutes: existingMeeting.DurationMinutes,
		Location:        existingMeeting.Location,
		MeetingType:     existingMeeting.MeetingType,
		OrganizerID:     existingMeeting.OrganizerID,
	}

	if req.Title != "" {
		params.Title = req.Title
	}
	if req.Agenda != "" {
		params.Agenda = sql.NullString{String: req.Agenda, Valid: true}
	}
	if req.MeetingDate != "" {
		if t, err := time.Parse(time.RFC3339, req.MeetingDate); err == nil {
			params.MeetingDate = t
		}
	}
	if req.DurationMinutes != "" {
		if d, err := strconv.Atoi(req.DurationMinutes); err == nil {
			params.DurationMinutes = sql.NullInt32{Int32: int32(d), Valid: true}
		}
	}
	if req.Location != "" {
		params.Location = sql.NullString{String: req.Location, Valid: true}
	}
	if req.MeetingType != "" {
		params.MeetingType = req.MeetingType
	}
	if req.OrganizerID != "" {
		if id, err := uuid.Parse(req.OrganizerID); err == nil {
			params.OrganizerID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	updatedMeeting, err := h.Queries.UpdateMeeting(r.Context(), params)
	if err != nil {
		middleware.InternalError(w, "Could not update meeting", err)
		return
	}

	json.NewEncoder(w).Encode(updatedMeeting)
}

func (h *MeetingHandler) DeleteMeeting(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	meetingID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid meeting ID", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	err = h.Queries.DeleteMeeting(r.Context(), db.DeleteMeetingParams{
		MeetingID: meetingID,
		SchoolID:  schoolID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not delete meeting", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MeetingHandler) AddMeetingAttendees(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	meetingID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid meeting ID", err)
		return
	}

	var req struct {
		UserIDs []string `json:"user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	for _, userIDStr := range req.UserIDs {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			h.Queries.AddMeetingAttendee(r.Context(), db.AddMeetingAttendeeParams{
				MeetingID: meetingID,
				UserID:    userID,
				SchoolID:  schoolID,
			})
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Attendees added successfully"})
}

func (h *MeetingHandler) RemoveMeetingAttendees(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	meetingID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid meeting ID", err)
		return
	}

	var req struct {
		UserIDs []string `json:"user_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	for _, userIDStr := range req.UserIDs {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			h.Queries.RemoveMeetingAttendee(r.Context(), db.RemoveMeetingAttendeeParams{
				MeetingID: meetingID,
				UserID:    userID,
				SchoolID:  schoolID,
			})
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Attendees removed successfully"})
}



