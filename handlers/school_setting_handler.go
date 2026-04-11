package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/sqlc-dev/pqtype"
)

type SchoolSettingHandler struct {
	Queries *db.Queries
}

func NewSchoolSettingHandler(q *db.Queries) *SchoolSettingHandler {
	return &SchoolSettingHandler{Queries: q}
}

func (h *SchoolSettingHandler) GetSchoolSettings(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	settings, err := h.Queries.GetSchoolSettings(r.Context(), schoolID)
	if err != nil {
		middleware.InternalError(w, "Could not fetch school settings", err)
		return
	}

	json.NewEncoder(w).Encode(settings)
}

func (h *SchoolSettingHandler) UpdateSchoolSettings(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

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

	settings, err := h.Queries.UpdateSchoolSettings(r.Context(), db.UpdateSchoolSettingsParams{
		SchoolID:             schoolID,
		BrandingLogoUrl:      toNullString(req.BrandingLogoUrl),
		BrandingColors:       req.BrandingColors,
		Timezone:             req.Timezone,
		Preferences:          req.Preferences,
		EmailTemplateConfigs: req.EmailTemplateConfigs,
		PaymentProviders:     pqtype.NullRawMessage{RawMessage: req.PaymentProviders, Valid: len(req.PaymentProviders) > 0},
	})

	if err != nil {
		middleware.InternalError(w, "Could not update school settings", err)
		return
	}

	json.NewEncoder(w).Encode(settings)
}



