package handlers

import (
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

func (h *BadgeHandler) CreateBadge(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BadgeName   string `json:"badge_name"`
		Description string `json:"description"`
		IconUrl     string `json:"icon_url"`
		Criteria    string `json:"criteria"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	badge, err := h.Queries.CreateBadge(r.Context(), db.CreateBadgeParams{
		SchoolID:    user.SchoolID.UUID,
		BadgeName:   req.BadgeName,
		Description: toNullString(req.Description),
		IconUrl:     toNullString(req.IconUrl),
		Criteria:    req.Criteria,
	})
	if err != nil {
		http.Error(w, `{"error":"Failed to create badge"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(badge)
}

func (h *BadgeHandler) ListBadgesBySchool(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	badges, err := h.Queries.GetBadgesBySchool(r.Context(), user.SchoolID.UUID)
	if err != nil {
		http.Error(w, `{"error":"Failed to list badges"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(badges)
}

func (h *BadgeHandler) GetBadgeByID(w http.ResponseWriter, r *http.Request) {
	badgeIDStr := chi.URLParam(r, "badgeID")
	badgeID, err := uuid.Parse(badgeIDStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid badge ID"}`, http.StatusBadRequest)
		return
	}

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	badge, err := h.Queries.GetBadgeByID(r.Context(), db.GetBadgeByIDParams{
		BadgeID:  badgeID,
		SchoolID: user.SchoolID.UUID,
	})
	if err != nil {
		http.Error(w, `{"error":"Badge not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(badge)
}

func (h *BadgeHandler) UpdateBadge(w http.ResponseWriter, r *http.Request) {
	badgeIDStr := chi.URLParam(r, "badgeID")
	badgeID, err := uuid.Parse(badgeIDStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid badge ID"}`, http.StatusBadRequest)
		return
	}

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		BadgeName   string `json:"badge_name"`
		Description string `json:"description"`
		IconUrl     string `json:"icon_url"`
		Criteria    string `json:"criteria"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	badge, err := h.Queries.UpdateBadge(r.Context(), db.UpdateBadgeParams{
		BadgeID:     badgeID,
		SchoolID:    user.SchoolID.UUID,
		BadgeName:   req.BadgeName,
		Description: toNullString(req.Description),
		IconUrl:     toNullString(req.IconUrl),
		Criteria:    req.Criteria,
	})
	if err != nil {
		http.Error(w, `{"error":"Failed to update badge"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(badge)
}

func (h *BadgeHandler) DeleteBadge(w http.ResponseWriter, r *http.Request) {
	badgeIDStr := chi.URLParam(r, "badgeID")
	badgeID, err := uuid.Parse(badgeIDStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid badge ID"}`, http.StatusBadRequest)
		return
	}

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	err = h.Queries.DeleteBadge(r.Context(), db.DeleteBadgeParams{
		BadgeID:  badgeID,
		SchoolID: user.SchoolID.UUID,
	})
	if err != nil {
		http.Error(w, `{"error":"Failed to delete badge"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Badge deleted successfully"})
}


func (h *BadgeHandler) AwardBadge(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StudentID string `json:"student_id"`
		BadgeID   string `json:"badge_id"`
		Notes     string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	studentID, err := uuid.Parse(req.StudentID)
	if err != nil {
		http.Error(w, `{"error":"Invalid student ID"}`, http.StatusBadRequest)
		return
	}
	badgeID, err := uuid.Parse(req.BadgeID)
	if err != nil {
		http.Error(w, `{"error":"Invalid badge ID"}`, http.StatusBadRequest)
		return
	}

	sb, err := h.Queries.AwardBadge(r.Context(), db.AwardBadgeParams{
		SchoolID:        user.SchoolID.UUID,
		StudentID:       studentID,
		BadgeID:         badgeID,
		AwardedByUserID: uuid.NullUUID{UUID: user.UserID, Valid: true},
		Notes:           toNullString(req.Notes),
	})
	if err != nil {
		http.Error(w, `{"error":"Failed to award badge"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sb)
}

func (h *BadgeHandler) RevokeBadge(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BadgeID   string `json:"badge_id"`
		StudentID string `json:"student_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	badgeID, _ := uuid.Parse(req.BadgeID)
	studentID, _ := uuid.Parse(req.StudentID)

	err := h.Queries.RevokeBadge(r.Context(), db.RevokeBadgeParams{
		BadgeID:   badgeID,
		StudentID: studentID,
		SchoolID:  user.SchoolID.UUID,
	})
	if err != nil {
		http.Error(w, `{"error":"Failed to revoke badge"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Badge revoked successfully"})
}

func (h *BadgeHandler) GetStudentBadges(w http.ResponseWriter, r *http.Request) {
	studentIDStr := chi.URLParam(r, "studentID")
	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		http.Error(w, `{"error":"Invalid student ID"}`, http.StatusBadRequest)
		return
	}

	user, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	badges, err := h.Queries.GetStudentBadges(r.Context(), db.GetStudentBadgesParams{
		StudentID: studentID,
		SchoolID:  user.SchoolID.UUID,
	})
	if err != nil {
		http.Error(w, `{"error":"Failed to get student badges"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(badges)
}
