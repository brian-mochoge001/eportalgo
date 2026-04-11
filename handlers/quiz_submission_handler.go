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

type QuizSubmissionHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewQuizSubmissionHandler(q *db.Queries, d *sql.DB) *QuizSubmissionHandler {
	return &QuizSubmissionHandler{Queries: q, DB: d}
}

func (h *QuizSubmissionHandler) CreateQuizSubmission(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())

	var req struct {
		QuizID  string `json:"quiz_id"`
		Answers []struct {
			QuestionID        string `json:"question_id"`
			StudentAnswerText string `json:"student_answer_text"`
			SelectedOptionID  string `json:"selected_option_id"`
		} `json:"answers"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	quizID, _ := uuid.Parse(req.QuizID)

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		middleware.SendError(w, "Could not start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	qtx := h.Queries.WithTx(tx)

	// Create Submission
	submission, err := qtx.CreateQuizSubmission(r.Context(), db.CreateQuizSubmissionParams{
		QuizID:    quizID,
		StudentID: userCtx.UserID,
		Status:    "completed",
	})
	if err != nil {
		middleware.SendError(w, "Could not create submission", http.StatusInternalServerError)
		return
	}

	// Create Answers
	for _, aReq := range req.Answers {
		questionID, _ := uuid.Parse(aReq.QuestionID)
		selectedOptionID := toNullUUID(aReq.SelectedOptionID)

		_, err = qtx.CreateQuizAnswer(r.Context(), db.CreateQuizAnswerParams{
			QuizSubmissionID:  submission.SubmissionID,
			QuestionID:        questionID,
			StudentAnswerText: toNullString(aReq.StudentAnswerText),
			SelectedOptionID:  selectedOptionID,
		})
		if err != nil {
			middleware.SendError(w, "Could not create answer", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		middleware.SendError(w, "Could not commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(submission)
}

func (h *QuizSubmissionHandler) GetQuizSubmissions(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	quizIDStr := r.URL.Query().Get("quizId")
	studentIDStr := r.URL.Query().Get("studentId")

	var quizID uuid.NullUUID
	if quizIDStr != "" {
		if id, err := uuid.Parse(quizIDStr); err == nil {
			quizID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	var studentID uuid.NullUUID
	if userCtx.RoleName == "Student" {
		studentID = uuid.NullUUID{UUID: userCtx.UserID, Valid: true}
	} else if studentIDStr != "" {
		if id, err := uuid.Parse(studentIDStr); err == nil {
			studentID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	submissions, err := h.Queries.GetQuizSubmissions(r.Context(), db.GetQuizSubmissionsParams{
		SchoolID:  schoolID,
		QuizID:    quizID,
		StudentID: studentID,
	})
	if err != nil {
		middleware.SendError(w, "Could not fetch submissions", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(submissions)
}

func (h *QuizSubmissionHandler) GetQuizSubmissionByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	submissionID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())

	submission, err := h.Queries.GetQuizSubmissionByID(r.Context(), submissionID)
	if err != nil {
		middleware.SendError(w, "Submission not found", http.StatusNotFound)
		return
	}

	// Basic authorization check
	if !middleware.IsAdmin(userCtx.RoleName) && submission.StudentID != userCtx.UserID && submission.TeacherID != userCtx.UserID {
		middleware.SendError(w, "Not authorized to view this submission", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(submission)
}

func (h *QuizSubmissionHandler) GradeQuizSubmission(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	submissionID, _ := uuid.Parse(idStr)

	var req struct {
		Score float64 `json:"score"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// Update score logic...
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"submission_id": submissionID, "score": req.Score, "status": "graded"})
}
