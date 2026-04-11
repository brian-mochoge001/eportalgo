package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ClassHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewClassHandler(q *db.Queries, d *sql.DB) *ClassHandler {
	return &ClassHandler{Queries: q, DB: d}
}

func (h *ClassHandler) GetClasses(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	classes, err := h.Queries.GetClassesBySchool(r.Context(), schoolID)
	if err != nil {
		middleware.InternalError(w, "Could not fetch classes", err)
		return
	}

	json.NewEncoder(w).Encode(classes)
}

func (h *ClassHandler) CreateClass(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		CourseID     string `json:"course_id"`
		TeacherID    string `json:"teacher_id"`
		ClassName    string `json:"class_name"`
		AcademicYear string `json:"academic_year"`
		Semester     string `json:"semester"`
		StartDate    string `json:"start_date"`
		EndDate      string `json:"end_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	courseID, _ := uuid.Parse(req.CourseID)
	teacherID, _ := uuid.Parse(req.TeacherID)
	startDate, _ := time.Parse("2006-01-02", req.StartDate)
	endDate, _ := time.Parse("2006-01-02", req.EndDate)

	newClass, err := h.Queries.CreateAcademicClass(r.Context(), db.CreateAcademicClassParams{
		SchoolID:     schoolID,
		CourseID:     courseID,
		TeacherID:    teacherID,
		ClassName:    req.ClassName,
		AcademicYear: req.AcademicYear,
		Semester:     sql.NullString{String: req.Semester, Valid: req.Semester != ""},
		StartDate:    sql.NullTime{Time: startDate, Valid: req.StartDate != ""},
		EndDate:      sql.NullTime{Time: endDate, Valid: req.EndDate != ""},
	})

	if err != nil {
		middleware.InternalError(w, "Could not create class", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newClass)
}

func (h *ClassHandler) AddStudentsToClass(w http.ResponseWriter, r *http.Request) {
	classIDStr := chi.URLParam(r, "class_id")
	classID, err := uuid.Parse(classIDStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid class ID", err)
		return
	}

	var req struct {
		StudentIds []string `json:"student_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	if !isAcademicAdmin(userCtx.RoleName) {
		middleware.ForbiddenError(w, "Forbidden", err)
		return
	}

	// Start transaction
	tx, err := h.DB.Begin()
	if err != nil {
		middleware.InternalError(w, "Internal Server Error", err)
		return
	}
	defer tx.Rollback()

	qtx := h.Queries.WithTx(tx)

	// Verify class
	_, err = qtx.GetClassByID(r.Context(), db.GetClassByIDParams{
		ClassID:  classID,
		SchoolID: schoolID,
	})
	if err != nil {
		middleware.NotFoundError(w, "Class not found", err)
		return
	}

	newEnrollmentsCount := 0
	alreadyEnrolledCount := 0

	for _, sidStr := range req.StudentIds {
		sid, err := uuid.Parse(sidStr)
		if err != nil {
			continue
		}

		// Check if already enrolled
		_, err = qtx.GetEnrollmentByStudentAndClass(r.Context(), db.GetEnrollmentByStudentAndClassParams{
			StudentID: sid,
			ClassID:   classID,
		})
		if err == nil {
			alreadyEnrolledCount++
			continue
		}

		// Verify student
		_, err = qtx.GetUser(r.Context(), db.GetUserParams{
			UserID:   sid,
			SchoolID: uuid.NullUUID{UUID: schoolID, Valid: true},
		})
		if err != nil {
			continue
		}

		// Create enrollment
		_, err = qtx.CreateEnrollment(r.Context(), db.CreateEnrollmentParams{
			SchoolID:       schoolID,
			StudentID:      sid,
			ClassID:        classID,
			EnrollmentDate: time.Now(),
			Status:         "Enrolled",
		})
		if err == nil {
			newEnrollmentsCount++
		}
	}

	tx.Commit()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":                fmt.Sprintf("Successfully enrolled %d new students.", newEnrollmentsCount),
		"newly_enrolled_count":   newEnrollmentsCount,
		"already_enrolled_count": alreadyEnrolledCount,
	})
}

func isAcademicAdmin(role string) bool {
	return role == "Teacher" || role == "Academic Administrator" || role == "Executive Administrator"
}



