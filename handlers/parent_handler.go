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

	w.Header().Set("Content-Type", "application/json")
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parent)
}

func (h *ParentHandler) GetChildren(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	children, err := h.Queries.GetChildrenForParent(r.Context(), db.GetChildrenForParentParams{
		ParentUserID: userCtx.UserID,
		SchoolID:     schoolID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not fetch children", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(children)
}

// validateParentChild checks that the authenticated parent has a relationship with the given child
func (h *ParentHandler) validateParentChild(r *http.Request) (uuid.UUID, error) {
	userCtx, _ := middleware.GetUser(r.Context())
	childIDStr := chi.URLParam(r, "childId")
	childID, err := uuid.Parse(childIDStr)
	if err != nil {
		return uuid.Nil, err
	}

	_, err = h.Queries.ValidateParentChildRelationship(r.Context(), db.ValidateParentChildRelationshipParams{
		ParentUserID:  userCtx.UserID,
		StudentUserID: childID,
		SchoolID:      userCtx.SchoolID.UUID,
	})
	if err != nil {
		return uuid.Nil, err
	}
	return childID, nil
}

func (h *ParentHandler) GetChildAttendance(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	childID, err := h.validateParentChild(r)
	if err != nil {
		middleware.ForbiddenError(w, "You do not have access to this child's data", err)
		return
	}

	attendance, err := h.Queries.GetChildAttendance(r.Context(), db.GetChildAttendanceParams{
		StudentID: childID,
		SchoolID:  userCtx.SchoolID.UUID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not fetch attendance", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(attendance)
}

func (h *ParentHandler) GetChildGrades(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	childID, err := h.validateParentChild(r)
	if err != nil {
		middleware.ForbiddenError(w, "You do not have access to this child's data", err)
		return
	}

	grades, err := h.Queries.GetChildGrades(r.Context(), db.GetChildGradesParams{
		StudentID: childID,
		SchoolID:  userCtx.SchoolID.UUID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not fetch grades", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(grades)
}

func (h *ParentHandler) GetChildAssignments(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	childID, err := h.validateParentChild(r)
	if err != nil {
		middleware.ForbiddenError(w, "You do not have access to this child's data", err)
		return
	}

	assignments, err := h.Queries.GetChildAssignments(r.Context(), db.GetChildAssignmentsParams{
		StudentID: childID,
		SchoolID:  userCtx.SchoolID.UUID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not fetch assignments", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(assignments)
}

func (h *ParentHandler) GetChildFees(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	childID, err := h.validateParentChild(r)
	if err != nil {
		middleware.ForbiddenError(w, "You do not have access to this child's data", err)
		return
	}

	fees, err := h.Queries.GetChildFees(r.Context(), db.GetChildFeesParams{
		StudentID: childID,
		SchoolID:  userCtx.SchoolID.UUID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not fetch fees", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fees)
}
