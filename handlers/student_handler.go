package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type StudentHandler struct {
	Queries *db.Queries
}

func NewStudentHandler(q *db.Queries) *StudentHandler {
	return &StudentHandler{Queries: q}
}

func (h *StudentHandler) GetStudents(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	students, err := h.Queries.GetStudentsBySchool(r.Context(), schoolID)
	if err != nil {
		middleware.SendError(w, "Could not fetch students", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(students)
}

func (h *StudentHandler) GetStudentByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	studentID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	student, err := h.Queries.GetStudentByUserID(r.Context(), db.GetStudentByUserIDParams{
		UserID:   studentID,
		SchoolID: schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Student not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(student)
}
