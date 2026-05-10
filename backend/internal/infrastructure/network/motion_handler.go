package network

import (
	"log"
	"net/http"
)

func (h *MotorHandler) handleMove(w http.ResponseWriter, r *http.Request) {
	var req struct {
		X     float64 `json:"x"`
		Y     float64 `json:"y"`
		Speed float64 `json:"speed"`
	}

	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if req.Speed <= 0 {
		respondError(w, http.StatusBadRequest, "Speed must be positive")
		return
	}

	if err := h.orchestrator.MoveToPoint(r.Context(), req.X, req.Y, req.Speed); err != nil {
		log.Printf("[HTTP] Move error: %v", err)
		respondError(w, http.StatusInternalServerError, "Movement command failed")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *MotorHandler) handleCalibrate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Speed float64 `json:"speed"`
	}
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if err := h.orchestrator.Calibrate(r.Context(), req.Speed); err != nil {
		log.Printf("[HTTP] Calibration error: %v", err)
		respondError(w, http.StatusInternalServerError, "Calibration failed")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *MotorHandler) handleGoHome(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Speed float64 `json:"speed"`
	}
	decode(r, &req)

	if err := h.orchestrator.GoHome(r.Context(), req.Speed); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to move home")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "moving_home"})
}

func (h *MotorHandler) handleEmergencyStop(w http.ResponseWriter, r *http.Request) {
	if err := h.orchestrator.EmergencyStop(r.Context()); err != nil {
		respondError(w, http.StatusInternalServerError, "Emergency stop failed")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}
