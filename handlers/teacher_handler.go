package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TeacherHandler struct {
	Queries *db.Queries
}

func NewTeacherHandler(q *db.Queries) *TeacherHandler {
	return &TeacherHandler{Queries: q}
}

func (h *TeacherHandler) GetTeachers(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	teachers, err := h.Queries.GetTeachersBySchool(r.Context(), schoolID)
	if err != nil {
		middleware.InternalError(w, "Could not fetch teachers", err)
		return
	}

	json.NewEncoder(w).Encode(teachers)
	}

	func (h *TeacherHandler) GetTeacherByID(w http.ResponseWriter, r *http.Request) {
	teacherIDStr := chi.URLParam(r, "id")
	teacherID, _ := uuid.Parse(teacherIDStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	teacher, err := h.Queries.GetTeacherByUserID(r.Context(), db.GetTeacherByUserIDParams{
		UserID:   teacherID,
		SchoolID: schoolID,
	})
	if err != nil {
		middleware.NotFoundError(w, "Teacher not found", err)
		return
	}

	json.NewEncoder(w).Encode(teacher)
}


