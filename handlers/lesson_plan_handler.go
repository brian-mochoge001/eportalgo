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

type LessonPlanHandler struct {
	Queries *db.Queries
}

func NewLessonPlanHandler(q *db.Queries) *LessonPlanHandler {
	return &LessonPlanHandler{Queries: q}
}

func (h *LessonPlanHandler) CreateLessonPlan(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID
	userID := userCtx.UserID

	var req struct {
		Title       string `json:"title"`
		Content     string `json:"content"`
		ClassID     string `json:"class_id"`
		DateCovered string `json:"date_covered"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.Content == "" {
		middleware.SendError(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	var classID uuid.NullUUID
	if req.ClassID != "" {
		if id, err := uuid.Parse(req.ClassID); err == nil {
			classID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	var dateCovered sql.NullTime
	if req.DateCovered != "" {
		if t, err := time.Parse(time.RFC3339, req.DateCovered); err == nil {
			dateCovered = sql.NullTime{Time: t, Valid: true}
		}
	}

	lessonPlan, err := h.Queries.CreateLessonPlan(r.Context(), db.CreateLessonPlanParams{
		SchoolID:    schoolID,
		TeacherID:   userID,
		ClassID:     classID,
		Title:       req.Title,
		Content:     sql.NullString{String: req.Content, Valid: req.Content != ""},
		DateCovered: dateCovered,
	})

	if err != nil {
		middleware.SendError(w, "Could not create lesson plan", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(lessonPlan)
}

func (h *LessonPlanHandler) GetLessonPlans(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	classIDStr := r.URL.Query().Get("classId")
	teacherIDStr := r.URL.Query().Get("teacherId")

	var classID uuid.NullUUID
	if classIDStr != "" {
		if id, err := uuid.Parse(classIDStr); err == nil {
			classID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	var teacherID uuid.NullUUID
	if teacherIDStr != "" {
		if id, err := uuid.Parse(teacherIDStr); err == nil {
			teacherID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	// Teachers can only see their own lesson plans unless they are an admin
	if userCtx.RoleName == "Teacher" {
		teacherID = uuid.NullUUID{UUID: userCtx.UserID, Valid: true}
	}

	lessonPlans, err := h.Queries.GetLessonPlans(r.Context(), db.GetLessonPlansParams{
		SchoolID:  schoolID,
		TeacherID: teacherID,
		ClassID:   classID,
	})
	if err != nil {
		middleware.SendError(w, "Could not fetch lesson plans", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(lessonPlans)
}

func (h *LessonPlanHandler) GetLessonPlanByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	lessonPlanID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.SendError(w, "Invalid lesson plan ID", http.StatusBadRequest)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	lessonPlan, err := h.Queries.GetLessonPlanByID(r.Context(), db.GetLessonPlanByIDParams{
		LessonPlanID: lessonPlanID,
		SchoolID:     schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Lesson plan not found", http.StatusNotFound)
		return
	}

	// Ensure teacher can only access their own lesson plans unless they are an admin
	if userCtx.RoleName == "Teacher" && lessonPlan.TeacherID != userCtx.UserID {
		middleware.SendError(w, "Not authorized to view this lesson plan", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(lessonPlan)
}

func (h *LessonPlanHandler) UpdateLessonPlan(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	lessonPlanID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.SendError(w, "Invalid lesson plan ID", http.StatusBadRequest)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		Title       string `json:"title"`
		Content     string `json:"content"`
		ClassID     string `json:"class_id"`
		DateCovered string `json:"date_covered"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	existingLessonPlan, err := h.Queries.GetLessonPlanByID(r.Context(), db.GetLessonPlanByIDParams{
		LessonPlanID: lessonPlanID,
		SchoolID:     schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Lesson plan not found", http.StatusNotFound)
		return
	}

	// Ensure teacher can only update their own lesson plans unless they are an admin
	if userCtx.RoleName == "Teacher" && existingLessonPlan.TeacherID != userCtx.UserID {
		middleware.SendError(w, "Not authorized to update this lesson plan", http.StatusForbidden)
		return
	}

	params := db.UpdateLessonPlanParams{
		LessonPlanID: lessonPlanID,
		SchoolID:     schoolID,
		TeacherID:    existingLessonPlan.TeacherID,
		Title:        existingLessonPlan.Title,
		Content:      existingLessonPlan.Content,
		ClassID:      existingLessonPlan.ClassID,
		DateCovered:  existingLessonPlan.DateCovered,
	}

	if req.Title != "" {
		params.Title = req.Title
	}
	if req.Content != "" {
		params.Content = sql.NullString{String: req.Content, Valid: true}
	}
	if req.ClassID != "" {
		if id, err := uuid.Parse(req.ClassID); err == nil {
			params.ClassID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}
	if req.DateCovered != "" {
		if t, err := time.Parse(time.RFC3339, req.DateCovered); err == nil {
			params.DateCovered = sql.NullTime{Time: t, Valid: true}
		}
	}

	updatedLessonPlan, err := h.Queries.UpdateLessonPlan(r.Context(), params)
	if err != nil {
		middleware.SendError(w, "Could not update lesson plan", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updatedLessonPlan)
}

func (h *LessonPlanHandler) DeleteLessonPlan(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	lessonPlanID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.SendError(w, "Invalid lesson plan ID", http.StatusBadRequest)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	existingLessonPlan, err := h.Queries.GetLessonPlanByID(r.Context(), db.GetLessonPlanByIDParams{
		LessonPlanID: lessonPlanID,
		SchoolID:     schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Lesson plan not found", http.StatusNotFound)
		return
	}

	// Ensure teacher can only delete their own lesson plans unless they are an admin
	if userCtx.RoleName == "Teacher" && existingLessonPlan.TeacherID != userCtx.UserID {
		middleware.SendError(w, "Not authorized to delete this lesson plan", http.StatusForbidden)
		return
	}

	err = h.Queries.DeleteLessonPlan(r.Context(), db.DeleteLessonPlanParams{
		LessonPlanID: lessonPlanID,
		SchoolID:     schoolID,
		TeacherID:    existingLessonPlan.TeacherID,
	})
	if err != nil {
		middleware.SendError(w, "Could not delete lesson plan", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
