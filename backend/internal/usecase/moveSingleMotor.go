package usecase

import (
	"context"
	"fmt"

	"github.com/losion445-max/motor-control-hub/internal/domain"
)

type MoveSingleMotor struct {
	registry domain.IMotorRegistry
}

func NewMoveSingleMotor(registry domain.IMotorRegistry) *MoveSingleMotor {
	return &MoveSingleMotor{registry: registry}
}

func (uc *MoveSingleMotor) Execute(ctx context.Context, motorID int, steps int, speedHz float64) error {
	idx := motorID - 1
	if idx < 0 || idx >= 4 {
		return fmt.Errorf("invalid motor id %d, must be 1-4", motorID)
	}
	motors := uc.registry.Motors()
	if motors[idx] == nil {
		return fmt.Errorf("motor %d offline", motorID)
	}
	return motors[idx].Move(ctx, steps, speedHz)
}
