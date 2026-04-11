package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

type NewsletterHandler struct {
	Queries *db.Queries
}

func NewNewsletterHandler(q *db.Queries) *NewsletterHandler {
	return &NewsletterHandler{Queries: q}
}

func (h *NewsletterHandler) CreateNewsletter(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	userID := userCtx.UserID

	var req struct {
		Title         string   `json:"title"`
		Content       string   `json:"content"`
		TargetSchools []string `json:"target_schools"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	if req.Title == "" || req.Content == "" {
		middleware.ValidationError(w, "Title and content are required", nil)
		return
	}

	targetSchoolsJSON, _ := json.Marshal(req.TargetSchools)
	if len(req.TargetSchools) == 0 {
		targetSchoolsJSON = []byte("[]")
	}

	newsletter, err := h.Queries.CreateNewsletter(r.Context(), db.CreateNewsletterParams{
		Title:         req.Title,
		Content:       req.Content,
		SentByUserID:  uuid.NullUUID{UUID: userID, Valid: true},
		TargetSchools: pqtype.NullRawMessage{RawMessage: targetSchoolsJSON, Valid: true},
	})

	if err != nil {
		middleware.InternalError(w, "Could not create newsletter", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newsletter)
}

func (h *NewsletterHandler) GetNewsletters(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	// Logic from Node.js:
	// School admins can only see newsletters targeted to their school or all-school newsletters
	// !['Developer', 'DB Manager', 'Data Analyst', 'Support Staff'].includes(role_name)

	// The query GetNewsletters already handles this:
	// WHERE (target_schools @> $1::jsonb OR target_schools = '[]'::jsonb OR target_schools IS NULL)
	
	schoolIDJSON, _ := json.Marshal([]string{schoolID.String()})

	newsletters, err := h.Queries.GetNewsletters(r.Context(), schoolIDJSON)
	if err != nil {
		middleware.InternalError(w, "Could not fetch newsletters", err)
		return
	}

	json.NewEncoder(w).Encode(newsletters)
}

func (h *NewsletterHandler) GetNewsletterByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	newsletterID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid newsletter ID", err)
		return
	}

	newsletter, err := h.Queries.GetNewsletterByID(r.Context(), newsletterID)
	if err != nil {
		middleware.NotFoundError(w, "Newsletter not found", err)
		return
	}

	// Authorization check from Node.js:
	// if (!['Developer', 'DB Manager', 'Data Analyst', 'Support Staff'].includes(role_name) &&
	//     !(newsletter.target_schools.includes(school_id) || newsletter.target_schools.length === 0 || newsletter.target_schools === null))

	userCtx, _ := middleware.GetUser(r.Context())
	isSuperAdmin := false
	superAdminRoles := []string{"Developer", "DB Manager", "Data Analyst", "Support Staff"}
	for _, role := range superAdminRoles {
		if userCtx.RoleName == role {
			isSuperAdmin = true
			break
		}
	}

	if !isSuperAdmin {
		var targetSchools []string
		if newsletter.TargetSchools.Valid {
			json.Unmarshal(newsletter.TargetSchools.RawMessage, &targetSchools)
		}

		isTargeted := false
		if len(targetSchools) == 0 {
			isTargeted = true
		} else {
			for _, ts := range targetSchools {
				if ts == userCtx.SchoolID.UUID.String() {
					isTargeted = true
					break
				}
			}
		}

		if !isTargeted {
			middleware.ForbiddenError(w, "Not authorized to view this newsletter", err)
			return
		}
	}

	json.NewEncoder(w).Encode(newsletter)
}

func (h *NewsletterHandler) UpdateNewsletter(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	newsletterID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid newsletter ID", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	isSuperAdmin := false
	superAdminRoles := []string{"Developer", "DB Manager", "Data Analyst", "Support Staff"}
	for _, role := range superAdminRoles {
		if userCtx.RoleName == role {
			isSuperAdmin = true
			break
		}
	}

	if !isSuperAdmin {
		middleware.ForbiddenError(w, "Not authorized to update newsletters", err)
		return
	}

	var req struct {
		Title         string   `json:"title"`
		Content       string   `json:"content"`
		TargetSchools []string `json:"target_schools"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	existingNewsletter, err := h.Queries.GetNewsletterByID(r.Context(), newsletterID)
	if err != nil {
		middleware.NotFoundError(w, "Newsletter not found", err)
		return
	}

	params := db.UpdateNewsletterParams{
		NewsletterID:  newsletterID,
		Title:         existingNewsletter.Title,
		Content:       existingNewsletter.Content,
		TargetSchools: existingNewsletter.TargetSchools,
	}

	if req.Title != "" {
		params.Title = req.Title
	}
	if req.Content != "" {
		params.Content = req.Content
	}
	if req.TargetSchools != nil {
		targetSchoolsJSON, _ := json.Marshal(req.TargetSchools)
		params.TargetSchools = pqtype.NullRawMessage{RawMessage: targetSchoolsJSON, Valid: true}
	}

	updatedNewsletter, err := h.Queries.UpdateNewsletter(r.Context(), params)
	if err != nil {
		middleware.InternalError(w, "Could not update newsletter", err)
		return
	}

	json.NewEncoder(w).Encode(updatedNewsletter)
}

func (h *NewsletterHandler) DeleteNewsletter(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	newsletterID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid newsletter ID", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	isSuperAdmin := false
	superAdminRoles := []string{"Developer", "DB Manager", "Data Analyst", "Support Staff"}
	for _, role := range superAdminRoles {
		if userCtx.RoleName == role {
			isSuperAdmin = true
			break
		}
	}

	if !isSuperAdmin {
		middleware.ForbiddenError(w, "Not authorized to delete newsletters", err)
		return
	}

	err = h.Queries.DeleteNewsletter(r.Context(), newsletterID)
	if err != nil {
		middleware.InternalError(w, "Could not delete newsletter", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}



