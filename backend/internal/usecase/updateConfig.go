package usecase

import (
	"github.com/losion445-max/motor-control-hub/internal/config"
	"github.com/losion445-max/motor-control-hub/internal/domain"
)

type UpdateConfigRequest struct {
	Width       *float64 `json:"width"`
	Height      *float64 `json:"height"`
	Diameter    *float64 `json:"diameter"`
	StepsPerRev *int     `json:"steps_per_rev"`
}

type UpdateConfig struct {
	cfg        *config.GlobalConfig
	kinematics domain.IKinematicsController
}

func NewUpdateConfig(cfg *config.GlobalConfig, kinematics domain.IKinematicsController) *UpdateConfig {
	return &UpdateConfig{cfg: cfg, kinematics: kinematics}
}

func (uc *UpdateConfig) Execute(req UpdateConfigRequest) {
	if req.Width != nil {
		uc.cfg.Kinematics.Width = *req.Width
	}
	if req.Height != nil {
		uc.cfg.Kinematics.Height = *req.Height
	}
	if req.Diameter != nil {
		uc.cfg.Kinematics.Diameter = *req.Diameter
	}
	if req.StepsPerRev != nil {
		uc.cfg.Kinematics.StepsPerRev = *req.StepsPerRev
	}

	uc.kinematics.SetWorkZone(domain.WorkZone{
		Width:  uc.cfg.Kinematics.Width,
		Height: uc.cfg.Kinematics.Height,
	})
}
