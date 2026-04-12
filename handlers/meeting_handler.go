package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/brian-mochoge001/eportalgo/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type MeetingHandler struct {
	Queries        *db.Queries
	MeetingService *services.MeetingService
}

func NewMeetingHandler(q *db.Queries, s *services.MeetingService) *MeetingHandler {
	return &MeetingHandler{Queries: q, MeetingService: s}
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

	meetingDate, err := time.Parse(time.RFC3339, req.MeetingDate)
	if err != nil {
		middleware.ValidationError(w, "Invalid meeting date format", err)
		return
	}

	var durationMinutes int32
	if d, err := strconv.Atoi(req.DurationMinutes); err == nil {
		durationMinutes = int32(d)
	}

	organizerID := userCtx.UserID
	if req.OrganizerID != "" {
		if id, err := uuid.Parse(req.OrganizerID); err == nil {
			organizerID = id
		}
	}

	var attendeeIDs []uuid.UUID
	for _, idStr := range req.Attendees {
		if id, err := uuid.Parse(idStr); err == nil {
			attendeeIDs = append(attendeeIDs, id)
		}
	}

	meeting, err := h.MeetingService.CreateMeeting(r.Context(), services.CreateMeetingParams{
		SchoolID:        schoolID,
		Title:           req.Title,
		Agenda:          req.Agenda,
		MeetingDate:     meetingDate,
		DurationMinutes: durationMinutes,
		Location:        req.Location,
		MeetingType:     req.MeetingType,
		OrganizerID:     organizerID,
		AttendeeIDs:     attendeeIDs,
	})

	if err != nil {
		middleware.InternalError(w, err.Error(), err)
		return
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
	meetingID, _ := uuid.Parse(idStr)

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
	meetingID, _ := uuid.Parse(idStr)

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

	params := services.UpdateMeetingParams{
		MeetingID:   meetingID,
		SchoolID:    schoolID,
		Title:       req.Title,
		Agenda:      req.Agenda,
		Location:    req.Location,
		MeetingType: req.MeetingType,
	}

	if req.MeetingDate != "" {
		if t, err := time.Parse(time.RFC3339, req.MeetingDate); err == nil {
			params.MeetingDate = &t
		}
	}
	if req.DurationMinutes != "" {
		if d, err := strconv.Atoi(req.DurationMinutes); err == nil {
			duration := int32(d)
			params.DurationMinutes = &duration
		}
	}
	if req.OrganizerID != "" {
		if id, err := uuid.Parse(req.OrganizerID); err == nil {
			params.OrganizerID = &id
		}
	}

	updated, err := h.MeetingService.UpdateMeeting(r.Context(), params)
	if err != nil {
		middleware.InternalError(w, err.Error(), err)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func (h *MeetingHandler) DeleteMeeting(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	meetingID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	err := h.Queries.DeleteMeeting(r.Context(), db.DeleteMeetingParams{
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
	meetingID, _ := uuid.Parse(idStr)

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
	meetingID, _ := uuid.Parse(idStr)

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
