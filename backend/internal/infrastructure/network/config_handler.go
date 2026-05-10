package network

import "net/http"

func (h *MotorHandler) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	cfg := h.orchestrator.GetConfig()
	motorConfigs, _ := h.orchestrator.GetAllAggregatedConfig(r.Context())

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"global":          cfg,
		"motors_hardware": motorConfigs,
	})
}

func (h *MotorHandler) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Width        *float64 `json:"width"`
		Height       *float64 `json:"height"`
		Diameter     *float64 `json:"diameter"`
		StepsPerRev  *int     `json:"steps_per_rev"`
		MotorMapping *[4]int  `json:"motor_mapping"`
	}

	if err := decode(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid config JSON")
		return
	}

	cfg := h.orchestrator.GetConfig()

	if req.Width != nil {
		cfg.Kinematics.Width = *req.Width
	}
	if req.Height != nil {
		cfg.Kinematics.Height = *req.Height
	}
	if req.Diameter != nil {
		cfg.Kinematics.Diameter = *req.Diameter
	}
	if req.StepsPerRev != nil {
		cfg.Kinematics.StepsPerRev = *req.StepsPerRev
	}
	if req.MotorMapping != nil {
		cfg.MotorMapping = *req.MotorMapping
	}

	h.orchestrator.Sync()

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":         "success",
		"applied_config": cfg,
	})
}
