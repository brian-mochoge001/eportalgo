package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/brian-mochoge001/eportalgo/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type EnrollmentHandler struct {
	Queries        *db.Queries
	StudentService *services.StudentService
}

func NewEnrollmentHandler(q *db.Queries, s *services.StudentService) *EnrollmentHandler {
	return &EnrollmentHandler{Queries: q, StudentService: s}
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

	err := h.StudentService.OnboardStudent(r.Context(), services.OnboardStudentRequest{
		SchoolID:          schoolID,
		StudentFirstName:  req.StudentFirstName,
		StudentLastName:   req.StudentLastName,
		StudentDob:        dob,
		StudentGender:     req.StudentGender,
		ParentFirstName:   req.ParentFirstName,
		ParentLastName:    req.ParentLastName,
		ParentEmail:       req.ParentEmail,
		ParentPhoneNumber: req.ParentPhoneNumber,
		ClassID:           classID,
		EnrollmentDate:    enrDate,
	})

	if err != nil {
		middleware.InternalError(w, "Failed to onboard student", err)
		return
	}

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

	transfer, err := h.StudentService.InitiateTransfer(r.Context(), sid, schoolID, dsid, userCtx.UserID)

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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	userCtx, _ := middleware.GetUser(r.Context())

	updated, err := h.StudentService.ProcessTransfer(r.Context(), id, req.Status, req.Notes, userCtx.SchoolID.UUID)
	if err != nil {
		middleware.InternalError(w, "Failed to process transfer", err)
		return
	}

	json.NewEncoder(w).Encode(updated)
}



