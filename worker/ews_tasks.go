package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	TypeCalculateRiskScores = "ews:calculate_risk"
)

type CalculateRiskScoresPayload struct {
	SchoolID uuid.UUID
}

func (h *TaskHandler) HandleCalculateRiskScores(ctx context.Context, t *asynq.Task) error {
	var p CalculateRiskScoresPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("Calculating risk scores for school %s", p.SchoolID)

	metrics, err := h.Queries.CalculateStudentMetrics(ctx, p.SchoolID)
	if err != nil {
		return err
	}

	for _, m := range metrics {
		// Calculate risk score (0-100)
		// Attendance weight: 60%, Grade weight: 40%
		
		attRate, _ := strconv.ParseFloat(m.AttendanceRate, 64)
		avgGrade, _ := strconv.ParseFloat(m.AverageGrade, 64)

		// Attendance component: 100% attendance = 0 risk, 0% attendance = 60 risk
		attendanceRisk := (100.0 - attRate) * 0.6
		
		// Grade component: 100 score = 0 risk, 0 score = 40 risk (assuming 100 is max)
		gradeRisk := (100.0 - avgGrade) * 0.4
		if gradeRisk < 0 { gradeRisk = 0 }

		totalRisk := int(attendanceRisk + gradeRisk)
		if totalRisk > 100 { totalRisk = 100 }

		level := db.RiskLevelLow
		if totalRisk >= 70 {
			level = db.RiskLevelHigh
		} else if totalRisk >= 40 {
			level = db.RiskLevelMedium
		}

		_, err := h.Queries.UpsertStudentRiskScore(ctx, db.UpsertStudentRiskScoreParams{
			SchoolID:       p.SchoolID,
			StudentID:      m.UserID,
			AttendanceRate: m.AttendanceRate,
			AverageGrade:   m.AverageGrade,
			RiskScore:      int32(totalRisk),
			RiskLevel:      level,
		})
		if err != nil {
			log.Printf("Failed to upsert risk score for student %s: %v", m.UserID, err)
		}
	}

	return nil
}
