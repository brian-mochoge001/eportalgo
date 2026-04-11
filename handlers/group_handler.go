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

type GroupHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewGroupHandler(q *db.Queries, d *sql.DB) *GroupHandler {
	return &GroupHandler{Queries: q, DB: d}
}

func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		Name        string      `json:"name"`
		Description string      `json:"description"`
		StudentIDs  []uuid.UUID `json:"student_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.StudentIDs) > 4 {
		middleware.SendError(w, "A group can have a maximum of 5 students, including the creator", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		middleware.SendError(w, "Could not start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	qtx := h.Queries.WithTx(tx)

	// Create Chat Room
	chatRoom, err := qtx.CreateChatRoom(r.Context(), db.CreateChatRoomParams{
		SchoolID:        schoolID,
		ChatName:        req.Name,
		ChatType:        "GROUP",
		CreatedByUserID: uuid.NullUUID{UUID: userCtx.UserID, Valid: true},
		IsActive:        true,
	})
	if err != nil {
		middleware.SendError(w, "Could not create chat room", http.StatusInternalServerError)
		return
	}

	// Create Group
	group, err := qtx.CreateGroup(r.Context(), db.CreateGroupParams{
		SchoolID:         schoolID,
		Name:             req.Name,
		Description:      toNullString(req.Description),
		CreatedByUserID:  userCtx.UserID,
		IsTeacherCreated: false,
		ChatRoomID:       uuid.NullUUID{UUID: chatRoom.ChatRoomID, Valid: true},
	})
	if err != nil {
		middleware.SendError(w, "Could not create group", http.StatusInternalServerError)
		return
	}

	// Add Creator
	_, err = qtx.AddGroupMember(r.Context(), db.AddGroupMemberParams{
		GroupID: group.GroupID,
		UserID:  userCtx.UserID,
		Status:  "accepted",
	})
	if err != nil {
		middleware.SendError(w, "Could not add creator to group", http.StatusInternalServerError)
		return
	}

	// Add invited students
	for _, studentID := range req.StudentIDs {
		_, err = qtx.AddGroupMember(r.Context(), db.AddGroupMemberParams{
			GroupID: group.GroupID,
			UserID:  studentID,
			Status:  "pending",
		})
		if err != nil {
			middleware.SendError(w, "Could not invite student to group", http.StatusInternalServerError)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		middleware.SendError(w, "Could not commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(group)
}

func (h *GroupHandler) TeacherCreateGroup(w http.ResponseWriter, r *http.Request) {
	// Teachers can create larger groups
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		Name        string      `json:"name"`
		Description string      `json:"description"`
		StudentIDs  []uuid.UUID `json:"student_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		middleware.SendError(w, "Could not start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	qtx := h.Queries.WithTx(tx)

	chatRoom, _ := qtx.CreateChatRoom(r.Context(), db.CreateChatRoomParams{
		SchoolID:        schoolID,
		ChatName:        req.Name,
		ChatType:        "GROUP",
		CreatedByUserID: uuid.NullUUID{UUID: userCtx.UserID, Valid: true},
		IsActive:        true,
	})

	group, err := qtx.CreateGroup(r.Context(), db.CreateGroupParams{
		SchoolID:         schoolID,
		Name:             req.Name,
		Description:      toNullString(req.Description),
		CreatedByUserID:  userCtx.UserID,
		IsTeacherCreated: true,
		ChatRoomID:       uuid.NullUUID{UUID: chatRoom.ChatRoomID, Valid: true},
	})

	qtx.AddGroupMember(r.Context(), db.AddGroupMemberParams{
		GroupID: group.GroupID,
		UserID:  userCtx.UserID,
		Status:  "accepted",
	})

	for _, sid := range req.StudentIDs {
		qtx.AddGroupMember(r.Context(), db.AddGroupMemberParams{
			GroupID: group.GroupID,
			UserID:  sid,
			Status:  "pending",
		})
	}

	tx.Commit()
	json.NewEncoder(w).Encode(group)
}

func (h *GroupHandler) RespondToGroupInvitation(w http.ResponseWriter, r *http.Request) {
	var req struct {
		GroupID uuid.UUID `json:"group_id"`
		Accept  bool      `json:"accept"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	status := "declined"
	if req.Accept {
		status = "accepted"
	}

	// Update logic would go here - for now just a placeholder
	// In a real app, you'd have an UpdateGroupMemberStatus query
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}

func (h *GroupHandler) GetGroupMembers(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	groupID, _ := uuid.Parse(idStr)

	// Fetch members logic...
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"group_id": groupID, "members": []string{}})
}
