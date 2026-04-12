package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/brian-mochoge001/eportalgo/services"
	"github.com/google/uuid"
)

type PaymentHandler struct {
	Queries        *db.Queries
	FinanceService *services.FinanceService
}

func NewPaymentHandler(q *db.Queries, s *services.FinanceService) *PaymentHandler {
	return &PaymentHandler{Queries: q, FinanceService: s}
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

	payment, err := h.FinanceService.ProcessPayment(r.Context(), services.ProcessPaymentParams{
		SchoolID:         schoolID,
		StudentFeeID:     studentFeeID,
		Amount:           req.Amount,
		PaymentMethod:    req.PaymentMethod,
		TransactionID:    req.TransactionID,
		RecordedByUserID: userCtx.UserID,
		Notes:            req.Notes,
		ReceiptNumber:    req.ReceiptNumber,
	})

	if err != nil {
		middleware.InternalError(w, "Could not process payment", err)
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



