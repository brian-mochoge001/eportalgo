package services

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
)

type FinanceService struct {
	Queries *db.Queries
	DB      *sql.DB
}

func NewFinanceService(q *db.Queries, d *sql.DB) *FinanceService {
	return &FinanceService{Queries: q, DB: d}
}

type ProcessPaymentParams struct {
	SchoolID         uuid.UUID
	StudentFeeID     uuid.UUID
	Amount           string
	PaymentMethod    string
	TransactionID    string
	RecordedByUserID uuid.UUID
	Notes            string
	ReceiptNumber    string
}

func (s *FinanceService) ProcessPayment(ctx context.Context, p ProcessPaymentParams) (db.Payment, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return db.Payment{}, err
	}
	defer tx.Rollback()

	qtx := s.Queries.WithTx(tx)

	// Create Payment
	payment, err := qtx.CreatePayment(ctx, db.CreatePaymentParams{
		SchoolID:          p.SchoolID,
		StudentFeeID:      p.StudentFeeID,
		Amount:            p.Amount,
		PaymentMethod:     sql.NullString{String: p.PaymentMethod, Valid: p.PaymentMethod != ""},
		TransactionID:     sql.NullString{String: p.TransactionID, Valid: p.TransactionID != ""},
		RecordedByUserID:  uuid.NullUUID{UUID: p.RecordedByUserID, Valid: true},
		Notes:             sql.NullString{String: p.Notes, Valid: p.Notes != ""},
		ReceiptNumber:     sql.NullString{String: p.ReceiptNumber, Valid: p.ReceiptNumber != ""},
	})
	if err != nil {
		return db.Payment{}, fmt.Errorf("could not create payment: %w", err)
	}

	// Update Student Fee Amount Paid
	_, err = qtx.UpdateStudentFeeAmountPaid(ctx, db.UpdateStudentFeeAmountPaidParams{
		StudentFeeID: p.StudentFeeID,
		Column2:      p.Amount,
	})
	if err != nil {
		return db.Payment{}, fmt.Errorf("could not update fee balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return db.Payment{}, err
	}

	return payment, nil
}
