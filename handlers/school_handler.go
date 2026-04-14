package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/brian-mochoge001/eportalgo/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sqlc-dev/pqtype"
)

type SchoolHandler struct {
	Queries       *db.Queries
	SchoolService *services.SchoolService
	Redis         *redis.Client
}

func NewSchoolHandler(q *db.Queries, s *services.SchoolService, r *redis.Client) *SchoolHandler {
	return &SchoolHandler{Queries: q, SchoolService: s, Redis: r}
}

func generateSchoolInitial(name string) string {
	words := strings.Fields(name)
	var initial string
	if len(words) > 1 {
		for _, word := range words {
			if len(word) > 0 {
				initial += strings.ToUpper(string(word[0]))
			}
		}
	} else if len(name) >= 3 {
		initial = strings.ToUpper(name[:3])
	} else {
		initial = strings.ToUpper(name)
	}

	b := make([]byte, 2)
	rand.Read(b)
	return initial + strings.ToUpper(hex.EncodeToString(b))
}

func (h *SchoolHandler) RegisterSchool(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SchoolName     string `json:"schoolName"`
		Subdomain      string `json:"subdomain"`
		Address        string `json:"address"`
		PhoneNumber    string `json:"phoneNumber"`
		Email          string `json:"email"`
		AdminFirstName string `json:"adminFirstName"`
		AdminLastName  string `json:"adminLastName"`
		AdminEmail     string `json:"adminEmail"`
		AdminRoleName  string `json:"adminRoleName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	resp, err := h.SchoolService.RegisterSchool(r.Context(), services.RegisterSchoolRequest{
		SchoolName:     req.SchoolName,
		Subdomain:      req.Subdomain,
		Address:        req.Address,
		PhoneNumber:    req.PhoneNumber,
		Email:          req.Email,
		AdminFirstName: req.AdminFirstName,
		AdminLastName:  req.AdminLastName,
		AdminEmail:     req.AdminEmail,
		AdminRoleName:  req.AdminRoleName,
	})

	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			middleware.SendError(w, err.Error(), http.StatusConflict, "CONFLICT", nil)
			return
		}
		slog.Error("Failed to register school", "error", err)
		middleware.InternalError(w, "Internal Server Error", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "School registration submitted for verification.",
		"school": map[string]interface{}{
			"id":      resp.School.SchoolID,
			"name":    resp.School.SchoolName,
			"status":  resp.School.Status,
			"initial": resp.School.SchoolInitial.String,
		},
		"adminUser": map[string]interface{}{
			"id":    resp.AdminUser.UserID,
			"email": resp.AdminUser.Email,
			"role":  resp.RoleName,
		},
	})
}

func (h *SchoolHandler) VerifySchool(w http.ResponseWriter, r *http.Request) {
	schoolIDStr := chi.URLParam(r, "schoolId")
	schoolID, err := uuid.Parse(schoolIDStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid school ID", err)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	if req.Status != "verified" && req.Status != "rejected" {
		middleware.ValidationError(w, "Invalid status. Must be 'verified' or 'rejected'", err)
		return
	}

	updatedSchool, err := h.SchoolService.VerifySchool(r.Context(), schoolID, req.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.NotFoundError(w, "School not found", err)
			return
		}
		middleware.InternalError(w, "Internal Server Error", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": fmt.Sprintf("School status updated to %s.", req.Status),
		"school":  updatedSchool,
	})
}

func (h *SchoolHandler) GetSchoolSettings(w http.ResponseWriter, r *http.Request) {
	schoolIDStr := chi.URLParam(r, "schoolId")
	schoolID, err := uuid.Parse(schoolIDStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid school ID", err)
		return
	}

	user, _ := middleware.GetUser(r.Context())
	if user.SchoolID.UUID != schoolID && !isParentCompanyAdmin(user.RoleName) {
		middleware.ForbiddenError(w, "Forbidden", err)
		return
	}

	cacheKey := fmt.Sprintf("schoolSettings:%s", schoolID)
	cached, err := h.Redis.Get(r.Context(), cacheKey).Result()
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}

	settings, err := h.Queries.GetSchoolSettings(r.Context(), schoolID)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.NotFoundError(w, "Settings not found", err)
			return
		}
		middleware.InternalError(w, "Internal Server Error", err)
		return
	}

	settingsJSON, _ := json.Marshal(settings)
	h.Redis.Set(r.Context(), cacheKey, settingsJSON, time.Hour)

	w.Header().Set("Content-Type", "application/json")
	w.Write(settingsJSON)
}

func (h *SchoolHandler) UpdateSchoolSettings(w http.ResponseWriter, r *http.Request) {
	schoolIDStr := chi.URLParam(r, "schoolId")
	schoolID, err := uuid.Parse(schoolIDStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid school ID", err)
		return
	}

	user, _ := middleware.GetUser(r.Context())
	if user.SchoolID.UUID != schoolID || !isExecutiveAdmin(user.RoleName) {
		middleware.ForbiddenError(w, "Forbidden", err)
		return
	}

	var req struct {
		BrandingLogoUrl      string          `json:"branding_logo_url"`
		BrandingColors       json.RawMessage `json:"branding_colors"`
		Timezone             string          `json:"timezone"`
		Preferences          json.RawMessage `json:"preferences"`
		EmailTemplateConfigs json.RawMessage `json:"email_template_configs"`
		PaymentProviders     json.RawMessage `json:"payment_providers"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	updated, err := h.Queries.UpdateSchoolSettings(r.Context(), db.UpdateSchoolSettingsParams{
		SchoolID:             schoolID,
		BrandingLogoUrl:      sql.NullString{String: req.BrandingLogoUrl, Valid: true},
		BrandingColors:       req.BrandingColors,
		Timezone:             req.Timezone,
		Preferences:          req.Preferences,
		EmailTemplateConfigs: req.EmailTemplateConfigs,
		PaymentProviders:     pqtype.NullRawMessage{RawMessage: req.PaymentProviders, Valid: len(req.PaymentProviders) > 0},
	})

	if err != nil {
		slog.Error("Failed to update settings", "error", err)
		middleware.InternalError(w, "Internal Server Error", err)
		return
	}

	h.Redis.Del(r.Context(), fmt.Sprintf("schoolSettings:%s", schoolID))

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "School settings updated successfully.",
		"settings": updated,
	})
}

func isParentCompanyAdmin(role string) bool {
	admins := map[string]bool{
		"Developer":    true,
		"DB Manager":   true,
		"Data Analyst": true,
		"Support Staff": true,
	}
	return admins[role]
}

func isExecutiveAdmin(role string) bool {
	return role == "Executive Administrator" || role == "School Bursar"
}



