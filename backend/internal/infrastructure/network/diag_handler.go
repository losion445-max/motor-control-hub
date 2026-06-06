package network

import (
	"log"
	"net/http"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

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
	if err := h.moveSingleMotor.Execute(r.Context(), req.MotorID, req.Steps, req.Speed); err != nil {
		log.Printf("[DIAG] MoveSingleMotor (ID:%d) error: %v", req.MotorID, err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *MotorHandler) handleSetEnabled(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	if err := h.setEnabled.Execute(r.Context(), req.Enabled); err != nil {
		log.Printf("[DIAG] SetEnabled error: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to toggle motors")
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":         "ok",
		"motors_enabled": req.Enabled,
	})
}

func (h *MotorHandler) handleSetHome(w http.ResponseWriter, r *http.Request) {
	var req struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}
	h.setHome.Execute(domain.Point2D{X: req.X, Y: req.Y})
	log.Printf("[DIAG] Home set to: X:%.2f, Y:%.2f", req.X, req.Y)
	respondJSON(w, http.StatusOK, map[string]string{"status": "position_updated"})
}
