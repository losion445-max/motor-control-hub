package network

import (
	"log"
	"net/http"
	"time"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

func (h *MotorHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	motorStatuses, err := h.orchestrator.GetAllAggregatedStatus(r.Context())
	if err != nil {
		log.Printf("[DIAG] Status error: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to fetch status")
		return
	}

	respondJSON(w, http.StatusOK, struct {
		Timestamp    int64                 `json:"timestamp"`
		Position     domain.Point2D        `json:"position"`
		Motors       []*domain.MotorStatus `json:"motors"`
		IsCalibrated bool                  `json:"is_calibrated"`
	}{
		Timestamp:    time.Now().UnixMilli(),
		Position:     h.orchestrator.GetCurrentPosition(),
		Motors:       motorStatuses,
		IsCalibrated: h.orchestrator.IsCalibrated(),
	})
}

func (h *MotorHandler) handleMoveSingleMotor(w http.ResponseWriter, r *http.Request) {
	var req struct {
		MotorID int     `json:"motor_id"`
		Steps   int     `json:"steps"`
		Speed   float64 `json:"speed"`
	}

	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if req.MotorID < 1 || req.MotorID > 4 {
		respondError(w, http.StatusBadRequest, "Invalid motor_id (expected 1-4)")
		return
	}

	if err := h.orchestrator.MoveSingleMotor(r.Context(), req.MotorID, req.Steps, req.Speed); err != nil {
		log.Printf("[DIAG] MoveSingleMotor (ID:%d) error: %v", req.MotorID, err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *MotorHandler) handleEnableMotors(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if err := h.orchestrator.SetMotorsEnabled(r.Context(), req.Enabled); err != nil {
		log.Printf("[DIAG] EnableMotors error: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to toggle motors")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":         "ok",
		"motors_enabled": req.Enabled,
	})
}

func (h *MotorHandler) handleSyncPosition(w http.ResponseWriter, r *http.Request) {
	var req struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}

	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	h.orchestrator.ManualSyncPosition(req.X, req.Y)
	log.Printf("[DIAG] Position manually synced to: X:%.2f, Y:%.2f", req.X, req.Y)

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "position_updated",
	})
}
