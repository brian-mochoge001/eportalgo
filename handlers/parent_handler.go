package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ParentHandler struct {
	Queries *db.Queries
}

func NewParentHandler(q *db.Queries) *ParentHandler {
	return &ParentHandler{Queries: q}
}

func (h *ParentHandler) GetParents(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	parents, err := h.Queries.GetParentsBySchool(r.Context(), schoolID)
	if err != nil {
		middleware.InternalError(w, "Could not fetch parents", err)
		return
	}

	json.NewEncoder(w).Encode(parents)
}

func (h *ParentHandler) GetParentByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	parentID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	parent, err := h.Queries.GetParentByUserID(r.Context(), db.GetParentByUserIDParams{
		UserID:   parentID,
		SchoolID: schoolID,
	})
	if err != nil {
		middleware.NotFoundError(w, "Parent not found", err)
		return
	}

	json.NewEncoder(w).Encode(parent)
}



