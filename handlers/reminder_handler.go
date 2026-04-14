package handlers

import (
	"encoding/json"
	"net/http"
	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ReminderHandler struct {
	Queries *db.Queries
}

func NewReminderHandler(q *db.Queries) *ReminderHandler {
	return &ReminderHandler{Queries: q}
}

func (h *ReminderHandler) ListReminderLists(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	lists, err := h.Queries.ListReminderLists(r.Context(), db.ListReminderListsParams{
		UserID:   userCtx.UserID,
		SchoolID: userCtx.SchoolID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not fetch reminder lists", err)
		return
	}
	json.NewEncoder(w).Encode(lists)
}

func (h *ReminderHandler) CreateReminderList(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	var req struct {
		Title string `json:"title"`
		Color string `json:"color"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}
	list, err := h.Queries.CreateReminderList(r.Context(), db.CreateReminderListParams{
		SchoolID: userCtx.SchoolID,
		UserID:   userCtx.UserID,
		Title:    req.Title,
		Color:    toNullString(req.Color),
	})
	if err != nil {
		middleware.InternalError(w, "Could not create reminder list", err)
		return
	}
	json.NewEncoder(w).Encode(list)
}

func (h *ReminderHandler) ListReminders(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	listIDStr := chi.URLParam(r, "listId")
	listID, err := uuid.Parse(listIDStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid list ID", err)
		return
	}
	reminders, err := h.Queries.ListRemindersByList(r.Context(), db.ListRemindersByListParams{
		ListID: listID,
		UserID: userCtx.UserID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not fetch reminders", err)
		return
	}
	json.NewEncoder(w).Encode(reminders)
}

func (h *ReminderHandler) CreateReminder(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	var req struct {
		ListID   string `json:"list_id"`
		Title    string `json:"title"`
		Notes    string `json:"notes"`
		DueDate  string `json:"due_date"`
		Priority string `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}
	listID, _ := uuid.Parse(req.ListID)
	dueDate, _ := parseDate(req.DueDate)
	reminder, err := h.Queries.CreateReminder(r.Context(), db.CreateReminderParams{
		ListID:   listID,
		UserID:   userCtx.UserID,
		Title:    req.Title,
		Notes:    toNullString(req.Notes),
		DueDate:  dueDate,
		Priority: toNullString(req.Priority),
	})
	if err != nil {
		middleware.InternalError(w, "Could not create reminder", err)
		return
	}
	json.NewEncoder(w).Encode(reminder)
}

func (h *ReminderHandler) UpdateReminderStatus(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	reminderIDStr := chi.URLParam(r, "id")
	reminderID, _ := uuid.Parse(reminderIDStr)
	var req struct {
		IsCompleted bool `json:"is_completed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}
	updated, err := h.Queries.UpdateReminderStatus(r.Context(), db.UpdateReminderStatusParams{
		ReminderID:  reminderID,
		IsCompleted: req.IsCompleted,
		UserID:      userCtx.UserID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not update reminder status", err)
		return
	}
	json.NewEncoder(w).Encode(updated)
}

func (h *ReminderHandler) DeleteReminder(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	reminderIDStr := chi.URLParam(r, "id")
	reminderID, _ := uuid.Parse(reminderIDStr)
	err := h.Queries.DeleteReminder(r.Context(), db.DeleteReminderParams{
		ReminderID: reminderID,
		UserID:     userCtx.UserID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not delete reminder", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
