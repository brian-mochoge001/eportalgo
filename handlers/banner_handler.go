package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type BannerHandler struct {
	Queries *db.Queries
}

func NewBannerHandler(q *db.Queries) *BannerHandler {
	return &BannerHandler{Queries: q}
}

func (h *BannerHandler) CreateBanner(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	
	var req struct {
		SchoolID  string `json:"school_id"`
		Title     string `json:"title"`
		ImageUrl  string `json:"image_url"`
		TargetUrl string `json:"target_url"`
		IsActive  bool   `json:"is_active"`
		Order     int32  `json:"order"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	// Only platform admins can create global banners (school_id is empty)
	// Executive admins can only create banners for their school
	var schoolID uuid.NullUUID
	if req.SchoolID != "" {
		sid, _ := uuid.Parse(req.SchoolID)
		if !isParentCompanyAdmin(userCtx.RoleName) && sid != userCtx.SchoolID.UUID {
			middleware.ForbiddenError(w, "You can only manage banners for your own school", nil)
			return
		}
		schoolID = uuid.NullUUID{UUID: sid, Valid: true}
	} else {
		if !isParentCompanyAdmin(userCtx.RoleName) {
			middleware.ForbiddenError(w, "Only platform admins can create global banners", nil)
			return
		}
	}

	banner, err := h.Queries.CreateBanner(r.Context(), db.CreateBannerParams{
		SchoolID:  schoolID,
		Title:     toNullString(req.Title),
		ImageUrl:  req.ImageUrl,
		TargetUrl: toNullString(req.TargetUrl),
		IsActive:  req.IsActive,
		Order:     req.Order,
	})

	if err != nil {
		middleware.InternalError(w, "Could not create banner", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(banner)
}

func (h *BannerHandler) GetActiveBanners(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	banners, err := h.Queries.GetActiveBanners(r.Context(), uuid.NullUUID{UUID: schoolID, Valid: true})
	if err != nil {
		middleware.InternalError(w, "Could not fetch banners", err)
		return
	}

	json.NewEncoder(w).Encode(banners)
}

func (h *BannerHandler) ListBanners(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID
	
	isSuperAdmin := isParentCompanyAdmin(userCtx.RoleName)

	banners, err := h.Queries.ListAllBanners(r.Context(), db.ListAllBannersParams{
		SchoolID:     uuid.NullUUID{UUID: schoolID, Valid: true},
		IsSuperAdmin: isSuperAdmin,
	})
	if err != nil {
		middleware.InternalError(w, "Could not fetch banners", err)
		return
	}

	json.NewEncoder(w).Encode(banners)
}

func (h *BannerHandler) UpdateBanner(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	bannerID, _ := uuid.Parse(idStr)

	var req struct {
		Title     string `json:"title"`
		ImageUrl  string `json:"image_url"`
		TargetUrl string `json:"target_url"`
		IsActive  bool   `json:"is_active"`
		Order     int32  `json:"order"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	updated, err := h.Queries.UpdateBanner(r.Context(), db.UpdateBannerParams{
		BannerID:  bannerID,
		Title:     toNullString(req.Title),
		ImageUrl:  req.ImageUrl,
		TargetUrl: toNullString(req.TargetUrl),
		IsActive:  req.IsActive,
		Order:     req.Order,
	})

	if err != nil {
		middleware.InternalError(w, "Could not update banner", err)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func (h *BannerHandler) DeleteBanner(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	bannerID, _ := uuid.Parse(idStr)

	err := h.Queries.DeleteBanner(r.Context(), bannerID)
	if err != nil {
		middleware.InternalError(w, "Could not delete banner", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
