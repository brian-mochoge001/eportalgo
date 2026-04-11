package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AttendanceHandler struct {
	Queries *db.Queries
}

func NewAttendanceHandler(q *db.Queries) *AttendanceHandler {
	return &AttendanceHandler{Queries: q}
}

func (h *AttendanceHandler) GetAttendanceByClass(w http.ResponseWriter, r *http.Request) {
	classIDStr := chi.URLParam(r, "class_id")
	classID, _ := uuid.Parse(classIDStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	attendance, err := h.Queries.GetAttendanceByClass(r.Context(), db.GetAttendanceByClassParams{
		ClassID:  classID,
		SchoolID: schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Could not fetch attendance", http.StatusInternalServerError)
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
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID
	classID, _ := uuid.Parse(req.ClassID)
	attDate, _ := time.Parse("2006-01-02", req.AttendanceDate)

	// Verify teacher
	academicClass, err := h.Queries.GetClassByID(r.Context(), db.GetClassByIDParams{
		ClassID:  classID,
		SchoolID: schoolID,
	})
	if err != nil || academicClass.TeacherID != userCtx.UserID {
		middleware.SendError(w, "Not authorized to mark attendance for this class", http.StatusForbidden)
		return
	}

	var results []db.AttendanceRecord
	for _, s := range req.StudentsAttendance {
		sid, _ := uuid.Parse(s.StudentID)
		
		// Check existing
		existing, err := h.Queries.GetAttendanceRecordByUnique(r.Context(), db.GetAttendanceRecordByUniqueParams{
			SchoolID:       schoolID,
			StudentID:      sid,
			ClassID:        classID,
			AttendanceDate: attDate,
		})

		var record db.AttendanceRecord
		if err == nil {
			record, _ = h.Queries.UpdateAttendanceRecord(r.Context(), db.UpdateAttendanceRecordParams{
				AttendanceID: existing.AttendanceID,
				Status:       s.Status,
				Notes:        sql.NullString{String: s.Notes, Valid: s.Notes != ""},
				SchoolID:     schoolID,
			})
		} else {
			record, _ = h.Queries.CreateAttendanceRecord(r.Context(), db.CreateAttendanceRecordParams{
				SchoolID:       schoolID,
				StudentID:      sid,
				ClassID:        classID,
				AttendanceDate: attDate,
				Status:         s.Status,
				Notes:          sql.NullString{String: s.Notes, Valid: s.Notes != ""},
			})
		}
		results = append(results, record)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Attendance marked successfully.",
		"data":    map[string]interface{}{"attendanceRecords": results},
	})
}

func (h *AttendanceHandler) GetStudentAttendance(w http.ResponseWriter, r *http.Request) {
	studentIDStr := chi.URLParam(r, "student_id")
	studentID, _ := uuid.Parse(studentIDStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	if userCtx.RoleName == "Student" && userCtx.UserID != studentID {
		middleware.SendError(w, "Forbidden", http.StatusForbidden)
		return
	}

	attendance, err := h.Queries.GetStudentAttendance(r.Context(), db.GetStudentAttendanceParams{
		StudentID: studentID,
		SchoolID:  schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Could not fetch attendance", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(attendance)
}
