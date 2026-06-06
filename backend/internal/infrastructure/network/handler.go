package network

import (
	"net/http"

	"github.com/losion445-max/motor-control-hub/internal/usecase"
)

type MotorHandler struct {
	moveTo          *usecase.MoveTo
	setHome         *usecase.SetHome
	stopAll         *usecase.StopAll
	getStatus       *usecase.GetStatus
	setEnabled      *usecase.SetEnabled
	moveSingleMotor *usecase.MoveSingleMotor
	getConfig       *usecase.GetConfig
	updateConfig    *usecase.UpdateConfig
}

func NewMotorHandler(
	moveTo *usecase.MoveTo,
	setHome *usecase.SetHome,
	stopAll *usecase.StopAll,
	getStatus *usecase.GetStatus,
	setEnabled *usecase.SetEnabled,
	moveSingleMotor *usecase.MoveSingleMotor,
	getConfig *usecase.GetConfig,
	updateConfig *usecase.UpdateConfig,
) *MotorHandler {
	return &MotorHandler{
		moveTo:          moveTo,
		setHome:         setHome,
		stopAll:         stopAll,
		getStatus:       getStatus,
		setEnabled:      setEnabled,
		moveSingleMotor: moveSingleMotor,
		getConfig:       getConfig,
		updateConfig:    updateConfig,
	}
}

func (h *MotorHandler) MapRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/move", h.handleMove)
	mux.HandleFunc("POST /api/calibrate", h.handleSetHome)
	mux.HandleFunc("POST /api/home", h.handleGoHome)
	mux.HandleFunc("POST /api/stop", h.handleStop)
	mux.HandleFunc("GET /api/status", h.handleStatus)
	mux.HandleFunc("POST /api/motors/enable", h.handleSetEnabled)
	mux.HandleFunc("POST /api/motors/move-single", h.handleMoveSingleMotor)
	mux.HandleFunc("POST /api/position/sync", h.handleSetHome)
	mux.HandleFunc("GET /api/config", h.handleGetConfig)
	mux.HandleFunc("POST /api/config", h.handleUpdateConfig)
}
