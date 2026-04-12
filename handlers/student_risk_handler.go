package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/brian-mochoge001/eportalgo/db"
	"github.com/brian-mochoge001/eportalgo/middleware"
	"github.com/brian-mochoge001/eportalgo/worker"
	"github.com/hibiken/asynq"
)

type StudentRiskHandler struct {
	Queries *db.Queries
	Asynq   *asynq.Client
}

func NewStudentRiskHandler(q *db.Queries, asynqClient *asynq.Client) *StudentRiskHandler {
	return &StudentRiskHandler{Queries: q, Asynq: asynqClient}
}

func (h *StudentRiskHandler) ListAtRiskStudents(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	riskLevelStr := r.URL.Query().Get("riskLevel")
	var riskLevel db.NullRiskLevel
	if riskLevelStr != "" {
		riskLevel = db.NullRiskLevel{RiskLevel: db.RiskLevel(riskLevelStr), Valid: true}
	}

	students, err := h.Queries.ListAtRiskStudents(r.Context(), db.ListAtRiskStudentsParams{
		SchoolID:  schoolID,
		RiskLevel: riskLevel,
	})

	if err != nil {
		middleware.InternalError(w, "Could not fetch at-risk students", err)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"results": len(students),
		"data": map[string]interface{}{"students": students},
	})
}

func (h *StudentRiskHandler) TriggerRiskCalculation(w http.ResponseWriter, r *http.Request) {
	userCtx, _ := middleware.GetUser(r.Context())
	schoolID := userCtx.SchoolID.UUID

	payload, _ := json.Marshal(worker.CalculateRiskScoresPayload{
		SchoolID: schoolID,
	})
	task := asynq.NewTask(worker.TypeCalculateRiskScores, payload)
	
	if _, err := h.Asynq.Enqueue(task); err != nil {
		middleware.InternalError(w, "Could not enqueue calculation task", err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"message": "Risk calculation triggered",
	})
}
