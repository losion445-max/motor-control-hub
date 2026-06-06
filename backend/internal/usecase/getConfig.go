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
	cfg      *config.GlobalConfig
	registry domain.IMotorRegistry
}

func NewGetConfig(cfg *config.GlobalConfig, registry domain.IMotorRegistry) *GetConfig {
	return &GetConfig{cfg: cfg, registry: registry}
}

func (uc *GetConfig) Execute(ctx context.Context) ConfigResult {
	motors := uc.registry.Motors()
	var (
		wg      sync.WaitGroup
		configs [4]domain.MotorConfig
	)
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			if motors[idx] == nil {
				return
			}
			cfg, err := motors[idx].GetConfig(ctx)
			if err != nil || cfg == nil {
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
