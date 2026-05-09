package network

import (
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
	mux.HandleFunc("GET /api/config", h.handleConfig)
	mux.HandleFunc("GET /api/status", h.handleStatus)
	mux.HandleFunc("POST /api/move", h.handleMove)
	mux.HandleFunc("POST /api/calibrate", h.handleCalibrate)
	mux.HandleFunc("POST /api/stop", h.handleEmergencyStop)
	mux.HandleFunc("POST /api/home", h.handleGoHome)
	mux.HandleFunc("POST /api/dimensions", h.handleUpdateDimensions)
}

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
		respondError(w, http.StatusBadRequest, "Speed must be positive with move command!")
		return
	}

	if err := h.orchestrator.MoveToPoint(r.Context(), req.X, req.Y, req.Speed); err != nil {
		log.Printf("[HTTP] Move error: %v", err)
		respondError(w, http.StatusInternalServerError, "Movement command failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *MotorHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	motorStatuses, err := h.orchestrator.GetAllAggregatedStatus(r.Context())
	if err != nil {
		log.Printf("[HTTP] Status error: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to fetch motor statuses")
		return
	}

	respondJSON(w, http.StatusOK, struct {
		Timestamp    int64                 `json:"timestamp"`
		Position     domain.Point          `json:"position"`
		Motors       []*domain.MotorStatus `json:"motors"`
		IsCalibrated bool                  `json:"is_calibrated"`
	}{
		Timestamp:    time.Now().UnixMilli(),
		Position:     h.orchestrator.GetCurrentPosition(),
		Motors:       motorStatuses,
		IsCalibrated: h.orchestrator.IsCalibrated(),
	})
}

func (h *MotorHandler) handleCalibrate(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Speed float64 `json:"speed"`
	}

	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if req.Speed <= 0 {
		respondError(w, http.StatusBadRequest, "Calibration speed must be positive")
		return
	}

	if err := h.orchestrator.Calibrate(r.Context(), req.Speed); err != nil {
		log.Printf("[HTTP] Calibration error: %v", err)
		respondError(w, http.StatusInternalServerError, "Calibration failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"message": "System calibrated at (0,0)",
		"speed":   req.Speed,
	})
}

func (h *MotorHandler) handleEmergencyStop(w http.ResponseWriter, r *http.Request) {
	if err := h.orchestrator.EmergencyStop(r.Context()); err != nil {
		log.Printf("[CRITICAL] Emergency stop error: %v", err)
		respondError(w, http.StatusInternalServerError, "Emergency stop failed to propagate")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "All motors stopped",
	})
}

func (h *MotorHandler) handleConfig(w http.ResponseWriter, r *http.Request) {
	motorConfigs, err := h.orchestrator.GetAllAggregatedConfig(r.Context())
	if err != nil {
		log.Printf("[HTTP] Config error: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to fetch hardware config")
		return
	}

	respondJSON(w, http.StatusOK, struct {
		FrameWidth  float64              `json:"frame_width"`
		FrameHeight float64              `json:"frame_height"`
		Motors      []domain.MotorConfig `json:"motors"`
	}{
		FrameWidth:  h.orchestrator.GetFrameWidth(),
		FrameHeight: h.orchestrator.GetFrameHeight(),
		Motors:      motorConfigs,
	})
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

	if err := h.orchestrator.GoHome(r.Context(), req.Speed); err != nil {
		log.Printf("[HTTP] GoHome error: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to move home")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "moving_home"})
}

func (h *MotorHandler) handleUpdateDimensions(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
	}

	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid JSON body")
		return
	}

	if req.Width <= 0 || req.Height <= 0 {
		respondError(w, http.StatusBadRequest, "Width and Height must be positive")
		return
	}

	h.orchestrator.SetDimensions(req.Width, req.Height)
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"dimensions": map[string]float64{
			"width":  req.Width,
			"height": req.Height,
		},
	})
}
