package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type UserHandler struct {
	Queries *db.Queries
	DB      *sql.DB
	FirebaseApp *firebase.App
}

func NewUserHandler(q *db.Queries, d *sql.DB, fbApp *firebase.App) *UserHandler {
	return &UserHandler{Queries: q, DB: d, FirebaseApp: fbApp}
}

func (h *UserHandler) AddUser(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		RoleName  string `json:"roleName"`
		Password  string `json:"password"` // Password for Firebase Auth
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.FirstName == "" || req.LastName == "" || req.RoleName == "" || req.Password == "" {
		middleware.SendError(w, "Please provide email, first name, last name, role name, and password.", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		middleware.SendError(w, "Could not start transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()
	qtx := h.Queries.WithTx(tx)

	// Find the role_id based on roleName
	role, err := qtx.GetRoleByName(r.Context(), req.RoleName)
	if err != nil {
		middleware.SendError(w, "Invalid role specified: "+req.RoleName, http.StatusBadRequest)
		return
	}

	// Ensure the role is a school-specific role
	if !role.IsSchoolRole {
		middleware.SendError(w, "Cannot add parent company role ("+req.RoleName+") via this endpoint.", http.StatusBadRequest)
		return
	}

	// Check if user already exists in our DB
	existingUser, err := qtx.GetUserByEmail(r.Context(), db.GetUserByEmailParams{
		Email:    req.Email,
		SchoolID: uuid.NullUUID{UUID: schoolID, Valid: true},
	})
	if err == nil && existingUser.UserID != uuid.Nil {
		middleware.SendError(w, "User with this email already exists in the database.", http.StatusConflict)
		return
	}

	// Check if email already exists in Firebase Auth
	authClient, _ := h.FirebaseApp.Auth(r.Context())
	firebaseUser, err := authClient.GetUserByEmail(r.Context(), req.Email)
	if err == nil && firebaseUser != nil {
		middleware.SendError(w, "Email already in use in Firebase Authentication.", http.StatusConflict)
		return
	}

	// Create user in Firebase Authentication
	userParams := (&auth.UserToCreate{}).Email(req.Email).Password(req.Password).DisplayName(req.FirstName + " " + req.LastName)
	userRecord, err := authClient.CreateUser(r.Context(), userParams)
	if err != nil {
		middleware.SendError(w, "Failed to create user in Firebase: "+err.Error(), http.StatusInternalServerError)
		return
	}
	firebaseUID := userRecord.UID

	// Create user in our database
	newUser, err := qtx.CreateUser(r.Context(), db.CreateUserParams{
		SchoolID: uuid.NullUUID{UUID: schoolID, Valid: true},
		RoleID:   role.RoleID,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		FirebaseUid: sql.NullString{String: firebaseUID, Valid: true},
		IsActive:  true,
	})
	if err != nil {
		middleware.SendError(w, "Failed to create user in database", http.StatusInternalServerError)
		return
	}

	// Set custom claims for the new user in Firebase
	claims := map[string]interface{}{
		"role": role.RoleName,
		"schoolId": schoolID.String(),
		"schoolStatus": "pending", // Default status
	}
	if err := authClient.SetCustomUserClaims(r.Context(), firebaseUID, claims); err != nil {
		middleware.SendError(w, "Failed to set custom claims in Firebase", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		middleware.SendError(w, "Could not commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User added successfully",
		"user": map[string]interface{}{
			"id": newUser.UserID,
			"email": newUser.Email,
			"firstName": newUser.FirstName,
			"lastName": newUser.LastName,
			"role": role.RoleName,
			"schoolId": schoolID.String(),
			"firebase_uid": firebaseUID,
			"isActive": newUser.IsActive,
		},
	})
}

func (h *UserHandler) GetUsersBySchool(w http.ResponseWriter, r *http.Request) {
	schoolIDStr := chi.URLParam(r, "schoolId")
	schoolID, err := uuid.Parse(schoolIDStr)
	if err != nil {
		middleware.SendError(w, "Invalid school ID", http.StatusBadRequest)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())

	// Authorization check: user must belong to the requested school or be a super admin
	isSuperAdmin := false
	for _, role := range []string{"Developer", "DB Manager", "Data Analyst", "Support Staff"} {
		if userCtx.RoleName == role {
			isSuperAdmin = true
			break
		}
	}
	if !isSuperAdmin && userCtx.SchoolID.UUID != schoolID {
		middleware.SendError(w, "Not authorized to view users for this school", http.StatusForbidden)
		return
	}

	users, err := h.Queries.ListUsersBySchool(r.Context(), uuid.NullUUID{UUID: schoolID, Valid: true})
	if err != nil {
		middleware.SendError(w, "Could not fetch users", http.StatusInternalServerError)
		return
	}

	// Map to desired output format
	mappedUsers := make([]map[string]interface{}, len(users))
	for i, user := range users {
		mappedUsers[i] = map[string]interface{}{
			"id":           user.UserID,
			"email":        user.Email,
			"firstName":    user.FirstName,
			"lastName":     user.LastName,
			"role":         user.RoleName,
			"schoolId":     user.SchoolID,
			"firebase_uid": user.FirebaseUid.String,
			"isActive":     user.IsActive,
		}
	}

	json.NewEncoder(w).Encode(mappedUsers)
}

func (h *UserHandler) AddStudentProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, _ := uuid.Parse(userIDStr)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		EnrollmentNumber string `json:"enrollment_number"`
		CurrentGradeLevel string `json:"current_grade_level"`
		AdmissionDate    string `json:"admission_date"`
		CurrentClassID   string `json:"current_class_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	admissionDate, _ := time.Parse("2006-01-02", req.AdmissionDate)
	currentClassID := toNullUUID(req.CurrentClassID)

	// Check if the user exists and belongs to the same school
	user, err := h.Queries.GetUser(r.Context(), db.GetUserParams{
		UserID:   userID,
		SchoolID: uuid.NullUUID{UUID: schoolID, Valid: true},
	})
	if err != nil {
		middleware.SendError(w, "User not found", http.StatusNotFound)
		return
	}

	// Ensure the user has the 'Student' role
	studentRole, err := h.Queries.GetRoleByName(r.Context(), "Student")
	if err != nil || user.RoleID != studentRole.RoleID {
		middleware.SendError(w, "User must have the Student role", http.StatusBadRequest)
		return
	}

	// Check if a student profile already exists for this user
	if _, err := h.Queries.GetStudentProfileByUserID(r.Context(), db.GetStudentProfileByUserIDParams{
		UserID:   userID,
		SchoolID: schoolID,
	}); err == nil {
		middleware.SendError(w, "Student profile already exists for this user", http.StatusConflict)
		return
	}

	studentProfile, err := h.Queries.CreateStudentProfile(r.Context(), db.CreateStudentProfileParams{
		UserID:           userID,
		SchoolID:         schoolID,
		EnrollmentNumber: req.EnrollmentNumber,
		CurrentGradeLevel: toNullString(req.CurrentGradeLevel),
		AdmissionDate:    admissionDate,
		CurrentClassID:   currentClassID,
	})
	if err != nil {
		middleware.SendError(w, "Failed to create student profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(studentProfile)
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
		middleware.SendError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Check if the user exists and belongs to the same school
	user, err := h.Queries.GetUser(r.Context(), db.GetUserParams{
		UserID:   userID,
		SchoolID: uuid.NullUUID{UUID: schoolID, Valid: true},
	})
	if err != nil {
		middleware.SendError(w, "User not found", http.StatusNotFound)
		return
	}

	// Ensure the user has the 'Parent' role
	parentRole, err := h.Queries.GetRoleByName(r.Context(), "Parent")
	if err != nil || user.RoleID != parentRole.RoleID {
		middleware.SendError(w, "User must have the Parent role", http.StatusBadRequest)
		return
	}

	// Check if a parent profile already exists for this user
	if _, err := h.Queries.GetParentProfileByUserID(r.Context(), db.GetParentProfileByUserIDParams{
		UserID:   userID,
		SchoolID: schoolID,
	}); err == nil {
		middleware.SendError(w, "Parent profile already exists for this user", http.StatusConflict)
		return
	}

	parentProfile, err := h.Queries.CreateParentProfile(r.Context(), db.CreateParentProfileParams{
		UserID:                userID,
		SchoolID:              schoolID,
		HomeAddress:           toNullString(req.HomeAddress),
		Occupation:            toNullString(req.Occupation),
		EmergencyContactName:  toNullString(req.EmergencyContactName),
		EmergencyContactPhone: toNullString(req.EmergencyContactPhone),
	})
	if err != nil {
		middleware.SendError(w, "Failed to create parent profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(parentProfile)
}
