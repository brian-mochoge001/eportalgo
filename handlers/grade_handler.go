package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type GradeHandler struct {
	Queries *db.Queries
}

func NewGradeHandler(q *db.Queries) *GradeHandler {
	return &GradeHandler{Queries: q}
}

func (h *GradeHandler) GetGradesBySubmission(w http.ResponseWriter, r *http.Request) {
	submissionIDStr := chi.URLParam(r, "submission_id")
	submissionID, _ := uuid.Parse(submissionIDStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	grades, err := h.Queries.GetGradesBySubmission(r.Context(), db.GetGradesBySubmissionParams{
		SubmissionID: submissionID,
		SchoolID:     schoolID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not fetch grades", err)
		return
	}

	json.NewEncoder(w).Encode(grades)
}

func (h *GradeHandler) CreateGrade(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		SubmissionID string `json:"submission_id"`
		Score        string `json:"score"`
		Feedback     string `json:"feedback"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	submissionID, _ := uuid.Parse(req.SubmissionID)

	grade, err := h.Queries.CreateGrade(r.Context(), db.CreateGradeParams{
		SchoolID:       schoolID,
		SubmissionID:   submissionID,
		GradedByUserID: uuid.NullUUID{UUID: userCtx.UserID, Valid: true},
		Score:          req.Score,
		Feedback:       toNullString(req.Feedback),
	})

	if err != nil {
		middleware.InternalError(w, "Could not create grade", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(grade)
}



