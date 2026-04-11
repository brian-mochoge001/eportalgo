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

type BillingContactHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewBillingContactHandler(q *db.Queries, d *sql.DB) *BillingContactHandler {
	return &BillingContactHandler{Queries: q, DB: d}
}

func (h *BillingContactHandler) GetBillingContacts(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	contacts, err := h.Queries.GetBillingContactsBySchool(r.Context(), schoolID)
	if err != nil {
		middleware.SendError(w, "Could not fetch billing contacts", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"results": len(contacts),
		"data": map[string]interface{}{"billingContacts": contacts},
	})
}

func (h *BillingContactHandler) GetBillingContactByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	contact, err := h.Queries.GetBillingContactByID(r.Context(), db.GetBillingContactByIDParams{
		BillingContactID: id,
		SchoolID:         schoolID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.SendError(w, "Billing contact not found", http.StatusNotFound)
			return
		}
		middleware.SendError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{"billingContact": contact},
	})
}

func (h *BillingContactHandler) CreateBillingContact(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phone_number"`
		Role        string `json:"role"`
		IsPrimary   bool   `json:"is_primary"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tx, _ := h.DB.Begin()
	defer tx.Rollback()
	qtx := h.Queries.WithTx(tx)

	if req.IsPrimary {
		qtx.ResetPrimaryBillingContact(r.Context(), db.ResetPrimaryBillingContactParams{
			SchoolID:         schoolID,
			BillingContactID: uuid.NullUUID{Valid: false},
		})
	}

	contact, err := qtx.CreateBillingContact(r.Context(), db.CreateBillingContactParams{
		SchoolID:    schoolID,
		Name:        req.Name,
		Email:       req.Email,
		PhoneNumber: sql.NullString{String: req.PhoneNumber, Valid: req.PhoneNumber != ""},
		Role:        sql.NullString{String: req.Role, Valid: req.Role != ""},
		IsPrimary:   req.IsPrimary,
	})

	if err != nil {
		middleware.SendError(w, "Could not create billing contact", http.StatusInternalServerError)
		return
	}

	tx.Commit()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{"billingContact": contact},
	})
}

func (h *BillingContactHandler) UpdateBillingContact(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phone_number"`
		Role        string `json:"role"`
		IsPrimary   bool   `json:"is_primary"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	tx, _ := h.DB.Begin()
	defer tx.Rollback()
	qtx := h.Queries.WithTx(tx)

	if req.IsPrimary {
		qtx.ResetPrimaryBillingContact(r.Context(), db.ResetPrimaryBillingContactParams{
			SchoolID:         schoolID,
			BillingContactID: uuid.NullUUID{UUID: id, Valid: true},
		})
	}

	contact, err := qtx.UpdateBillingContact(r.Context(), db.UpdateBillingContactParams{
		BillingContactID: id,
		Name:             req.Name,
		Email:            req.Email,
		PhoneNumber:      sql.NullString{String: req.PhoneNumber, Valid: req.PhoneNumber != ""},
		Role:             sql.NullString{String: req.Role, Valid: req.Role != ""},
		IsPrimary:        req.IsPrimary,
		SchoolID:         schoolID,
	})

	if err != nil {
		middleware.SendError(w, "Could not update billing contact", http.StatusInternalServerError)
		return
	}

	tx.Commit()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{"billingContact": contact},
	})
}

func (h *BillingContactHandler) DeleteBillingContact(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	err := h.Queries.DeleteBillingContact(r.Context(), db.DeleteBillingContactParams{
		BillingContactID: id,
		SchoolID:         schoolID,
	})

	if err != nil {
		middleware.SendError(w, "Could not delete billing contact", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
