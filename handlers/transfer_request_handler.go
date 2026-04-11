package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type TransferRequestHandler struct {
	Queries *db.Queries
}

func NewTransferRequestHandler(q *db.Queries) *TransferRequestHandler {
	return &TransferRequestHandler{Queries: q}
}

func (h *TransferRequestHandler) CreateTransferRequest(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())

	var req struct {
		EntityType        string `json:"entity_type"`
		EntityID          string `json:"entity_id"`
		SourceSchoolID    string `json:"source_school_id"`
		DestinationSchoolID string `json:"destination_school_id"`
		Notes             string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	entityID, _ := uuid.Parse(req.EntityID)
	sourceSchoolID, _ := uuid.Parse(req.SourceSchoolID)
	destinationSchoolID, _ := uuid.Parse(req.DestinationSchoolID)

	transferRequest, err := h.Queries.CreateTransferRequest(r.Context(), db.CreateTransferRequestParams{
		EntityType:        req.EntityType,
		EntityID:          entityID,
		SourceSchoolID:    sourceSchoolID,
		DestinationSchoolID: destinationSchoolID,
		InitiatedByUserID: userCtx.UserID,
		Notes:             toNullString(req.Notes),
	})

	if err != nil {
		middleware.SendError(w, "Could not create transfer request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transferRequest)
}

func (h *TransferRequestHandler) GetTransferRequests(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	status := r.URL.Query().Get("status")
	entityType := r.URL.Query().Get("entityType")
	sourceSchoolIDStr := r.URL.Query().Get("sourceSchoolId")
	destinationSchoolIDStr := r.URL.Query().Get("destinationSchoolId")

	var sourceSchoolID uuid.NullUUID
	if sourceSchoolIDStr != "" {
		if id, err := uuid.Parse(sourceSchoolIDStr); err == nil {
			sourceSchoolID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	var destinationSchoolID uuid.NullUUID
	if destinationSchoolIDStr != "" {
		if id, err := uuid.Parse(destinationSchoolIDStr); err == nil {
			destinationSchoolID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	// Authorization check: Super admins can see all, others are limited to their school (source or destination)
	isSuperAdmin := false
	for _, role := range []string{"Developer", "DB Manager", "Data Analyst", "Support Staff"} {
		if userCtx.RoleName == role {
			isSuperAdmin = true
			break
		}
	}

	if !isSuperAdmin {
		if !sourceSchoolID.Valid && !destinationSchoolID.Valid {
			// If not super admin and no specific school filter, default to user's school
			if userCtx.SchoolID.Valid {
				sourceSchoolID = userCtx.SchoolID
				destinationSchoolID = userCtx.SchoolID
			}
		}
	}

	transferRequests, err := h.Queries.ListTransferRequests(r.Context(), db.ListTransferRequestsParams{
		Status:              toNullString(status),
		EntityType:          toNullString(entityType),
		SourceSchoolID:      sourceSchoolID,
		DestinationSchoolID: destinationSchoolID,
		SchoolID:            uuid.NullUUID{UUID: schoolID, Valid: !isSuperAdmin}, // Use user's school if not super admin
	})
	if err != nil {
		middleware.SendError(w, "Could not fetch transfer requests", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(transferRequests)
}

func (h *TransferRequestHandler) GetTransferRequestByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	transferID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())

	transferRequest, err := h.Queries.GetTransferRequestByID(r.Context(), transferID)
	if err != nil {
		middleware.SendError(w, "Transfer request not found", http.StatusNotFound)
		return
	}

	// Authorization check
	isSuperAdmin := false
	for _, role := range []string{"Developer", "DB Manager", "Data Analyst", "Support Staff"} {
		if userCtx.RoleName == role {
			isSuperAdmin = true
			break
		}
	}
	if !isSuperAdmin && transferRequest.SourceSchoolID != userCtx.SchoolID.UUID && transferRequest.DestinationSchoolID != userCtx.SchoolID.UUID {
		middleware.SendError(w, "Not authorized to view this transfer request", http.StatusForbidden)
		return
	}

	json.NewEncoder(w).Encode(transferRequest)
}

func (h *TransferRequestHandler) UpdateTransferRequest(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	transferID, _ := uuid.Parse(idStr)

	var req struct {
		Status string `json:"status"`
		Notes  string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	updated, err := h.Queries.UpdateTransferRequestStatus(r.Context(), db.UpdateTransferRequestStatusParams{
		TransferID:     transferID,
		Status:         req.Status,
		CompletionDate: sql.NullTime{Time: time.Now(), Valid: true},
		Notes:          toNullString(req.Notes),
	})
	if err != nil {
		middleware.SendError(w, "Could not update transfer request", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func (h *TransferRequestHandler) DeleteTransferRequest(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	transferID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	err := h.Queries.DeleteTransferRequest(r.Context(), db.DeleteTransferRequestParams{
		TransferID:     transferID,
		SourceSchoolID: schoolID,
	})
	if err != nil {
		middleware.SendError(w, "Could not delete transfer request", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
