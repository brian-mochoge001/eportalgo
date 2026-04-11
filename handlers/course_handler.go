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

type CourseHandler struct {
	Queries *db.Queries
}

func NewCourseHandler(q *db.Queries) *CourseHandler {
	return &CourseHandler{Queries: q}
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

	course, err := h.Queries.CreateCourse(r.Context(), db.CreateCourseParams{
		SchoolID:              schoolID,
		CourseCode:            req.CourseCode,
		CourseName:            req.CourseName,
		Description:           sql.NullString{String: req.Description, Valid: req.Description != ""},
		IsShortCourse:         req.IsShortCourse,
		Price:                 sql.NullString{String: req.Price, Valid: req.Price != ""},
		IsGradedIndependently: req.IsGradedIndependently,
	})

	if err != nil {
		middleware.InternalError(w, "Could not create course", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(course)
}

func (h *CourseHandler) EnrollShortCourse(w http.ResponseWriter, r *http.Request) {
	courseIDStr := chi.URLParam(r, "course_id")
	courseID, err := uuid.Parse(courseIDStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid course ID", err)
		return
	}

	var req struct {
		StudentID string `json:"student_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	studentID, err := uuid.Parse(req.StudentID)
	if err != nil {
		middleware.ValidationError(w, "Invalid student ID", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	// Verify course
	course, err := h.Queries.GetCourseByID(r.Context(), db.GetCourseByIDParams{
		CourseID: courseID,
		SchoolID: schoolID,
	})
	if err != nil || !course.IsShortCourse {
		middleware.NotFoundError(w, "Short course not found", err)
		return
	}

	// Auth check
	if userCtx.RoleName == "Student" && userCtx.UserID != studentID {
		middleware.ForbiddenError(w, "Students can only enroll themselves", err)
		return
	}

	// Check existing enrollment
	_, err = h.Queries.CheckShortCourseEnrollment(r.Context(), db.CheckShortCourseEnrollmentParams{
		StudentID: studentID,
		CourseID:  courseID,
	})
	if err == nil {
		middleware.SendError(w, "Student is already enrolled", http.StatusConflict, "FETCH_ERROR", err)
		return
	}

	enrollment, err := h.Queries.EnrollShortCourse(r.Context(), db.EnrollShortCourseParams{
		StudentID: studentID,
		CourseID:  courseID,
		SchoolID:  schoolID,
		Status:    "Enrolled",
	})

	if err != nil {
		middleware.InternalError(w, "Could not enroll student", err)
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

	err := h.Queries.UnenrollShortCourse(r.Context(), db.UnenrollShortCourseParams{
		StudentID: studentID,
		CourseID:  courseID,
		SchoolID:  schoolID,
	})

	if err != nil {
		middleware.InternalError(w, "Could not unenroll student", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Student unenrolled successfully"})
}



