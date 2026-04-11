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

type ExternalCertificationHandler struct {
	Queries *db.Queries
}

func NewExternalCertificationHandler(q *db.Queries) *ExternalCertificationHandler {
	return &ExternalCertificationHandler{Queries: q}
}

func (h *ExternalCertificationHandler) CreateExternalCertification(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	var req struct {
		Name            string `json:"name"`
		Issuer          string `json:"issuer"`
		CredentialID    string `json:"credential_id"`
		VerificationURL string `json:"verification_url"`
		IssueDate       string `json:"issue_date"`
		ExpiryDate      string `json:"expiry_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	if req.Name == "" || req.Issuer == "" || req.VerificationURL == "" {
		middleware.ValidationError(w, "Name, issuer, and verification URL are required", nil)
		return
	}

	var issueDate sql.NullTime
	if req.IssueDate != "" {
		if t, err := time.Parse(time.RFC3339, req.IssueDate); err == nil {
			issueDate = sql.NullTime{Time: t, Valid: true}
		} else if t, err := time.Parse("2006-01-02", req.IssueDate); err == nil {
			issueDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	var expiryDate sql.NullTime
	if req.ExpiryDate != "" {
		if t, err := time.Parse(time.RFC3339, req.ExpiryDate); err == nil {
			expiryDate = sql.NullTime{Time: t, Valid: true}
		} else if t, err := time.Parse("2006-01-02", req.ExpiryDate); err == nil {
			expiryDate = sql.NullTime{Time: t, Valid: true}
		}
	}

	cert, err := h.Queries.CreateExternalCertification(r.Context(), db.CreateExternalCertificationParams{
		StudentID:       userCtx.UserID,
		Name:            req.Name,
		Issuer:          req.Issuer,
		CredentialID:    sql.NullString{String: req.CredentialID, Valid: req.CredentialID != ""},
		VerificationUrl: sql.NullString{String: req.VerificationURL, Valid: req.VerificationURL != ""},
		IssueDate:       issueDate,
		ExpiryDate:      expiryDate,
		IsVerified:      false,
	})

	if err != nil {
		middleware.InternalError(w, "Could not add external certification", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cert)
}

func (h *ExternalCertificationHandler) GetExternalCertifications(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	studentIDStr := r.URL.Query().Get("studentId")

	var studentID uuid.NullUUID
	if userCtx.RoleName == "Student" {
		studentID = uuid.NullUUID{UUID: userCtx.UserID, Valid: true}
	} else if studentIDStr != "" {
		if id, err := uuid.Parse(studentIDStr); err == nil {
			studentID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	certs, err := h.Queries.GetExternalCertifications(r.Context(), studentID.UUID)
	if err != nil {
		middleware.InternalError(w, "Could not fetch external certifications", err)
		return
	}

	json.NewEncoder(w).Encode(certs)
}

func (h *ExternalCertificationHandler) GetExternalCertificationByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	certID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid certification ID", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())

	cert, err := h.Queries.GetExternalCertificationByID(r.Context(), certID)
	if err != nil {
		middleware.NotFoundError(w, "External certification not found", err)
		return
	}

	if userCtx.RoleName == "Student" && cert.StudentID != userCtx.UserID {
		middleware.ForbiddenError(w, "Not authorized to view this external certification", err)
		return
	}

	json.NewEncoder(w).Encode(cert)
}

func (h *ExternalCertificationHandler) UpdateExternalCertification(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	certID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid certification ID", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())

	var req struct {
		Name            string `json:"name"`
		Issuer          string `json:"issuer"`
		CredentialID    string `json:"credential_id"`
		VerificationURL string `json:"verification_url"`
		IssueDate       string `json:"issue_date"`
		ExpiryDate      string `json:"expiry_date"`
		IsVerified      *bool  `json:"is_verified"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	existingCert, err := h.Queries.GetExternalCertificationByID(r.Context(), certID)
	if err != nil {
		middleware.NotFoundError(w, "External certification not found", err)
		return
	}

	isAdmin := userCtx.RoleName == "Executive Administrator" || userCtx.RoleName == "Academic Administrator"
	if existingCert.StudentID != userCtx.UserID && !isAdmin {
		middleware.ForbiddenError(w, "Not authorized to update this external certification", err)
		return
	}

	params := db.UpdateExternalCertificationParams{
		CertID:          certID,
		Name:            existingCert.Name,
		Issuer:          existingCert.Issuer,
		CredentialID:    existingCert.CredentialID,
		VerificationUrl: existingCert.VerificationUrl,
		IssueDate:       existingCert.IssueDate,
		ExpiryDate:      existingCert.ExpiryDate,
		IsVerified:      existingCert.IsVerified,
	}

	if req.Name != "" {
		params.Name = req.Name
	}
	if req.Issuer != "" {
		params.Issuer = req.Issuer
	}
	if req.CredentialID != "" {
		params.CredentialID = sql.NullString{String: req.CredentialID, Valid: true}
	}
	if req.VerificationURL != "" {
		params.VerificationUrl = sql.NullString{String: req.VerificationURL, Valid: true}
	}
	if req.IssueDate != "" {
		if t, err := time.Parse(time.RFC3339, req.IssueDate); err == nil {
			params.IssueDate = sql.NullTime{Time: t, Valid: true}
		} else if t, err := time.Parse("2006-01-02", req.IssueDate); err == nil {
			params.IssueDate = sql.NullTime{Time: t, Valid: true}
		}
	}
	if req.ExpiryDate != "" {
		if t, err := time.Parse(time.RFC3339, req.ExpiryDate); err == nil {
			params.ExpiryDate = sql.NullTime{Time: t, Valid: true}
		} else if t, err := time.Parse("2006-01-02", req.ExpiryDate); err == nil {
			params.ExpiryDate = sql.NullTime{Time: t, Valid: true}
		}
	}
	if req.IsVerified != nil && isAdmin {
		params.IsVerified = *req.IsVerified
	}

	updated, err := h.Queries.UpdateExternalCertification(r.Context(), params)
	if err != nil {
		middleware.InternalError(w, "Could not update external certification", err)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func (h *ExternalCertificationHandler) DeleteExternalCertification(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	certID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid certification ID", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())

	existingCert, err := h.Queries.GetExternalCertificationByID(r.Context(), certID)
	if err != nil {
		middleware.NotFoundError(w, "External certification not found", err)
		return
	}

	isAdmin := userCtx.RoleName == "Executive Administrator" || userCtx.RoleName == "Academic Administrator"
	if existingCert.StudentID != userCtx.UserID && !isAdmin {
		middleware.ForbiddenError(w, "Not authorized to delete this external certification", err)
		return
	}

	err = h.Queries.DeleteExternalCertification(r.Context(), certID)
	if err != nil {
		middleware.InternalError(w, "Could not delete external certification", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}



