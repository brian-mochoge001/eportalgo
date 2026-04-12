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

type BadgeHandler struct {
	Queries *db.Queries
}

func NewBadgeHandler(q *db.Queries) *BadgeHandler {
	return &BadgeHandler{Queries: q}
}

func (h *BadgeHandler) GetBadges(w http.ResponseWriter, r *http.Request) {
	q := GetQueries(r.Context(), h.Queries)
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	badges, err := q.GetBadgesBySchool(r.Context(), schoolID)
	if err != nil {
		middleware.InternalError(w, "Could not fetch badges", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"results": len(badges),
		"data": map[string]interface{}{"badges": badges},
	})
}

func (h *BadgeHandler) GetBadgeByID(w http.ResponseWriter, r *http.Request) {
	q := GetQueries(r.Context(), h.Queries)
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	badge, err := q.GetBadgeByID(r.Context(), db.GetBadgeByIDParams{
		BadgeID: id,
		SchoolID: schoolID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.NotFoundError(w, "Badge not found", err)
			return
		}
		middleware.InternalError(w, "Internal Server Error", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{"badge": badge},
	})
}

func (h *BadgeHandler) CreateBadge(w http.ResponseWriter, r *http.Request) {
	q := GetQueries(r.Context(), h.Queries)
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		BadgeName   string `json:"badge_name"`
		Description string `json:"description"`
		IconURL     string `json:"icon_url"`
		Criteria    string `json:"criteria"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	badge, err := q.CreateBadge(r.Context(), db.CreateBadgeParams{
		SchoolID:    schoolID,
		BadgeName:   req.BadgeName,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		IconUrl:     sql.NullString{String: req.IconURL, Valid: req.IconURL != ""},
		Criteria:    req.Criteria,
	})

	if err != nil {
		middleware.InternalError(w, "Could not create badge", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{"badge": badge},
	})
}

func (h *BadgeHandler) UpdateBadge(w http.ResponseWriter, r *http.Request) {
	q := GetQueries(r.Context(), h.Queries)
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		BadgeName   string `json:"badge_name"`
		Description string `json:"description"`
		IconURL     string `json:"icon_url"`
		Criteria    string `json:"criteria"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	updated, err := q.UpdateBadge(r.Context(), db.UpdateBadgeParams{
		BadgeID:     id,
		BadgeName:   req.BadgeName,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		IconUrl:     sql.NullString{String: req.IconURL, Valid: req.IconURL != ""},
		Criteria:    req.Criteria,
		SchoolID:    schoolID,
	})

	if err != nil {
		middleware.InternalError(w, "Could not update badge", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{"badge": updated},
	})
}

func (h *BadgeHandler) DeleteBadge(w http.ResponseWriter, r *http.Request) {
	q := GetQueries(r.Context(), h.Queries)
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	err := q.DeleteBadge(r.Context(), db.DeleteBadgeParams{
		BadgeID: id,
		SchoolID: schoolID,
	})

	if err != nil {
		middleware.InternalError(w, "Could not delete badge", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BadgeHandler) AwardBadge(w http.ResponseWriter, r *http.Request) {
	q := GetQueries(r.Context(), h.Queries)
	badgeIDStr := chi.URLParam(r, "badgeId")
	badgeID, _ := uuid.Parse(badgeIDStr)

	var req struct {
		StudentID string `json:"student_id"`
		Notes     string `json:"notes"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID
	studentID, _ := uuid.Parse(req.StudentID)

	// Verify student
	student, err := q.GetUser(r.Context(), db.GetUserParams{
		UserID:   studentID,
		SchoolID: uuid.NullUUID{UUID: schoolID, Valid: true},
	})
	if err != nil || student.UserID == uuid.Nil { // Check if user exists and belongs to the school
		middleware.NotFoundError(w, "Student not found in your school", err)
		return
	}

	award, err := q.AwardBadge(r.Context(), db.AwardBadgeParams{
		SchoolID:         schoolID,
		StudentID:        studentID,
		BadgeID:          badgeID,
		AwardedByUserID:  uuid.NullUUID{UUID: userCtx.UserID, Valid: true},
		Notes:            sql.NullString{String: req.Notes, Valid: req.Notes != ""},
	})

	if err != nil {
		middleware.InternalError(w, "Could not award badge", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"message": "Badge awarded successfully",
		"data": map[string]interface{}{"studentBadge": award},
	})
}

func (h *BadgeHandler) RevokeBadge(w http.ResponseWriter, r *http.Request) {
	q := GetQueries(r.Context(), h.Queries)
	badgeIDStr := chi.URLParam(r, "badgeId")
	studentIDStr := chi.URLParam(r, "studentId")
	badgeID, _ := uuid.Parse(badgeIDStr)
	studentID, _ := uuid.Parse(studentIDStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	err := q.RevokeBadge(r.Context(), db.RevokeBadgeParams{
		BadgeID:   badgeID,
		StudentID: studentID,
		SchoolID:  schoolID,
	})

	if err != nil {
		middleware.InternalError(w, "Could not revoke badge", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Badge revoked successfully"})
}



