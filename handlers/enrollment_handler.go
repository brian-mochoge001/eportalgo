package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type EnrollmentHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewEnrollmentHandler(q *db.Queries, d *sql.DB) *EnrollmentHandler {
	return &EnrollmentHandler{Queries: q, DB: d}
}

func (h *EnrollmentHandler) GetEnrollments(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	enrollments, err := h.Queries.GetEnrollmentsBySchool(r.Context(), schoolID)
	if err != nil {
		middleware.InternalError(w, "Could not fetch enrollments", err)
		return
	}

	json.NewEncoder(w).Encode(enrollments)
}

func (h *EnrollmentHandler) OnboardNewStudent(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		StudentFirstName string `json:"studentFirstName"`
		StudentLastName  string `json:"studentLastName"`
		StudentDob       string `json:"studentDob"`
		StudentGender    string `json:"studentGender"`
		ParentFirstName  string `json:"parentFirstName"`
		ParentLastName   string `json:"parentLastName"`
		ParentEmail      string `json:"parentEmail"`
		ParentPhoneNumber string `json:"parentPhoneNumber"`
		ClassId          string `json:"classId"`
		EnrollmentDate   string `json:"enrollmentDate"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	classID, _ := uuid.Parse(req.ClassId)
	dob, _ := time.Parse("2006-01-02", req.StudentDob)
	enrDate, _ := time.Parse("2006-01-02", req.EnrollmentDate)

	tx, err := h.DB.Begin()
	if err != nil {
		middleware.InternalError(w, "Internal Server Error", err)
		return
	}
	defer tx.Rollback()

	qtx := h.Queries.WithTx(tx)

	// Roles
	studentRole, _ := qtx.GetRoleByName(r.Context(), "Student")
	parentRole, _ := qtx.GetRoleByName(r.Context(), "Parent")

	// Create Student
	studentEmail := fmt.Sprintf("%s.%s@student.edu", strings.ToLower(req.StudentFirstName), strings.ToLower(req.StudentLastName))
	pass, _ := bcrypt.GenerateFromPassword([]byte("password123"), 10)

	studentUser, err := qtx.CreateUser(r.Context(), db.CreateUserParams{
		SchoolID:     uuid.NullUUID{UUID: schoolID, Valid: true},
		RoleID:       studentRole.RoleID,
		FirstName:    req.StudentFirstName,
		LastName:     req.StudentLastName,
		Email:        studentEmail,
		PasswordHash: sql.NullString{String: string(pass), Valid: true},
		DateOfBirth:  sql.NullTime{Time: dob, Valid: true},
		Gender:       sql.NullString{String: req.StudentGender, Valid: true},
		IsActive:     true,
	})
	if err != nil {
		middleware.InternalError(w, "Failed to create student user", err)
		return
	}

	qtx.CreateStudentProfile(r.Context(), db.CreateStudentProfileParams{
		UserID:           studentUser.UserID,
		SchoolID:         schoolID,
		EnrollmentNumber: fmt.Sprintf("ENR-%d", time.Now().Unix()),
		AdmissionDate:    enrDate,
		CurrentClassID:   uuid.NullUUID{UUID: classID, Valid: true},
	})

	// Parent
	parentUser, err := qtx.GetUserByEmail(r.Context(), db.GetUserByEmailParams{
		Email:    req.ParentEmail,
		SchoolID: uuid.NullUUID{UUID: schoolID, Valid: true},
	})
	if err != nil { // If parent user does not exist, create one
		if err == sql.ErrNoRows {
			parentUser, err = qtx.CreateUser(r.Context(), db.CreateUserParams{
				SchoolID:     uuid.NullUUID{UUID: schoolID, Valid: true},
				RoleID:       parentRole.RoleID,
				FirstName:    req.ParentFirstName,
				LastName:     req.ParentLastName,
				Email:        req.ParentEmail,
				PasswordHash: sql.NullString{String: string(pass), Valid: true},
				PhoneNumber:  sql.NullString{String: req.ParentPhoneNumber, Valid: true},
				IsActive:     true,
			})
			if err != nil {
				middleware.InternalError(w, "Failed to create parent user", err)
				return
			}
			_, err = qtx.CreateParentProfile(r.Context(), db.CreateParentProfileParams{
				UserID:   parentUser.UserID,
				SchoolID: schoolID,
			})
			if err != nil {
				middleware.InternalError(w, "Failed to create parent profile", err)
				return
			}
		} else {
			middleware.InternalError(w, "Failed to check parent user", err)
			return
		}
	}

	qtx.CreateEnrollment(r.Context(), db.CreateEnrollmentParams{
		SchoolID:       schoolID,
		StudentID:      studentUser.UserID,
		ClassID:        classID,
		EnrollmentDate: enrDate,
		Status:         "Enrolled",
	})

	tx.Commit()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "Student onboarded successfully"})
}

func (h *EnrollmentHandler) InitiateStudentTransfer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StudentID           string `json:"studentId"`
		DestinationSchoolID string `json:"destinationSchoolId"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	sid, _ := uuid.Parse(req.StudentID)
	dsid, _ := uuid.Parse(req.DestinationSchoolID)

	transfer, err := h.Queries.CreateTransferRequest(r.Context(), db.CreateTransferRequestParams{
		EntityType:          "Student",
		EntityID:            sid,
		SourceSchoolID:      schoolID,
		DestinationSchoolID: dsid,
		InitiatedByUserID:   userCtx.UserID,
	})

	if err != nil {
		middleware.InternalError(w, "Could not initiate transfer", err)
		return
	}

	json.NewEncoder(w).Encode(transfer)
}

func (h *EnrollmentHandler) ProcessIncomingTransfer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "transferRequestId")
	id, _ := uuid.Parse(idStr)

	var req struct {
		Status string `json:"status"` // approved or denied
		Notes  string `json:"notes"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	userCtx, _ := middleware.GetUser(r.Context())
	
	tx, _ := h.DB.Begin()
	defer tx.Rollback()
	qtx := h.Queries.WithTx(tx)

	tr, _ := qtx.GetTransferRequestByID(r.Context(), id)
	
	if req.Status == "approved" {
		// Update student school
		// qtx.UpdateUserSchool(r.Context(), ...) // Need this query if not there
		qtx.CreateEnrollment(r.Context(), db.CreateEnrollmentParams{
			SchoolID:       userCtx.SchoolID.UUID,
			StudentID:      tr.EntityID,
			EnrollmentDate: time.Now(),
			Status:         "Enrolled",
		})
	}

	updated, _ := qtx.UpdateTransferRequestStatus(r.Context(), db.UpdateTransferRequestStatusParams{
		TransferID:     id,
		Status:         req.Status,
		CompletionDate: sql.NullTime{Time: time.Now(), Valid: true},
		Notes:          sql.NullString{String: req.Notes, Valid: true},
	})

	tx.Commit()
	json.NewEncoder(w).Encode(updated)
}



