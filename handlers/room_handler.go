package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type RoomHandler struct {
	Queries *db.Queries
}

func NewRoomHandler(q *db.Queries) *RoomHandler {
	return &RoomHandler{Queries: q}
}

func (h *RoomHandler) GetRooms(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	rooms, err := h.Queries.GetRoomsBySchool(r.Context(), schoolID)
	if err != nil {
		middleware.InternalError(w, "Could not fetch rooms", err)
		return
	}

	json.NewEncoder(w).Encode(rooms)
}

func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		RoomName     string `json:"room_name"`
		Capacity     int32  `json:"capacity"`
		RoomType     string `json:"room_type"`
		DepartmentID string `json:"department_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	room, err := h.Queries.CreateRoom(r.Context(), db.CreateRoomParams{
		SchoolID:     schoolID,
		RoomName:     req.RoomName,
		Capacity:     req.Capacity,
		RoomType:     db.RoomType(req.RoomType),
		DepartmentID: toNullUUID(req.DepartmentID),
	})

	if err != nil {
		middleware.InternalError(w, "Could not create room", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(room)
}

func (h *RoomHandler) GetRoomByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	roomID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	room, err := h.Queries.GetRoomByID(r.Context(), db.GetRoomByIDParams{
		RoomID:   roomID,
		SchoolID: schoolID,
	})
	if err != nil {
		middleware.NotFoundError(w, "Room not found", err)
		return
	}

	json.NewEncoder(w).Encode(room)
}

func (h *RoomHandler) UpdateRoom(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	roomID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		RoomName     string `json:"room_name"`
		Capacity     int32  `json:"capacity"`
		RoomType     string `json:"room_type"`
		DepartmentID string `json:"department_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	existingRoom, err := h.Queries.GetRoomByID(r.Context(), db.GetRoomByIDParams{
		RoomID:   roomID,
		SchoolID: schoolID,
	})
	if err != nil {
		middleware.NotFoundError(w, "Room not found", err)
		return
	}

	params := db.UpdateRoomParams{
		RoomID:       roomID,
		SchoolID:     schoolID,
		RoomName:     existingRoom.RoomName,
		Capacity:     existingRoom.Capacity,
		RoomType:     existingRoom.RoomType,
		DepartmentID: existingRoom.DepartmentID,
	}

	if req.RoomName != "" {
		params.RoomName = req.RoomName
	}
	if req.Capacity != 0 {
		params.Capacity = req.Capacity
	}
	if req.RoomType != "" {
		params.RoomType = db.RoomType(req.RoomType)
	}
	if req.DepartmentID != "" {
		params.DepartmentID = toNullUUID(req.DepartmentID)
	}

	updated, err := h.Queries.UpdateRoom(r.Context(), params)
	if err != nil {
		middleware.InternalError(w, "Could not update room", err)
		return
	}

	json.NewEncoder(w).Encode(updated)
}

func (h *RoomHandler) DeleteRoom(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	roomID, _ := uuid.Parse(idStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	err := h.Queries.DeleteRoom(r.Context(), db.DeleteRoomParams{
		RoomID:   roomID,
		SchoolID: schoolID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not delete room", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}



