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

type CourseHandler struct {
	Queries       *db.Queries
	CourseService *services.CourseService
}

func NewCourseHandler(q *db.Queries, s *services.CourseService) *CourseHandler {
	return &CourseHandler{Queries: q, CourseService: s}
}

func (h *CourseHandler) GetCourses(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	courses, err := h.Queries.GetCoursesBySchool(r.Context(), schoolID)
	if err != nil {
		middleware.InternalError(w, "Could not fetch courses", err)
		return
	}

	json.NewEncoder(w).Encode(courses)
}

func (h *CourseHandler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		CourseCode            string `json:"course_code"`
		CourseName            string `json:"course_name"`
		Description           string `json:"description"`
		IsShortCourse         bool   `json:"is_short_course"`
		Price                 string `json:"price"`
		IsGradedIndependently bool   `json:"is_graded_independently"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	course, err := h.CourseService.CreateCourse(r.Context(), services.CreateCourseParams{
		SchoolID:              schoolID,
		CourseCode:            req.CourseCode,
		CourseName:            req.CourseName,
		Description:           req.Description,
		IsShortCourse:         req.IsShortCourse,
		Price:                 req.Price,
		IsGradedIndependently: req.IsGradedIndependently,
	})

	if err != nil {
		middleware.InternalError(w, err.Error(), err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(course)
}

func (h *CourseHandler) EnrollShortCourse(w http.ResponseWriter, r *http.Request) {
	courseIDStr := chi.URLParam(r, "course_id")
	courseID, _ := uuid.Parse(courseIDStr)

	var req struct {
		StudentID string `json:"student_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	studentID, _ := uuid.Parse(req.StudentID)
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	// Auth check
	if userCtx.RoleName == "Student" && userCtx.UserID != studentID {
		middleware.ForbiddenError(w, "Students can only enroll themselves", nil)
		return
	}

	enrollment, err := h.CourseService.EnrollShortCourse(r.Context(), courseID, studentID, schoolID)
	if err != nil {
		middleware.InternalError(w, err.Error(), err)
		return
	}

	json.NewEncoder(w).Encode(enrollment)
}

func (h *CourseHandler) UnenrollShortCourse(w http.ResponseWriter, r *http.Request) {
	courseIDStr := chi.URLParam(r, "course_id")
	studentIDStr := chi.URLParam(r, "student_id")
	courseID, _ := uuid.Parse(courseIDStr)
	studentID, _ := uuid.Parse(studentIDStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	// Auth check
	if userCtx.RoleName == "Student" && userCtx.UserID != studentID {
		middleware.ForbiddenError(w, "Students can only unenroll themselves", nil)
		return
	}

	err := h.CourseService.UnenrollShortCourse(r.Context(), courseID, studentID, schoolID)
	if err != nil {
		middleware.InternalError(w, err.Error(), err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Student unenrolled successfully"})
}
