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
		log.Printf("[HTTP] Status error: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to fetch status")
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
