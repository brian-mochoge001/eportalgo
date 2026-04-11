package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewPaymentHandler(q *db.Queries, d *sql.DB) *PaymentHandler {
	return &PaymentHandler{Queries: q, DB: d}
}

func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	var req struct {
		StudentFeeID  string `json:"student_fee_id"`
		Amount        string `json:"amount"`
		PaymentMethod string `json:"payment_method"`
		TransactionID string `json:"transaction_id"`
		Notes         string `json:"notes"`
		ReceiptNumber string `json:"receipt_number"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.ValidationError(w, "Invalid request body", err)
		return
	}

	studentFeeID, _ := uuid.Parse(req.StudentFeeID)

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		middleware.InternalError(w, "Could not start transaction", err)
		return
	}
	defer tx.Rollback()

	qtx := h.Queries.WithTx(tx)

	// Create Payment
	payment, err := qtx.CreatePayment(r.Context(), db.CreatePaymentParams{
		SchoolID:          schoolID,
		StudentFeeID:      studentFeeID,
		Amount:            req.Amount,
		PaymentMethod:     toNullString(req.PaymentMethod),
		TransactionID:     toNullString(req.TransactionID),
		RecordedByUserID:  uuid.NullUUID{UUID: userCtx.UserID, Valid: true},
		Notes:             toNullString(req.Notes),
		ReceiptNumber:     toNullString(req.ReceiptNumber),
	})
	if err != nil {
		middleware.InternalError(w, "Could not create payment", err)
		return
	}

	// Update Student Fee Amount Paid
	_, err = qtx.UpdateStudentFeeAmountPaid(r.Context(), db.UpdateStudentFeeAmountPaidParams{
		StudentFeeID: studentFeeID,
		Column2:    req.Amount,
	})
	if err != nil {
		middleware.InternalError(w, "Could not update fee balance", err)
		return
	}

	if err := tx.Commit(); err != nil {
		middleware.InternalError(w, "Could not commit transaction", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(payment)
}

func (h *PaymentHandler) GetPayments(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	payments, err := h.Queries.ListPayments(r.Context(), schoolID)
	if err != nil {
		middleware.InternalError(w, "Could not fetch payments", err)
		return
	}

	json.NewEncoder(w).Encode(payments)
}



