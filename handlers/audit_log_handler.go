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

type AuditLogHandler struct {
	Queries *db.Queries
}

func NewAuditLogHandler(q *db.Queries) *AuditLogHandler {
	return &AuditLogHandler{Queries: q}
}

func (h *AuditLogHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	
	// Query params for filtering
	userIDStr := r.URL.Query().Get("userId")
	entityType := r.URL.Query().Get("entityType")
	entityIDStr := r.URL.Query().Get("entityId")
	action := r.URL.Query().Get("action")

	var userID uuid.NullUUID
	if userIDStr != "" {
		id, _ := uuid.Parse(userIDStr)
		userID = uuid.NullUUID{UUID: id, Valid: true}
	}

	var entityID uuid.NullUUID
	if entityIDStr != "" {
		id, _ := uuid.Parse(entityIDStr)
		entityID = uuid.NullUUID{UUID: id, Valid: true}
	}

	isSuperAdmin := isParentCompanyAdmin(userCtx.RoleName)

	logs, err := h.Queries.ListAuditLogs(r.Context(), db.ListAuditLogsParams{
		SchoolID:     userCtx.SchoolID,
		IsSuperAdmin: isSuperAdmin,
		UserID:       userID,
		EntityType:   sql.NullString{String: entityType, Valid: entityType != ""},
		EntityID:     entityID,
		Action:       sql.NullString{String: action, Valid: action != ""},
	})

	if err != nil {
		middleware.InternalError(w, "Could not fetch audit logs", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"results": len(logs),
		"data": map[string]interface{}{"auditLogs": logs},
	})
}

func (h *AuditLogHandler) GetAuditLog(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())

	auditLog, err := h.Queries.GetAuditLog(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.NotFoundError(w, "Audit log not found", err)
			return
		}
		middleware.InternalError(w, "Internal Server Error", err)
		return
	}

	// Auth check
	if !isParentCompanyAdmin(userCtx.RoleName) && auditLog.SchoolID.UUID != userCtx.SchoolID.UUID {
		middleware.ForbiddenError(w, "Forbidden", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data": map[string]interface{}{"auditLog": auditLog},
	})
}



