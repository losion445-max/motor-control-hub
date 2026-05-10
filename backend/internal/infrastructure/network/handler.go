package network

import (
	"net/http"

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
	mux.HandleFunc("POST /api/calibrate", h.handleCalibrate)
	mux.HandleFunc("POST /api/home", h.handleGoHome)
	mux.HandleFunc("POST /api/stop", h.handleEmergencyStop)

	mux.HandleFunc("GET /api/config", h.handleGetConfig)
	mux.HandleFunc("POST /api/config", h.handleUpdateConfig)

	mux.HandleFunc("GET /api/status", h.handleStatus)
	mux.HandleFunc("POST /api/motors/enable", h.handleEnableMotors)
	mux.HandleFunc("POST /api/motors/move-single", h.handleMoveSingleMotor)
	mux.HandleFunc("POST /api/position/sync", h.handleSyncPosition)
}
