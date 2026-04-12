package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/brian-mochoge001/eportalgo/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AttendanceHandler struct {
	Queries           *db.Queries
	AttendanceService *services.AttendanceService
}

func NewAttendanceHandler(q *db.Queries, s *services.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{Queries: q, AttendanceService: s}
}

func (h *AttendanceHandler) GetAttendanceByClass(w http.ResponseWriter, r *http.Request) {
	q := GetQueries(r.Context(), h.Queries)
	classIDStr := chi.URLParam(r, "class_id")
	classID, _ := uuid.Parse(classIDStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	attendance, err := q.GetAttendanceByClass(r.Context(), db.GetAttendanceByClassParams{
		ClassID:  classID,
		SchoolID: schoolID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not fetch attendance", err)
		return
	}

	json.NewEncoder(w).Encode(attendance)
}

func (h *AttendanceHandler) MarkAttendance(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ClassID            string `json:"class_id"`
		AttendanceDate     string `json:"attendance_date"`
		StudentsAttendance []struct {
			StudentID string `json:"student_id"`
			Status    string `json:"status"`
			Notes     string `json:"notes"`
		} `json:"students_attendance"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID
	classID, _ := uuid.Parse(req.ClassID)
	attDate, _ := time.Parse("2006-01-02", req.AttendanceDate)

	var studentAtts []services.StudentAttendance
	for _, s := range req.StudentsAttendance {
		sid, _ := uuid.Parse(s.StudentID)
		studentAtts = append(studentAtts, services.StudentAttendance{
			StudentID: sid,
			Status:    s.Status,
			Notes:     s.Notes,
		})
	}

	results, err := h.AttendanceService.MarkAttendance(r.Context(), services.MarkAttendanceParams{
		SchoolID:           schoolID,
		ClassID:            classID,
		TeacherID:          userCtx.UserID,
		AttendanceDate:     attDate,
		StudentsAttendance: studentAtts,
	})

	if err != nil {
		if err.Error() == "not authorized to mark attendance for this class" {
			middleware.ForbiddenError(w, err.Error(), err)
			return
		}
		middleware.InternalError(w, "Could not mark attendance", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Attendance marked successfully.",
		"data":    map[string]interface{}{"attendanceRecords": results},
	})
}

func (h *AttendanceHandler) GetStudentAttendance(w http.ResponseWriter, r *http.Request) {
	q := GetQueries(r.Context(), h.Queries)
	studentIDStr := chi.URLParam(r, "student_id")
	studentID, _ := uuid.Parse(studentIDStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	if userCtx.RoleName == "Student" && userCtx.UserID != studentID {
		middleware.ForbiddenError(w, "Forbidden", nil)
		return
	}

	attendance, err := q.GetStudentAttendance(r.Context(), db.GetStudentAttendanceParams{
		StudentID: studentID,
		SchoolID:  schoolID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not fetch attendance", err)
		return
	}

	json.NewEncoder(w).Encode(attendance)
}


