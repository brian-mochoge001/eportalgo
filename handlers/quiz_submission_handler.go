package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/brian-mochoge001/eportalgo/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type QuizSubmissionHandler struct {
	Queries     *db.Queries
	QuizService *services.QuizService
}

func NewQuizSubmissionHandler(q *db.Queries, s *services.QuizService) *QuizSubmissionHandler {
	return &QuizSubmissionHandler{Queries: q, QuizService: s}
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
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	quizID, _ := uuid.Parse(req.QuizID)

	var answers []services.QuizAnswerRequest
	for _, a := range req.Answers {
		qid, _ := uuid.Parse(a.QuestionID)
		answers = append(answers, services.QuizAnswerRequest{
			QuestionID:        qid,
			StudentAnswerText: a.StudentAnswerText,
			SelectedOptionID:  toNullUUID(a.SelectedOptionID),
		})
	}

	submission, err := h.QuizService.SubmitQuiz(r.Context(), services.QuizSubmissionRequest{
		QuizID:    quizID,
		StudentID: userCtx.UserID,
		Answers:   answers,
	})

	if err != nil {
		middleware.InternalError(w, "Could not submit quiz", err)
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
		middleware.InternalError(w, "Could not fetch submissions", err)
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
		middleware.NotFoundError(w, "Submission not found", err)
		return
	}

	// Basic authorization check
	if !middleware.IsAdmin(userCtx.RoleName) && submission.StudentID != userCtx.UserID && submission.TeacherID != userCtx.UserID {
		middleware.ForbiddenError(w, "Not authorized to view this submission", err)
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



