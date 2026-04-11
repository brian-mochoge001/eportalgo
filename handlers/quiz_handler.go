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

type QuizHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewQuizHandler(q *db.Queries, d *sql.DB) *QuizHandler {
	return &QuizHandler{Queries: q, DB: d}
}

func (h *QuizHandler) GetQuizzes(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	quizzes, err := h.Queries.GetQuizzes(r.Context(), schoolID)
	if err != nil {
		middleware.InternalError(w, "Could not fetch quizzes", err)
		return
	}

	json.NewEncoder(w).Encode(quizzes)
}

func (h *QuizHandler) GetQuizByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	quizID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	quiz, err := h.Queries.GetQuizByID(r.Context(), db.GetQuizByIDParams{
		QuizID:   quizID,
		SchoolID: schoolID,
	})
	if err != nil {
		middleware.NotFoundError(w, "Quiz not found", err)
		return
	}

	json.NewEncoder(w).Encode(quiz)
}

func (h *QuizHandler) CreateQuiz(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		AssignmentID    string `json:"assignment_id"`
		SubjectID       string `json:"subject_id"`
		Title           string `json:"title"`
		Description     string `json:"description"`
		QuizType        string `json:"quiz_type"`
		DurationMinutes int32  `json:"duration_minutes"`
		StartTime       string `json:"start_time"`
		EndTime         string `json:"end_time"`
		Questions       []struct {
			QuestionText string `json:"question_text"`
			QuestionType string `json:"question_type"`
			Order        int32  `json:"order"`
			Options      []struct {
				OptionText string `json:"option_text"`
				IsCorrect  bool   `json:"is_correct"`
			} `json:"options"`
		} `json:"questions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	assignmentID := toNullUUID(req.AssignmentID)
	subjectID := toNullUUID(req.SubjectID)
	startTime, err := parseDate(req.StartTime)
	if err != nil {
		middleware.ValidationError(w, "Invalid start time format. Please use YYYY-MM-DD or RFC3339 format.", err)
		return
	}
	endTime, err := parseDate(req.EndTime)
	if err != nil {
		middleware.ValidationError(w, "Invalid end time format. Please use YYYY-MM-DD or RFC3339 format.", err)
		return
	}


	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		middleware.InternalError(w, "Could not start transaction", err)
		return
	}
	defer tx.Rollback()

	qtx := h.Queries.WithTx(tx)

	// Create Quiz
	quiz, err := qtx.CreateQuiz(r.Context(), db.CreateQuizParams{
		SchoolID:        schoolID,
		TeacherID:       userCtx.UserID,
		AssignmentID:    assignmentID,
		SubjectID:       subjectID,
		Title:           req.Title,
		Description:     toNullString(req.Description),
		QuizType:        req.QuizType,
		DurationMinutes: toNullInt32(&req.DurationMinutes),
		StartTime:       startTime,
		EndTime:         endTime,
	})
	if err != nil {
		middleware.InternalError(w, "Could not create quiz", err)
		return
	}

	// Create Questions and Options
	for _, qReq := range req.Questions {
		question, err := qtx.CreateQuestion(r.Context(), db.CreateQuestionParams{
			QuizID:       uuid.NullUUID{UUID: quiz.QuizID, Valid: true},
			QuestionText: qReq.QuestionText,
			QuestionType: qReq.QuestionType,
			Order:        qReq.Order,
		})
		if err != nil {
			middleware.InternalError(w, "Could not create question", err)
			return
		}

		if qReq.QuestionType == "multi_choice" {
			for _, oReq := range qReq.Options {
				_, err = qtx.CreateOption(r.Context(), db.CreateOptionParams{
					QuestionID: question.QuestionID,
					OptionText: oReq.OptionText,
					IsCorrect:  oReq.IsCorrect,
				})
				if err != nil {
					middleware.InternalError(w, "Could not create option", err)
					return
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		middleware.InternalError(w, "Could not commit transaction", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(quiz)
}

func (h *QuizHandler) UpdateQuiz(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	quizID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		Title           string `json:"title"`
		Description     string `json:"description"`
		QuizType        string `json:"quiz_type"`
		DurationMinutes int32  `json:"duration_minutes"`
		StartTime       string `json:"start_time"`
		EndTime         string `json:"end_time"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	startTime, err := parseDate(req.StartTime)
	if err != nil {
		middleware.ValidationError(w, "Invalid start time format. Please use YYYY-MM-DD or RFC3339 format.", err)
		return
	}
	endTime, err := parseDate(req.EndTime)
	if err != nil {
		middleware.ValidationError(w, "Invalid end time format. Please use YYYY-MM-DD or RFC3339 format.", err)
		return
	}

	updated, err := h.Queries.UpdateQuiz(r.Context(), db.UpdateQuizParams{
		QuizID:          quizID,
		Title:           req.Title,
		Description:     toNullString(req.Description),
		QuizType:        req.QuizType,
		DurationMinutes: toNullInt32(&req.DurationMinutes),
		StartTime:       startTime,
		EndTime:         endTime,
		SchoolID:        schoolID,
	})


	if err != nil {
		middleware.InternalError(w, "Could not update quiz", err)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func (h *QuizHandler) DeleteQuiz(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	quizID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	err := h.Queries.DeleteQuiz(r.Context(), db.DeleteQuizParams{
		QuizID:   quizID,
		SchoolID: schoolID,
	})

	if err != nil {
		middleware.InternalError(w, "Could not delete quiz", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}



