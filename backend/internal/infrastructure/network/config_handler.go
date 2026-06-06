package network

import (
	"log"
	"net/http"

	"github.com/losion445-max/motor-control-hub/internal/usecase"
)

func (h *MotorHandler) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	result := h.getConfig.Execute(r.Context())
	respondJSON(w, http.StatusOK, result)
}

func (h *MotorHandler) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req usecase.UpdateConfigRequest
	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid config JSON")
		return
	}
	h.updateConfig.Execute(req)
	respondJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func (h *MotorHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	result, err := h.getStatus.Execute(r.Context())
	if err != nil {
		log.Printf("[DIAG] Status error: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to fetch status")
		return
	}
	respondJSON(w, http.StatusOK, result)
}
