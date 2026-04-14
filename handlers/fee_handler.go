package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/google/uuid"
)

type FeeHandler struct {
	Queries *db.Queries
}

func NewFeeHandler(q *db.Queries) *FeeHandler {
	return &FeeHandler{Queries: q}
}

func (h *FeeHandler) GetFeeStructures(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	fees, err := h.Queries.GetFeeStructuresBySchool(r.Context(), schoolID)
	if err != nil {
		middleware.InternalError(w, "Could not fetch fee structures", err)
		return
	}

	json.NewEncoder(w).Encode(fees)
}

func (h *FeeHandler) CreateFeeStructure(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		FeeName      string `json:"fee_name"`
		Amount       string `json:"amount"`
		Currency     string `json:"currency"`
		AcademicYear string `json:"academic_year"`
		Description  string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	fee, err := h.Queries.CreateFeeStructure(r.Context(), db.CreateFeeStructureParams{
		SchoolID:     schoolID,
		FeeName:      req.FeeName,
		Amount:       req.Amount,
		Currency:     req.Currency,
		AcademicYear: req.AcademicYear,
		Description:  toNullString(req.Description),
		IsActive:     true,
	})

	if err != nil {
		middleware.InternalError(w, "Could not create fee structure", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(fee)
}

func (h *FeeHandler) GetStudentFees(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	studentIDStr := r.URL.Query().Get("studentId")
	var studentID uuid.UUID
	if studentIDStr != "" {
		studentID, _ = uuid.Parse(studentIDStr)
	} else if userCtx.RoleName == "Student" {
		studentID = userCtx.UserID
	}

	var studentFees interface{}
	var err error

	if studentID != uuid.Nil {
		// Authorization check: Students can only see their own fees.
		// Parents check would need more logic to verify relationship.
		if userCtx.RoleName == "Student" && userCtx.UserID != studentID {
			middleware.ForbiddenError(w, "You can only view your own fees", nil)
			return
		}
		studentFees, err = h.Queries.GetStudentFeesByStudent(r.Context(), db.GetStudentFeesByStudentParams{
			StudentID: studentID,
			SchoolID:  schoolID,
		})
	} else {
		// If no studentId provided and not a student, fetch all for school (Admin view)
		if !middleware.IsAdmin(userCtx.RoleName) {
			middleware.ForbiddenError(w, "Insufficient permissions", nil)
			return
		}
		studentFees, err = h.Queries.GetStudentFeesBySchool(r.Context(), schoolID)
	}

	if err != nil {
		middleware.InternalError(w, "Could not fetch student fees", err)
		return
	}

	json.NewEncoder(w).Encode(studentFees)
}

func (h *FeeHandler) CreateStudentFee(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		StudentID      string `json:"student_id"`
		FeeStructureID string `json:"fee_structure_id"`
		AmountDue      string `json:"amount_due"`
		DueDate        string `json:"due_date"`
		Notes          string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	studentID, _ := uuid.Parse(req.StudentID)
	feeStructureID, _ := uuid.Parse(req.FeeStructureID)
	dueDate, _ := parseDate(req.DueDate)

	studentFee, err := h.Queries.CreateStudentFee(r.Context(), db.CreateStudentFeeParams{
		SchoolID:       schoolID,
		StudentID:      studentID,
		FeeStructureID: feeStructureID,
		AmountDue:      req.AmountDue,
		DueDate:        dueDate,
		Notes:          toNullString(req.Notes),
	})

	if err != nil {
		middleware.InternalError(w, "Could not create student fee", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(studentFee)
}



