package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ClassRepresentativeHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewClassRepresentativeHandler(q *db.Queries, d *sql.DB) *ClassRepresentativeHandler {
	return &ClassRepresentativeHandler{Queries: q, DB: d}
}

func (h *ClassRepresentativeHandler) GetClassRepresentatives(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	academicClassIDStr := r.URL.Query().Get("academicClassId")
	var academicClassID uuid.NullUUID
	if academicClassIDStr != "" {
		id, _ := uuid.Parse(academicClassIDStr)
		academicClassID = uuid.NullUUID{UUID: id, Valid: true}
	}

	reps, err := h.Queries.GetClassRepresentativesBySchool(r.Context(), db.GetClassRepresentativesBySchoolParams{
		SchoolID:        schoolID,
		AcademicClassID: academicClassID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not fetch class representatives", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"results": len(reps),
		"data":    map[string]interface{}{"classRepresentatives": reps},
	})
}

func (h *ClassRepresentativeHandler) GetClassRepresentativeByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	rep, err := h.Queries.GetClassRepresentativeByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			middleware.NotFoundError(w, "Class representative not found", err)
			return
		}
		middleware.InternalError(w, "Internal Server Error", err)
		return
	}

	if rep.SchoolID != schoolID {
		middleware.ForbiddenError(w, "Forbidden", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   map[string]interface{}{"classRepresentative": rep},
	})
}

func (h *ClassRepresentativeHandler) CreateClassRepresentative(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		StudentUserID               string `json:"student_user_id"`
		AcademicClassID             string `json:"academic_class_id"`
		CanCommunicateTeacher       bool   `json:"can_communicate_teacher"`
		CanCommunicateDepartmentHead bool   `json:"can_communicate_department_head"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	tx, _ := h.DB.Begin()
	defer tx.Rollback()
	qtx := h.Queries.WithTx(tx)

	studentID, _ := uuid.Parse(req.StudentUserID)
	classID, _ := uuid.Parse(req.AcademicClassID)

	// Verify student
	student, err := qtx.GetUser(r.Context(), db.GetUserParams{
		UserID:   studentID,
		SchoolID: uuid.NullUUID{UUID: schoolID, Valid: true},
	})
	if err != nil || student.SchoolID.UUID != schoolID {
		middleware.NotFoundError(w, "Student not found in your school", err)
		return
	}

	// Verify class and teacher
	classDetails, err := qtx.GetClassWithDetails(r.Context(), db.GetClassWithDetailsParams{
		ClassID:  classID,
		SchoolID: schoolID,
	})
	if err != nil {
		middleware.NotFoundError(w, "Class not found", err)
		return
	}

	if classDetails.TeacherUserID.UUID != userCtx.UserID {
		middleware.ForbiddenError(w, "Only the assigned teacher can appoint a representative", err)
		return
	}

	rep, err := qtx.CreateClassRepresentative(r.Context(), db.CreateClassRepresentativeParams{
		StudentUserID:               studentID,
		AcademicClassID:             classID,
		CanCommunicateTeacher:       req.CanCommunicateTeacher,
		CanCommunicateDepartmentHead: req.CanCommunicateDepartmentHead,
	})

	if err != nil {
		middleware.InternalError(w, "Could not create representative", err)
		return
	}

	// Setup chats
	if req.CanCommunicateTeacher && classDetails.TeacherUserID.Valid {
		h.setupChat(r.Context(), qtx, schoolID, student, classDetails.TeacherUserID.UUID, "Teacher")
	}
	if req.CanCommunicateDepartmentHead && classDetails.HeadOfDepartmentID.Valid {
		h.setupChat(r.Context(), qtx, schoolID, student, classDetails.HeadOfDepartmentID.UUID, "Department Head")
	}

	tx.Commit()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   map[string]interface{}{"classRepresentative": rep},
	})
}

func (h *ClassRepresentativeHandler) setupChat(ctx context.Context, qtx *db.Queries, schoolID uuid.UUID, student db.User, targetID uuid.UUID, targetType string) {
	room, _ := qtx.CreateChatRoom(ctx, db.CreateChatRoomParams{
		SchoolID:         schoolID,
		ChatName:         fmt.Sprintf("Class Rep (%s) - %s", student.FirstName, targetType),
		ChatType:         "one_on_one",
		CreatedByUserID:  uuid.NullUUID{UUID: student.UserID, Valid: true},
		IsActive:         true,
	})

	qtx.AddChatParticipant(ctx, db.AddChatParticipantParams{
		SchoolID:   schoolID,
		ChatRoomID: room.ChatRoomID,
		UserID:     student.UserID,
	})
	qtx.AddChatParticipant(ctx, db.AddChatParticipantParams{
		SchoolID:   schoolID,
		ChatRoomID: room.ChatRoomID,
		UserID:     targetID,
	})
}

func (h *ClassRepresentativeHandler) UpdateClassRepresentative(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		CanCommunicateTeacher       bool `json:"can_communicate_teacher"`
		CanCommunicateDepartmentHead bool `json:"can_communicate_department_head"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	rep, err := h.Queries.GetClassRepresentativeByID(r.Context(), id)
	if err != nil || rep.SchoolID != schoolID {
		middleware.NotFoundError(w, "Representative not found", err)
		return
	}

	updated, err := h.Queries.UpdateClassRepresentative(r.Context(), db.UpdateClassRepresentativeParams{
		ClassRepID:                  id,
		CanCommunicateTeacher:       req.CanCommunicateTeacher,
		CanCommunicateDepartmentHead: req.CanCommunicateDepartmentHead,
	})

	if err != nil {
		middleware.InternalError(w, "Could not update representative", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   map[string]interface{}{"classRepresentative": updated},
	})
}

func (h *ClassRepresentativeHandler) DeleteClassRepresentative(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	rep, err := h.Queries.GetClassRepresentativeByID(r.Context(), id)
	if err != nil || rep.SchoolID != schoolID {
		middleware.NotFoundError(w, "Representative not found", err)
		return
	}

	tx, _ := h.DB.Begin()
	defer tx.Rollback()
	qtx := h.Queries.WithTx(tx)

	qtx.DeactivateChatRoomsByParticipant(r.Context(), rep.StudentUserID)
	qtx.DeleteClassRepresentative(r.Context(), id)

	tx.Commit()

	w.WriteHeader(http.StatusNoContent)
}



