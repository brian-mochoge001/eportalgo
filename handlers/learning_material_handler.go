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

type LearningMaterialHandler struct {
	Queries *db.Queries
}

func NewLearningMaterialHandler(q *db.Queries) *LearningMaterialHandler {
	return &LearningMaterialHandler{Queries: q}
}

func (h *LearningMaterialHandler) CreateLearningMaterial(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID
	userID := userCtx.UserID

	var req struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		FileUrl      string `json:"file_url"`
		MaterialType string `json:"material_type"`
		ClassID      string `json:"class_id"`
		CourseID     string `json:"course_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	if req.Title == "" || req.FileUrl == "" || req.MaterialType == "" {
		middleware.ValidationError(w, "Title, file URL, and material type are required", nil)
		return
	}

	var classID uuid.NullUUID
	if req.ClassID != "" {
		if id, err := uuid.Parse(req.ClassID); err == nil {
			classID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	var courseID uuid.NullUUID
	if req.CourseID != "" {
		if id, err := uuid.Parse(req.CourseID); err == nil {
			courseID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	material, err := h.Queries.CreateLearningMaterial(r.Context(), db.CreateLearningMaterialParams{
		SchoolID:          schoolID,
		UploadedByUserID: userID,
		Title:             req.Title,
		Description:       sql.NullString{String: req.Description, Valid: req.Description != ""},
		FileUrl:           sql.NullString{String: req.FileUrl, Valid: req.FileUrl != ""},
		MaterialType:      db.MaterialType(req.MaterialType),
		ClassID:           classID,
		CourseID:          courseID,
	})

	if err != nil {
		middleware.InternalError(w, "Could not create learning material", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(material)
}

func (h *LearningMaterialHandler) GetLearningMaterials(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	classIDStr := r.URL.Query().Get("classId")
	courseIDStr := r.URL.Query().Get("courseId")

	var classID uuid.NullUUID
	if classIDStr != "" {
		if id, err := uuid.Parse(classIDStr); err == nil {
			classID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	var courseID uuid.NullUUID
	if courseIDStr != "" {
		if id, err := uuid.Parse(courseIDStr); err == nil {
			courseID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	materials, err := h.Queries.GetLearningMaterials(r.Context(), db.GetLearningMaterialsParams{
		SchoolID: schoolID,
		ClassID:  classID,
		CourseID: courseID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not fetch learning materials", err)
		return
	}

	json.NewEncoder(w).Encode(materials)
}

func (h *LearningMaterialHandler) GetLearningMaterialByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	materialID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid material ID", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	material, err := h.Queries.GetLearningMaterialByID(r.Context(), db.GetLearningMaterialByIDParams{
		MaterialID: materialID,
		SchoolID:   schoolID,
	})
	if err != nil {
		middleware.NotFoundError(w, "Learning material not found", err)
		return
	}

	json.NewEncoder(w).Encode(material)
}

func (h *LearningMaterialHandler) UpdateLearningMaterial(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	materialID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid material ID", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		Title        string `json:"title"`
		Description  string `json:"description"`
		FileUrl      string `json:"file_url"`
		MaterialType string `json:"material_type"`
		ClassID      string `json:"class_id"`
		CourseID     string `json:"course_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	existingMaterial, err := h.Queries.GetLearningMaterialByID(r.Context(), db.GetLearningMaterialByIDParams{
		MaterialID: materialID,
		SchoolID:   schoolID,
	})
	if err != nil {
		middleware.NotFoundError(w, "Learning material not found", err)
		return
	}

	params := db.UpdateLearningMaterialParams{
		MaterialID:   materialID,
		SchoolID:     schoolID,
		Title:        existingMaterial.Title,
		Description:  existingMaterial.Description,
		FileUrl:      existingMaterial.FileUrl,
		MaterialType: existingMaterial.MaterialType,
		ClassID:      existingMaterial.ClassID,
		CourseID:     existingMaterial.CourseID,
	}

	if req.Title != "" {
		params.Title = req.Title
	}
	if req.Description != "" {
		params.Description = sql.NullString{String: req.Description, Valid: true}
	}
	if req.FileUrl != "" {
		params.FileUrl = sql.NullString{String: req.FileUrl, Valid: true}
	}
	if req.MaterialType != "" {
		params.MaterialType = db.MaterialType(req.MaterialType)
	}
	if req.ClassID != "" {
		if id, err := uuid.Parse(req.ClassID); err == nil {
			params.ClassID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}
	if req.CourseID != "" {
		if id, err := uuid.Parse(req.CourseID); err == nil {
			params.CourseID = uuid.NullUUID{UUID: id, Valid: true}
		}
	}

	updatedMaterial, err := h.Queries.UpdateLearningMaterial(r.Context(), params)
	if err != nil {
		middleware.InternalError(w, "Could not update learning material", err)
		return
	}

	json.NewEncoder(w).Encode(updatedMaterial)
}

func (h *LearningMaterialHandler) DeleteLearningMaterial(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	materialID, err := uuid.Parse(idStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid material ID", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	err = h.Queries.DeleteLearningMaterial(r.Context(), db.DeleteLearningMaterialParams{
		MaterialID: materialID,
		SchoolID:   schoolID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not delete learning material", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}



