package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ShortCourseGradeHandler struct {
	Queries *db.Queries
}

func NewShortCourseGradeHandler(q *db.Queries) *ShortCourseGradeHandler {
	return &ShortCourseGradeHandler{Queries: q}
}

func (h *ShortCourseGradeHandler) GradeShortCourse(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		EnrollmentID string `json:"enrollment_id"`
		Score        string `json:"score"`
		Feedback     string `json:"feedback"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	enrollmentID, _ := uuid.Parse(req.EnrollmentID)

	// Fetch enrollment details to get course_id, student_id, and school_id for validation
	enrollment, err := h.Queries.GetEnrollmentByID(r.Context(), enrollmentID)
	if err != nil {
		middleware.SendError(w, "Enrollment not found", http.StatusNotFound)
		return
	}

	if enrollment.SchoolID != schoolID {
		middleware.SendError(w, "Enrollment does not belong to this school", http.StatusForbidden)
		return
	}

	grade, err := h.Queries.GradeShortCourse(r.Context(), db.GradeShortCourseParams{
		EnrollmentID: enrollmentID,
		CourseID:     enrollment.CourseID,
		StudentID:    enrollment.StudentID,
		Score:        req.Score,
		Feedback:     toNullString(req.Feedback),
		GradedByUserID: uuid.NullUUID{UUID: userCtx.UserID, Valid: true},
	})

	if err != nil {
		middleware.SendError(w, "Could not grade short course", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(grade)
}

func (h *ShortCourseGradeHandler) GetShortCourseGrades(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	courseIDStr := r.URL.Query().Get("courseId")
	studentIDStr := r.URL.Query().Get("studentId")

	var courseID uuid.NullUUID
	if courseIDStr != "" {
		if id, err := uuid.Parse(courseIDStr); err == nil {
			courseID = uuid.NullUUID{UUID: id, Valid: true}
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

	grades, err := h.Queries.GetShortCourseGrades(r.Context(), db.GetShortCourseGradesParams{
		SchoolID:  schoolID,
		CourseID:  courseID,
		StudentID: studentID,
	})
	if err != nil {
		middleware.SendError(w, "Could not fetch short course grades", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(grades)
}

func (h *ShortCourseGradeHandler) GetShortCourseGradeByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	gradeID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	grade, err := h.Queries.GetShortCourseGradeByID(r.Context(), gradeID)
	if err != nil {
		middleware.SendError(w, "Short course grade not found", http.StatusNotFound)
		return
	}

	if grade.SchoolID != schoolID {
		middleware.SendError(w, "Not authorized to view this grade", http.StatusForbidden)
		return
	}

	if userCtx.RoleName == "Student" && grade.StudentID != userCtx.UserID {
		middleware.SendError(w, "Not authorized to view this grade", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(grade)
}
