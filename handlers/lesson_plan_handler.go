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

type LessonPlanHandler struct {
	Queries           *db.Queries
	LessonPlanService *services.LessonPlanService
}

func NewLessonPlanHandler(q *db.Queries, s *services.LessonPlanService) *LessonPlanHandler {
	return &LessonPlanHandler{Queries: q, LessonPlanService: s}
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
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	classID := toNullUUID(req.ClassID)
	var dateCovered *time.Time
	if req.DateCovered != "" {
		if t, err := time.Parse(time.RFC3339, req.DateCovered); err == nil {
			dateCovered = &t
		}
	}

	lessonPlan, err := h.LessonPlanService.CreateLessonPlan(r.Context(), services.CreateLessonPlanParams{
		SchoolID:    schoolID,
		TeacherID:   userID,
		ClassID:     classID,
		Title:       req.Title,
		Content:     req.Content,
		DateCovered: dateCovered,
	})

	if err != nil {
		middleware.InternalError(w, err.Error(), err)
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
	if id, err := uuid.Parse(classIDStr); err == nil {
		classID = uuid.NullUUID{UUID: id, Valid: true}
	}

	var teacherID uuid.NullUUID
	if id, err := uuid.Parse(teacherIDStr); err == nil {
		teacherID = uuid.NullUUID{UUID: id, Valid: true}
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
		middleware.InternalError(w, "Could not fetch lesson plans", err)
		return
	}

	json.NewEncoder(w).Encode(lessonPlans)
}

func (h *LessonPlanHandler) GetLessonPlanByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	lessonPlanID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	lessonPlan, err := h.Queries.GetLessonPlanByID(r.Context(), db.GetLessonPlanByIDParams{
		LessonPlanID: lessonPlanID,
		SchoolID:     schoolID,
	})
	if err != nil {
		middleware.NotFoundError(w, "Lesson plan not found", err)
		return
	}

	// Ensure teacher can only access their own lesson plans unless they are an admin
	if userCtx.RoleName == "Teacher" && lessonPlan.TeacherID != userCtx.UserID {
		middleware.ForbiddenError(w, "Not authorized to view this lesson plan", nil)
		return
	}

	json.NewEncoder(w).Encode(lessonPlan)
}

func (h *LessonPlanHandler) UpdateLessonPlan(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	lessonPlanID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		Title       string `json:"title"`
		Content     string `json:"content"`
		ClassID     string `json:"class_id"`
		DateCovered string `json:"date_covered"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	params := services.UpdateLessonPlanParams{
		LessonPlanID: lessonPlanID,
		SchoolID:     schoolID,
		TeacherID:    userCtx.UserID,
		RoleName:     userCtx.RoleName,
		Title:        req.Title,
		Content:      req.Content,
	}

	if req.ClassID != "" {
		if id, err := uuid.Parse(req.ClassID); err == nil {
			params.ClassID = &id
		}
	}
	if req.DateCovered != "" {
		if t, err := time.Parse(time.RFC3339, req.DateCovered); err == nil {
			params.DateCovered = &t
		}
	}

	updated, err := h.LessonPlanService.UpdateLessonPlan(r.Context(), params)
	if err != nil {
		middleware.InternalError(w, err.Error(), err)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func (h *LessonPlanHandler) DeleteLessonPlan(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	lessonPlanID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())

	err := h.LessonPlanService.DeleteLessonPlan(r.Context(), lessonPlanID, userCtx.SchoolID.UUID, userCtx.UserID, userCtx.RoleName)
	if err != nil {
		middleware.InternalError(w, err.Error(), err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
