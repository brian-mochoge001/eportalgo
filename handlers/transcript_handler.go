package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TranscriptHandler struct {
	Queries *db.Queries
}

func NewTranscriptHandler(q *db.Queries) *TranscriptHandler {
	return &TranscriptHandler{Queries: q}
}

func (h *TranscriptHandler) CreateTranscript(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		StudentID     string `json:"student_id"`
		AcademicYear  string `json:"academic_year"`
		CumulativeGPA string `json:"cumulative_gpa"`
		TranscriptData string `json:"transcript_data"` // Assuming JSON string
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	studentID, _ := uuid.Parse(req.StudentID)

	transcript, err := h.Queries.CreateTranscript(r.Context(), db.CreateTranscriptParams{
		SchoolID:       schoolID,
		StudentID:      studentID,
		AcademicYear:   req.AcademicYear,
		CumulativeGpa:  sql.NullString{String: req.CumulativeGPA, Valid: req.CumulativeGPA != ""},
		TranscriptData: json.RawMessage(req.TranscriptData),
		IssuedByUserID: uuid.NullUUID{UUID: userCtx.UserID, Valid: true},
	})

	if err != nil {
		middleware.SendError(w, "Could not create transcript", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transcript)
}

func (h *TranscriptHandler) GetTranscripts(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID
	studentIDStr := r.URL.Query().Get("studentId")

	var studentID uuid.NullUUID
	if userCtx.RoleName == "Student" {
		studentID = uuid.NullUUID{UUID: userCtx.UserID, Valid: true}
	} else if studentIDStr != "" {
		if id, err := uuid.Parse(studentIDStr); err == nil {
			studentID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	transcripts, err := h.Queries.GetTranscripts(r.Context(), db.GetTranscriptsParams{
		SchoolID:  schoolID,
		StudentID: studentID,
	})
	if err != nil {
		middleware.SendError(w, "Could not fetch transcripts", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(transcripts)
}

func (h *TranscriptHandler) GetTranscriptByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	transcriptID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	transcript, err := h.Queries.GetTranscriptByID(r.Context(), db.GetTranscriptByIDParams{
		TranscriptID: transcriptID,
		SchoolID:     schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Transcript not found", http.StatusNotFound)
		return
	}

	if userCtx.RoleName == "Student" && transcript.StudentID != userCtx.UserID {
		middleware.SendError(w, "Not authorized to view this transcript", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(transcript)
}

func (h *TranscriptHandler) UpdateTranscript(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	transcriptID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		AcademicYear string `json:"academic_year"`
		CumulativeGPA string `json:"cumulative_gpa"`
		TranscriptData string `json:"transcript_data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	params := db.UpdateTranscriptParams{
		TranscriptID: transcriptID,
		SchoolID:     schoolID,
	}

	// Fetch existing to preserve fields
	existing, err := h.Queries.GetTranscriptByID(r.Context(), db.GetTranscriptByIDParams{
		TranscriptID: transcriptID,
		SchoolID:     schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Transcript not found", http.StatusNotFound)
		return
	}

	params.AcademicYear = existing.AcademicYear
	params.CumulativeGpa = existing.CumulativeGpa
	params.TranscriptData = existing.TranscriptData

	if req.AcademicYear != "" {
		params.AcademicYear = req.AcademicYear
	}
	if req.CumulativeGPA != "" {
		params.CumulativeGpa = sql.NullString{String: req.CumulativeGPA, Valid: true}
	}
	if req.TranscriptData != "" {
		params.TranscriptData = json.RawMessage(req.TranscriptData)
	}

	updated, err := h.Queries.UpdateTranscript(r.Context(), params)
	if err != nil {
		middleware.SendError(w, "Could not update transcript", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func (h *TranscriptHandler) DeleteTranscript(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	transcriptID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	err := h.Queries.DeleteTranscript(r.Context(), db.DeleteTranscriptParams{
		TranscriptID: transcriptID,
		SchoolID:     schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Could not delete transcript", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
