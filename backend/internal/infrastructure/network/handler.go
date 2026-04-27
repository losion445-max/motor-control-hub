package network

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/losion445-max/motor-control-hub/internal/domain"
	"github.com/losion445-max/motor-control-hub/internal/usecase"
)

type MotorHandler struct {
	orchestrator *usecase.MotorOrchestrator
}

func NewMotorHandler(orc *usecase.MotorOrchestrator) *MotorHandler {
	return &MotorHandler{orchestrator: orc}
}

func (h *MotorHandler) MapRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/move", h.handleMove)
	mux.HandleFunc("GET /api/status", h.handleStatus)
	mux.HandleFunc("POST /api/calibrate", h.handleCalibrate)
}

func (h *MotorHandler) handleMove(w http.ResponseWriter, r *http.Request) {
	var req struct {
		X     float64 `json:"x"`
		Y     float64 `json:"y"`
		Speed float64 `json:"speed"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	if err := h.orchestrator.MoveToPoint(r.Context(), req.X, req.Y, req.Speed); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"result": "success"})
}

func (h *MotorHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	motorStatuses, err := h.orchestrator.GetAllAggregatedStatus(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch motor statuses: %v", err), http.StatusInternalServerError)
		return
	}

	currentPos := h.orchestrator.GetCurrentPosition()

	response := struct {
		Timestamp    int64                 `json:"timestamp"`
		Position     domain.Point          `json:"position"`
		Motors       []*domain.MotorStatus `json:"motors"`
		IsCalibrated bool                  `json:"is_calibrated"`
	}{
		Timestamp:    time.Now().UnixMilli(),
		Position:     currentPos,
		Motors:       motorStatuses,
		IsCalibrated: h.orchestrator.IsCalibrated(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("JSON Encode Error: %v", err)
	}
}

func (h *MotorHandler) handleCalibrate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Speed float64 `json:"speed"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body. Expected format: {\"speed\": float.value}", http.StatusBadRequest)
		return
	}

	if req.Speed <= 0 {
		http.Error(w, "Calibration speed must be greater than 0", http.StatusBadRequest)
		return
	}
	if err := h.orchestrator.Calibrate(r.Context(), req.Speed); err != nil {
		http.Error(w, fmt.Sprintf("Calibration failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status":  "success",
		"message": "System successfully calibrated at physical position (0,0)",
		"speed":   req.Speed,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[HTTP] Error encoding response: %v", err)
	}
}

func (h *MotorHandler) handleEmergencyStop(w http.ResponseWriter, r *http.Request) {
	if err := h.orchestrator.EmergencyStop(r.Context()); err != nil {
		http.Error(w, fmt.Sprintf("Emergency stop failed: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{
		"status":  "success",
		"message": "Emergency stop executed successfully. All motors halted.",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("[HTTP] Error encoding response: %v", err)
	}
}
