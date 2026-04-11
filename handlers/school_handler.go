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

	"firebase.google.com/go/v4/auth"
	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sqlc-dev/pqtype"
)

type SchoolHandler struct {
	Queries      *db.Queries
	FirebaseAuth *auth.Client
	Redis        *redis.Client
}

func NewSchoolHandler(q *db.Queries, fb *auth.Client, r *redis.Client) *SchoolHandler {
	return &SchoolHandler{Queries: q, FirebaseAuth: fb, Redis: r}
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
		AdminFirebaseUid string `json:"adminFirebaseUid"`
		AdminRoleName  string `json:"adminRoleName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AdminRoleName == "" {
		req.AdminRoleName = "Executive Administrator"
	}

	// Check if school exists
	existing, _ := h.Queries.GetSchoolByNameOrSubdomain(r.Context(), db.GetSchoolByNameOrSubdomainParams{
		SchoolName: req.SchoolName,
		Subdomain:  sql.NullString{String: req.Subdomain, Valid: req.Subdomain != ""},
	})
	if existing.SchoolID != uuid.Nil {
		middleware.SendError(w, "School with this name or subdomain already exists", http.StatusConflict)
		return
	}

	// Generate unique initial
	schoolInitial := generateSchoolInitial(req.SchoolName)
	for {
		exists, _ := h.Queries.GetSchoolByInitial(r.Context(), sql.NullString{String: schoolInitial, Valid: true})
		if exists.SchoolID == uuid.Nil {
			break
		}
		schoolInitial = generateSchoolInitial(req.SchoolName)
	}

	// Create School
	school, err := h.Queries.CreateSchool(r.Context(), db.CreateSchoolParams{
		SchoolName:    req.SchoolName,
		Subdomain:     sql.NullString{String: req.Subdomain, Valid: req.Subdomain != ""},
		Status:        "pending",
		SchoolInitial: sql.NullString{String: schoolInitial, Valid: true},
		Address:       sql.NullString{String: req.Address, Valid: req.Address != ""},
		PhoneNumber:   sql.NullString{String: req.PhoneNumber, Valid: req.PhoneNumber != ""},
		Email:         sql.NullString{String: req.Email, Valid: req.Email != ""},
	})
	if err != nil {
		slog.Error("Failed to create school", "error", err)
		middleware.SendError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Find Role
	role, err := h.Queries.GetRoleByName(r.Context(), req.AdminRoleName)
	if err != nil {
		middleware.SendError(w, fmt.Sprintf("Role '%s' not found", req.AdminRoleName), http.StatusInternalServerError)
		return
	}

	// Create Admin User
	adminUser, err := h.Queries.CreateUser(r.Context(), db.CreateUserParams{
		SchoolID:    uuid.NullUUID{UUID: school.SchoolID, Valid: true},
		RoleID:      role.RoleID,
		FirstName:   req.AdminFirstName,
		LastName:    req.AdminLastName,
		Email:       req.AdminEmail,
		FirebaseUid: sql.NullString{String: req.AdminFirebaseUid, Valid: true},
		IsActive:    true,
	})
	if err != nil {
		slog.Error("Failed to create admin user", "error", err)
		middleware.SendError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create Default School Settings
	_, err = h.Queries.CreateSchoolSetting(r.Context(), school.SchoolID)
	if err != nil {
		slog.Error("Failed to create school settings", "error", err)
	}

	// Set Firebase Custom Claims
	claims := map[string]interface{}{
		"role":         role.RoleName,
		"schoolId":     school.SchoolID.String(),
		"schoolStatus": school.Status,
	}
	err = h.FirebaseAuth.SetCustomUserClaims(r.Context(), req.AdminFirebaseUid, claims)
	if err != nil {
		slog.Error("Failed to set custom claims", "error", err)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "School registration submitted for verification.",
		"school": map[string]interface{}{
			"id":      school.SchoolID,
			"name":    school.SchoolName,
			"status":  school.Status,
			"initial": school.SchoolInitial.String,
		},
		"adminUser": map[string]interface{}{
			"id":    adminUser.UserID,
			"email": adminUser.Email,
			"role":  role.RoleName,
		},
	})
}

func (h *SchoolHandler) VerifySchool(w http.ResponseWriter, r *http.Request) {
	schoolIDStr := chi.URLParam(r, "schoolId")
	schoolID, err := uuid.Parse(schoolIDStr)
	if err != nil {
		middleware.SendError(w, "Invalid school ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Status != "verified" && req.Status != "rejected" {
		middleware.SendError(w, "Invalid status. Must be 'verified' or 'rejected'", http.StatusBadRequest)
		return
	}

	school, err := h.Queries.GetSchoolWithAdmin(r.Context(), schoolID)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.SendError(w, "School not found", http.StatusNotFound)
			return
		}
		middleware.SendError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	updatedSchool, err := h.Queries.UpdateSchoolStatus(r.Context(), db.UpdateSchoolStatusParams{
		SchoolID: schoolID,
		Status:   req.Status,
	})
	if err != nil {
		middleware.SendError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Update Firebase claims for admin
	if school.AdminFirebaseUid.Valid {
		claims := map[string]interface{}{
			"schoolStatus": updatedSchool.Status,
			"schoolId":     updatedSchool.SchoolID.String(),
			"role":         school.AdminRoleName,
		}
		h.FirebaseAuth.SetCustomUserClaims(r.Context(), school.AdminFirebaseUid.String, claims)
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
		middleware.SendError(w, "Invalid school ID", http.StatusBadRequest)
		return
	}

	user, _ := middleware.GetUser(r.Context())
	if user.SchoolID.UUID != schoolID && !isParentCompanyAdmin(user.RoleName) {
		middleware.SendError(w, "Forbidden", http.StatusForbidden)
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
			middleware.SendError(w, "Settings not found", http.StatusNotFound)
			return
		}
		middleware.SendError(w, "Internal Server Error", http.StatusInternalServerError)
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
		middleware.SendError(w, "Invalid school ID", http.StatusBadRequest)
		return
	}

	user, _ := middleware.GetUser(r.Context())
	if user.SchoolID.UUID != schoolID || !isExecutiveAdmin(user.RoleName) {
		middleware.SendError(w, "Forbidden", http.StatusForbidden)
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
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
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
		middleware.SendError(w, "Internal Server Error", http.StatusInternalServerError)
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
