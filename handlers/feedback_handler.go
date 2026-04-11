package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type FeedbackHandler struct {
	Queries *db.Queries
}

func NewFeedbackHandler(q *db.Queries) *FeedbackHandler {
	return &FeedbackHandler{Queries: q}
}

func (h *FeedbackHandler) CreateFeedback(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID

	var req struct {
		Subject      string `json:"subject"`
		Message      string `json:"message"`
		Rating       int32  `json:"rating"`
		FeedbackType string `json:"feedback_type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	feedback, err := h.Queries.CreateFeedback(r.Context(), db.CreateFeedbackParams{
		SchoolID:     schoolID,
		UserID:       userCtx.UserID,
		Subject:      toNullString(req.Subject),
		Message:      req.Message,
		Rating:       toNullInt32(&req.Rating),
		FeedbackType: req.FeedbackType,
	})

	if err != nil {
		middleware.InternalError(w, "Could not create feedback", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(feedback)
}

func (h *FeedbackHandler) ListFeedbacks(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	// Check if user is authorized to see all feedbacks
	isSuperAdmin := false
	for _, role := range []string{"Developer", "DB Manager", "Data Analyst", "Support Staff", "Executive Administrator"} {
		if userCtx.RoleName == role {
			isSuperAdmin = true
			break
		}
	}

	feedbacks, err := h.Queries.ListFeedbacks(r.Context(), db.ListFeedbacksParams{
		SchoolID:     uuid.NullUUID{UUID: schoolID, Valid: true},
		IsSuperAdmin: isSuperAdmin,
	})
	if err != nil {
		middleware.InternalError(w, "Could not fetch feedbacks", err)
		return
	}

	json.NewEncoder(w).Encode(feedbacks)
}

func (h *FeedbackHandler) GetFeedbackByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	feedbackID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())

	feedback, err := h.Queries.GetFeedbackByID(r.Context(), feedbackID)
	if err != nil {
		middleware.NotFoundError(w, "Feedback not found", err)
		return
	}

	isSuperAdmin := false
	for _, role := range []string{"Developer", "DB Manager", "Data Analyst", "Support Staff", "Executive Administrator"} {
		if userCtx.RoleName == role {
			isSuperAdmin = true
			break
		}
	}

	if !isSuperAdmin && feedback.UserID != userCtx.UserID && feedback.SchoolID.UUID != userCtx.SchoolID.UUID {
		middleware.ForbiddenError(w, "Not authorized to view this feedback", err)
		return
	}

	json.NewEncoder(w).Encode(feedback)
}

func (h *FeedbackHandler) UpdateFeedback(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	feedbackID, _ := uuid.Parse(idStr)

	var req struct {
		Subject      string `json:"subject"`
		Message      string `json:"message"`
		Rating       int32  `json:"rating"`
		FeedbackType string `json:"feedback_type"`
		Status       string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	updated, err := h.Queries.UpdateFeedback(r.Context(), db.UpdateFeedbackParams{
		FeedbackID:   feedbackID,
		Subject:      toNullString(req.Subject),
		Message:      req.Message,
		Rating:       toNullInt32(&req.Rating),
		FeedbackType: req.FeedbackType,
		Status:       req.Status,
	})

	if err != nil {
		middleware.InternalError(w, "Could not update feedback", err)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func (h *FeedbackHandler) DeleteFeedback(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	feedbackID, _ := uuid.Parse(idStr)

	err := h.Queries.DeleteFeedback(r.Context(), feedbackID)
	if err != nil {
		middleware.InternalError(w, "Could not delete feedback", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}



