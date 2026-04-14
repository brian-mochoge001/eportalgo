package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/brian-mochoge001/eportalgo/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UserHandler struct {
	Queries     *db.Queries
	UserService *services.UserService
}

func NewUserHandler(q *db.Queries, s *services.UserService) *UserHandler {
	return &UserHandler{Queries: q, UserService: s}
}

func (h *UserHandler) AddUser(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		RoleName  string `json:"roleName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	newUser, err := h.UserService.AddUser(r.Context(), services.AddUserParams{
		SchoolID:  schoolID,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		RoleName:  req.RoleName,
	})
	if err != nil {
		middleware.InternalError(w, err.Error(), err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User added successfully",
		"user":    newUser,
	})
}

func (h *UserHandler) GetUsersBySchool(w http.ResponseWriter, r *http.Request) {
	schoolIDStr := chi.URLParam(r, "schoolId")
	schoolID, err := uuid.Parse(schoolIDStr)
	if err != nil {
		middleware.ValidationError(w, "Invalid school ID", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())

	// Authorization check
	isSuperAdmin := false
	for _, role := range []string{"Developer", "DB Manager", "Data Analyst", "Support Staff"} {
		if userCtx.RoleName == role {
			isSuperAdmin = true
			break
		}
	}
	if !isSuperAdmin && userCtx.SchoolID.UUID != schoolID {
		middleware.ForbiddenError(w, "Not authorized to view users for this school", nil)
		return
	}

	query := r.URL.Query().Get("query")

	users, err := h.Queries.ListUsersBySchool(r.Context(), db.ListUsersBySchoolParams{
		SchoolID: uuid.NullUUID{UUID: schoolID, Valid: true},
		Query:    sql.NullString{String: query, Valid: query != ""},
	})
	if err != nil {
		middleware.InternalError(w, "Could not fetch users", err)
		return
	}

	json.NewEncoder(w).Encode(users)
}

func (h *UserHandler) AddStudentProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, _ := uuid.Parse(userIDStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		EnrollmentNumber  string `json:"enrollment_number"`
		CurrentGradeLevel string `json:"current_grade_level"`
		AdmissionDate     string `json:"admission_date"`
		CurrentClassID    string `json:"current_class_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	admissionDate, _ := time.Parse("2006-01-02", req.AdmissionDate)
	currentClassID := toNullUUID(req.CurrentClassID)

	profile, err := h.UserService.CreateStudentProfile(r.Context(), services.CreateStudentProfileParams{
		UserID:            userID,
		SchoolID:          schoolID,
		EnrollmentNumber:  req.EnrollmentNumber,
		CurrentGradeLevel: req.CurrentGradeLevel,
		AdmissionDate:     admissionDate,
		CurrentClassID:    currentClassID,
	})
	if err != nil {
		middleware.InternalError(w, err.Error(), err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(profile)
}

func (h *UserHandler) AddParentProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, _ := uuid.Parse(userIDStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		HomeAddress           string `json:"home_address"`
		Occupation            string `json:"occupation"`
		EmergencyContactName  string `json:"emergency_contact_name"`
		EmergencyContactPhone string `json:"emergency_contact_phone"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	profile, err := h.UserService.CreateParentProfile(r.Context(), services.CreateParentProfileParams{
		UserID:                userID,
		SchoolID:              schoolID,
		HomeAddress:           req.HomeAddress,
		Occupation:            req.Occupation,
		EmergencyContactName:  req.EmergencyContactName,
		EmergencyContactPhone: req.EmergencyContactPhone,
	})
	if err != nil {
		middleware.InternalError(w, err.Error(), err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(profile)
}

func (h *UserHandler) GetFullProfile(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())

	if userCtx.RoleName == "Student" {
		profile, err := h.Queries.GetStudentFullProfile(r.Context(), userCtx.UserID)
		if err != nil {
			middleware.InternalError(w, "Could not fetch student profile", err)
			return
		}
		json.NewEncoder(w).Encode(profile)
	} else if userCtx.RoleName == "Parent" {
		profile, err := h.Queries.GetParentFullProfile(r.Context(), userCtx.UserID)
		if err != nil {
			middleware.InternalError(w, "Could not fetch parent profile", err)
			return
		}
		json.NewEncoder(w).Encode(profile)
	} else {
		// Default to base user info if no special profile
		user, err := h.Queries.GetUserByID(r.Context(), userCtx.UserID)
		if err != nil {
			middleware.InternalError(w, "Could not fetch user", err)
			return
		}
		json.NewEncoder(w).Encode(user)
	}
}

func (h *UserHandler) GetDetailedGrades(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	studentID := userCtx.UserID

	// If parent, they can specify studentId in query
	if userCtx.RoleName == "Parent" {
		sidStr := r.URL.Query().Get("studentId")
		if sidStr != "" {
			sid, _ := uuid.Parse(sidStr)
			studentID = sid
		}
	}

	grades, err := h.Queries.GetDetailedGrades(r.Context(), db.GetDetailedGradesParams{
		StudentID: studentID,
		SchoolID:  userCtx.SchoolID.UUID,
	})
	if err != nil {
		middleware.InternalError(w, "Could not fetch detailed grades", err)
		return
	}
	json.NewEncoder(w).Encode(grades)
}
