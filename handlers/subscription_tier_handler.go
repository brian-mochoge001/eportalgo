package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type SubscriptionTierHandler struct {
	Queries *db.Queries
}

func NewSubscriptionTierHandler(q *db.Queries) *SubscriptionTierHandler {
	return &SubscriptionTierHandler{Queries: q}
}

func (h *SubscriptionTierHandler) GetSubscriptionTiers(w http.ResponseWriter, r *http.Request) {
	tiers, err := h.Queries.GetSubscriptionTiers(r.Context())
	if err != nil {
		middleware.InternalError(w, "Could not fetch subscription tiers", err)
		return
	}

	json.NewEncoder(w).Encode(tiers)
}

func (h *SubscriptionTierHandler) GetSubscriptionTierByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	tierID, _ := uuid.Parse(idStr)

	tier, err := h.Queries.GetSubscriptionTierByID(r.Context(), tierID)
	if err != nil {
		middleware.NotFoundError(w, "Subscription tier not found", err)
		return
	}

	json.NewEncoder(w).Encode(tier)
}


