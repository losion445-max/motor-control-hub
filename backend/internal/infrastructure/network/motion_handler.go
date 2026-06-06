package network

import (
	"log"
	"net/http"

	"github.com/losion445-max/motor-control-hub/internal/domain"
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
	if err := h.moveTo.Execute(r.Context(), domain.Point2D{X: req.X, Y: req.Y}, req.Speed); err != nil {
		log.Printf("[HTTP] Move error: %v", err)
		respondError(w, http.StatusInternalServerError, "Movement command failed")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *MotorHandler) handleGoHome(w http.ResponseWriter, r *http.Request) {
	var req struct {
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
	if err := h.moveTo.Execute(r.Context(), domain.Point2D{X: 0, Y: 0}, req.Speed); err != nil {
		log.Printf("[HTTP] GoHome error: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to move home")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "moving_home"})
}

func (h *MotorHandler) handleStop(w http.ResponseWriter, r *http.Request) {
	if err := h.stopAll.Execute(r.Context()); err != nil {
		log.Printf("[HTTP] Stop error: %v", err)
		respondError(w, http.StatusInternalServerError, "Emergency stop failed")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "stopped"})
}
