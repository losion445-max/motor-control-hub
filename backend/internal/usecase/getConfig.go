package usecase

import (
	"context"
	"sync"

	"github.com/losion445-max/motor-control-hub/internal/config"
	"github.com/losion445-max/motor-control-hub/internal/domain"
)

type ConfigResult struct {
	Global *config.GlobalConfig  `json:"global"`
	Motors [4]domain.MotorConfig `json:"motors_hardware"`
}

type GetConfig struct {
	cfg    *config.GlobalConfig
	motors [4]domain.IMotor
}

func NewGetConfig(cfg *config.GlobalConfig, motors [4]domain.IMotor) *GetConfig {
	return &GetConfig{cfg: cfg, motors: motors}
}

func (uc *GetConfig) Execute(ctx context.Context) ConfigResult {
	var (
		wg      sync.WaitGroup
		configs [4]domain.MotorConfig
	)

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cfg, err := uc.motors[idx].GetConfig(ctx)
			if err != nil || cfg == nil {
				configs[idx] = domain.MotorConfig{}
				return
			}
			configs[idx] = *cfg
		}(i)
	}
	wg.Wait()

	return ConfigResult{
		Global: uc.cfg,
		Motors: configs,
	}
}
